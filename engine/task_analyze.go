package engine

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/cdproto/storage"
	"github.com/chromedp/chromedp"
	"github.com/rs/zerolog"
	"golang.org/x/net/html"
)

//go:embed js/*.js
var jsFS embed.FS

type AnalyzeResult struct {
	ClipboardTexts  []string         `json:"clipboard_texts"`
	Cookies         []Cookie         `json:"cookies"`
	CookiePairs     []string         `json:"cookie_pairs"`
	Forms           []Form           `json:"forms"`
	FormSubmissions []FormSubmission `json:"form_submissions"`
	Head            Head             `json:"head"`
	InitialTitle    string           `json:"initial_title"`
	Links           []Link           `json:"links"`
	VisibleText     string           `json:"visible_text"`
	*Visit
}

type Cookie struct {
	Name      string    `json:"name"`
	Value     string    `json:"value"`
	Domain    string    `json:"domain"`
	Path      string    `json:"path"`
	ExpiresAt time.Time `json:"expires_at,omitzero"`
	Expires   float64   `json:"expires"`
	HTTPOnly  bool      `json:"http_only"`
	Secure    bool      `json:"secure"`
	Session   bool      `json:"session"`
	SameSite  string    `json:"same_site,omitempty"`
}

type Input struct {
	Class       string `json:"class,omitempty"`
	ID          string `json:"id,omitempty"`
	Label       string `json:"label,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	Name        string `json:"name,omitempty"`
	Type        string `json:"type,omitempty"`
	Value       string `json:"value,omitempty"`
	xpath       string
}

type Form struct {
	Action string  `json:"action"`
	Method string  `json:"method"`
	Class  string  `json:"class"`
	ID     string  `json:"id"`
	Inputs []Input `json:"fields"`
	xpath  string
}

type FormSubmission struct {
	Method string `json:"method"`
	*Visit
}

type Link struct {
	Href  string `json:"href"`
	Text  string `json:"text,omitempty"`
	Class string `json:"class,omitempty"`
}

type Head struct {
	Title           string `json:"title"`
	Description     string `json:"description"`
	FaviconUrl      string `json:"favicon_url"`
	ShortcutIconUrl string `json:"shortcut_icon_url"`
	Viewport        string `json:"viewport"`
}

func performAnalyzeTask(ctx context.Context, task *Task, logger *zerolog.Logger) (AnalyzeResult, error) {
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	crawler := NewCrawler(task.userAgent, int64(task.winWidth), int64(task.winHeight))
	crawler.SetupListeners(ctx, logger)

	err := crawler.Visit(ctx, task.url, logger)
	if err != nil {
		return AnalyzeResult{}, err
	}

	visit := crawler.LastVisit()
	if visit == nil {
		return AnalyzeResult{}, ErrNoCrawlerVisit
	}

	result := AnalyzeResult{Visit: visit}
	result.InitialTitle = titleFromHTML(result.InitialBody, 100)

	result.ClipboardTexts = []string{}
	result.Cookies = []Cookie{}
	result.CookiePairs = []string{}
	result.Forms = []Form{}
	result.Head = Head{}
	result.Links = []Link{}

	wait := int64(task.IntParam("wait", 50))
	maxFormSubmits := task.IntParam("max-form-submits", 1)
	doClipboardInteraction := task.BoolParam("clipboard", true)
	doFormInteraction := task.BoolParam("forms", true)

	if err = extractVisibleText(ctx, &result); err != nil {
		logger.Warn().Msgf("visible text error: %v", err)
	}

	if err = runLinkAnalysis(ctx, wait, &result); err != nil {
		logger.Warn().Msgf("href analysis error: %v", err)
	}

	if err = runHeadAnalysis(ctx, wait, &result); err != nil {
		logger.Warn().Msgf("head analysis error: %v", err)
	}

	if err = runCookieAnalysis(ctx, &result); err != nil {
		logger.Warn().Msgf("cookie analysis error: %v", err)
	}

	if err = runInitialFormScan(ctx, wait, &result, logger); err != nil {
		logger.Warn().Msgf("initial form scan error: %v", err)
	}

	if doClipboardInteraction {
		if err = runClipboardInteractions(ctx, wait, &result, logger); err != nil {
			logger.Warn().Msgf("clipboard interaction error: %v", err)
		}
	}

	if doFormInteraction {
		if err := runFormInteractions(ctx, maxFormSubmits, &crawler, &result, logger); err != nil {
			logger.Warn().Msgf("form interaction error: %v", err)
		}
	}

	return result, nil
}

func extractVisibleText(ctx context.Context, result *AnalyzeResult) error {
	visibleText, err := evaluateAsString(ctx, "js/visible_text.js")
	if err != nil {
		return err
	}

	result.VisibleText = visibleText

	return nil
}

func runLinkAnalysis(ctx context.Context, wait int64, result *AnalyzeResult) error {
	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var anchorNodes []*cdp.Node

		if err := queryWithDeadline(ctx, wait, func(ctx context.Context) error {
			return chromedp.Nodes("a[href]", &anchorNodes, chromedp.ByQueryAll).Do(ctx)
		}); err != nil {
			return err
		}

		result.Links = make([]Link, 0, len(anchorNodes))

		anchorNodeAttrs := attributesFromNodes(ctx, anchorNodes, []string{"href", "class"})
		for idx, anchorNode := range anchorNodes {
			anchorAttrs := anchorNodeAttrs[idx]
			if anchorAttrs[0] == "" {
				continue
			}

			link := Link{Href: anchorAttrs[0], Class: anchorAttrs[1]}

			err := chromedp.Run(ctx, chromedp.TextContent(anchorNode.FullXPath(), &link.Text, chromedp.BySearch))
			if err != nil {
				return err
			}

			result.Links = append(result.Links, link)
		}

		return nil
	}))
}
func runHeadAnalysis(ctx context.Context, wait int64, result *AnalyzeResult) error {
	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var children []*cdp.Node

		err := queryWithDeadline(ctx, wait, func(ctx context.Context) error {
			return chromedp.Nodes("head > *", &children, chromedp.ByQueryAll).Do(ctx)
		})
		if err != nil {
			return err
		}

		for _, child := range children {
			switch child.NodeName {
			case "LINK":
				var rel, href string

				err := chromedp.Run(
					ctx,
					chromedp.JavascriptAttribute(child.FullXPath(), "rel", &rel, chromedp.BySearch),
					chromedp.JavascriptAttribute(child.FullXPath(), "href", &href, chromedp.BySearch),
				)
				if err != nil {
					return err
				}

				switch rel {
				case "icon":
					result.Head.FaviconUrl = href
				case "shortcut icon", "icon shortcut":
					result.Head.ShortcutIconUrl = href
				}
			case "META":
				var name, content string

				err := chromedp.Run(
					ctx,
					chromedp.JavascriptAttribute(child.FullXPath(), "name", &name, chromedp.BySearch),
					chromedp.JavascriptAttribute(child.FullXPath(), "content", &content, chromedp.BySearch),
				)
				if err != nil {
					return err
				}

				switch name {
				case "description":
					result.Head.Description = content
				case "viewport":
					result.Head.Viewport = content
				}
			case "TITLE":
				err := chromedp.Run(ctx, chromedp.TextContent(child.FullXPath(), &result.Head.Title, chromedp.BySearch))
				if err != nil {
					return err
				}
			}
		}

		return nil
	}))
}

func runCookieAnalysis(ctx context.Context, result *AnalyzeResult) error {
	var cookies []*network.Cookie

	if err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error

			if cookies, err = storage.GetCookies().Do(ctx); err != nil {
				return err
			}

			return nil
		}),
	); err != nil {
		return err
	}

	// Print the cookies
	result.Cookies = make([]Cookie, len(cookies))
	result.CookiePairs = make([]string, len(cookies))

	for idx, cookie := range cookies {
		result.Cookies[idx] = Cookie{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Domain:   cookie.Domain,
			Path:     cookie.Path,
			HTTPOnly: cookie.HTTPOnly,
			Secure:   cookie.Secure,
			Session:  cookie.Session,
			SameSite: string(cookie.SameSite),
		}

		if cookie.Expires != -1 {
			result.Cookies[idx].Expires = cookie.Expires
			result.Cookies[idx].ExpiresAt = time.Unix(int64(cookie.Expires), 0)
		}

		result.CookiePairs[idx] = fmt.Sprintf("%s=%s", cookie.Name, cookie.Value)
	}

	return nil
}

func runInitialFormScan(ctx context.Context, wait int64, result *AnalyzeResult, logger *zerolog.Logger) error {
	return scanForForms(ctx, wait, &result.Forms, logger)
}

func runClipboardInteractions(ctx context.Context, wait int64, result *AnalyzeResult, logger *zerolog.Logger) error {
	clipboardJS, err := jsFS.ReadFile("js/clipboard.js")
	if err != nil {
		return err
	}

	clearJS, err := jsFS.ReadFile("js/clipboard_clear.js")
	if err != nil {
		return err
	}

	if err := setupNavigationLock(ctx, logger); err != nil {
		return err
	}

	// Grant permissions to access the clipboard and reset clipboard content
	if err := chromedp.Run(
		ctx,
		browser.
			GrantPermissions([]browser.PermissionType{browser.PermissionTypeClipboardReadWrite}).
			WithOrigin(result.Location),
		chromedp.Evaluate(string(clearJS), nil, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
			return p.WithAwaitPromise(true)
		}),
	); err != nil {
		return err
	}

	// Enumerate all nodes and click them; see if anything was added to the clipboard
	var allNodes []*cdp.Node

	// var allNodesSelector = "body *:not(script, style, a[href], a[href] *, button *)"
	// Find all leaf nodes that aren't part of links
	var allNodesSelector = "body *:not(:has(*)):not(script, style, a[href], a[href] *, option, svg *)"

	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		err := queryWithDeadline(ctx, wait, func(ctx context.Context) error {
			return chromedp.Nodes(
				allNodesSelector,
				&allNodes,
				chromedp.ByQueryAll,
			).Do(ctx)
		})
		if err != nil {
			return err
		}

		return nil
	})); err != nil {
		return err
	}

	logger.Debug().Msgf("found %d nodes to click for clipboard", len(allNodes))

	cc := NewClipboardCapture()
	for _, node := range allNodes {
		var clipboardText string

		err := queryWithDeadline(ctx, wait, func(ctx context.Context) error {
			return chromedp.Run(
				ctx,
				chromedp.Click(node.FullXPath(), chromedp.BySearch),
				chromedp.Evaluate(string(clipboardJS), &clipboardText, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
					return p.WithAwaitPromise(true)
				}),
			)
		})
		if err != nil {
			return err
		}

		if clipboardText != "" {
			cc.AddTo(clipboardText)
		}
	}

	if _, err := evaluateAsString(ctx, "js/navigation_unlock.js"); err != nil {
		return err
	}

	result.ClipboardTexts = cc.Values()

	return nil
}

// TODO:
// - using the initially scanned forms, prioritize which forms to submit first.
//   - add a task parameter for max form submissions and default to 1.
//
// - submit each submittable form and return navigation to initial page.
func runFormInteractions(ctx context.Context, maxSubmits int, crawler *Crawler, result *AnalyzeResult, logger *zerolog.Logger) error {
	// Create a slice of form pointers and sort them by priority. We'll only
	// perform maxSubmits and will prioritize forms using a custom ranking based
	// upon the fields defined.
	// knownForms := make([]*Form, len(result.Forms))
	knownForms := []*Form{}
	for _, f := range result.Forms {
		knownForms = append(knownForms, &f)
	}

	slices.SortFunc(knownForms, func(aform, bform *Form) int {
		a := aform.Score()
		b := bform.Score()

		if b < a {
			return -1
		} else if b > a {
			return 1
		}

		return 0
	})

	if len(knownForms) == 0 {
		return nil
	}

	result.FormSubmissions = []FormSubmission{}

	for index, nextForm := range knownForms {
		// When we're submitting against multiple forms, we have to return the
		// browser back to the original Location and ensure the form is still
		// present.
		if index > 0 {
			if err := chromedp.Run(ctx, chromedp.Navigate(result.Location)); err != nil {
				return err
			}

			exists, err := checkElementExists(ctx, nextForm.xpath)
			if err != nil || !exists {
				logger.Warn().Msgf("form at index %d not found on return visit", index)

				continue
			}
		}

		// Find the first not known form
		if nextForm == nil {
			logger.Debug().Msgf("found %d forms total", len(knownForms))

			break
		}

		submitted := false
		submission := FormSubmission{}
		if err := crawler.captureVisit(ctx, func(ctx context.Context) error {
			var err error

			visit := crawler.currentVisit

			if submitted, err = analyzeForm(ctx, nextForm, visit, logger); err != nil {
				return err
			}

			return nil
		}); err != nil {
			return err
		}

		if !submitted {
			continue
		}

		submission.Method = nextForm.Method
		submission.Visit = crawler.LastVisit()

		result.FormSubmissions = append(result.FormSubmissions, submission)

		// We have hit our max form submission limit
		if index+1 >= maxSubmits {
			break
		}
	}

	return nil
}

func setupNavigationLock(ctx context.Context, logger *zerolog.Logger) error {
	if _, err := evaluateAsString(ctx, "js/navigation_lock.js"); err != nil {
		return err
	}

	// Listen for JavaScript onbeforeunload dialog and cancel navigation
	chromedp.ListenTarget(ctx, func(ev any) {
		if ev, ok := ev.(*page.EventJavascriptDialogOpening); ok {
			l := logger.With().Str("type", string(ev.Type)).Logger()
			l.Debug().Msgf("dialog opening")

			if ev.Type != page.DialogTypeBeforeunload {
				l.Warn().Msgf("unexpected dialog blocking evaluation: %s", ev.Message)

				return
			}

			go func() {
				err := chromedp.Run(ctx, page.HandleJavaScriptDialog(false))
				if err != nil {
					l.Debug().Msgf("navigation lock dialog error: %v", err)
				} else {
					l.Debug().Msg("navigation lock dialog dismissed")
				}
			}()
		}
	})

	return nil
}

// Evaluate JS snippets from the jsFS as a string.
func evaluateAsString(ctx context.Context, jsPath string) (string, error) {
	var result string

	js, err := jsFS.ReadFile(jsPath)
	if err != nil {
		return "", err
	}

	if err = chromedp.Run(ctx, chromedp.Evaluate(string(js), &result)); err != nil {
		return "", err
	}

	return result, nil
}

func queryWithDeadline(ctx context.Context, wait int64, callback func(context.Context) error) error {
	queryCtx, queryCancel := context.WithTimeout(ctx, time.Duration(wait)*time.Millisecond)
	defer queryCancel()

	err := callback(queryCtx)
	if errors.Is(err, context.DeadlineExceeded) {
		return nil
	}

	return err
}

func attributesFromNodes(ctx context.Context, nodes []*cdp.Node, attributes []string) [][]string {
	values := make([][]string, len(nodes))

	for i, node := range nodes {
		values[i] = make([]string, len(attributes))

		for j, attribute := range attributes {
			err := chromedp.Run(ctx,
				chromedp.JavascriptAttribute(
					node.FullXPath(), attribute, &values[i][j], chromedp.BySearch,
				),
			)

			// Fallback to Attribute method on the Node type
			if err != nil {
				if val, ok := node.Attribute(attribute); ok {
					values[i][j] = val
				}
			}
		}
	}

	return values
}

// checkElementExists is a custom action to check for the presence of an element via XPath.
func checkElementExists(ctx context.Context, xpath string) (bool, error) {
	var nodes []*cdp.Node

	err := chromedp.Run(ctx, chromedp.Nodes(xpath, &nodes, chromedp.BySearch, chromedp.AtLeast(0)))
	if err != nil {
		return false, err
	}

	if len(nodes) > 0 {
		return true, nil
	}

	return false, nil
}

func findParentByTagName(name string, node *cdp.Node) *cdp.Node {
	for {
		parent := node.Parent
		if parent == nil {
			return nil
		}

		if parent.NodeName == name {
			return parent
		}

		node = parent
	}
}

func titleFromHTML(htmlStr string, maxNodes int) string {
	if maxNodes <= 0 {
		maxNodes = 100
	}

	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return ""
	}

	var (
		nodeCount int
		traverse  func(*html.Node) string
	)

	traverse = func(n *html.Node) string {
		if nodeCount >= maxNodes {
			return ""
		}
		nodeCount++

		if n.Type == html.ElementNode && n.Data == "title" {
			for p := n.Parent; p != nil; p = p.Parent {
				if p.Type == html.ElementNode && p.Data == "head" {
					if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
						return n.FirstChild.Data
					}

					return ""
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if title := traverse(c); title != "" {
				return title
			}
		}

		return ""
	}

	return traverse(doc)
}
