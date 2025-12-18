package engine

import (
	"context"
	"embed"
	"errors"
	"fmt"
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
	ClipboardTexts []string `json:"clipboard_texts"`
	Cookies        []Cookie `json:"cookies"`
	CookiePairs    []string `json:"cookie_pairs"`
	Forms          []Form   `json:"forms"`
	Head           Head     `json:"head"`
	InitialTitle   string   `json:"initial_title"`
	Links          []Link   `json:"links"`
	VisibleText    string   `json:"visible_text"`
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
}

type Form struct {
	Action string  `json:"action"`
	Method string  `json:"method"`
	Class  string  `json:"class"`
	ID     string  `json:"id"`
	Inputs []Input `json:"fields"`
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
	result.Head = Head{}

	wait := int64(task.IntParam("wait", 50))

	if err = extractVisibleText(ctx, &result); err != nil {
		logger.Warn().Msgf("visible text error: %v", err)
	}

	if err = runFormAnalysis(ctx, wait, &result, logger); err != nil {
		logger.Warn().Msgf("form analysis error: %v", err)
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

	if err = runInteractions(ctx, wait, &result, logger); err != nil {
		logger.Warn().Msgf("interaction error: %v", err)
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

func runFormAnalysis(ctx context.Context, wait int64, result *AnalyzeResult, logger *zerolog.Logger) error {
	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var formNodes []*cdp.Node
		var labelNodes []*cdp.Node

		if err := queryWithDeadline(ctx, wait, func(ctx context.Context) error {
			return chromedp.Nodes("form", &formNodes, chromedp.ByQueryAll).Do(ctx)
		}); err != nil {
			return err
		}

		if err := queryWithDeadline(ctx, wait, func(ctx context.Context) error {
			return chromedp.Nodes("label", &labelNodes, chromedp.ByQueryAll).Do(ctx)
		}); err != nil {
			return err
		}

		// Map label nodes by ID
		labelsByID := make(map[string]*cdp.Node, len(labelNodes))
		for _, label := range labelNodes {
			labelsByID[label.AttributeValue("for")] = label
		}

		result.Forms = make([]Form, 0, len(formNodes))

		formNodeAttrs := attributesFromNodes(ctx, formNodes, []string{"action", "method", "class", "id"})
		for idx, formNode := range formNodes {
			var inputNodes []*cdp.Node

			if err := queryWithDeadline(ctx, wait, func(ctx context.Context) error {
				return chromedp.Nodes("input, textarea, select", &inputNodes, chromedp.ByQueryAll, chromedp.FromNode(formNode)).Do(ctx)
			}); err != nil {
				return err
			}

			formAttrs := formNodeAttrs[idx]
			form := Form{
				Action: formAttrs[0],
				Method: strings.ToUpper(formAttrs[1]),
				Class:  formAttrs[2],
				ID:     formAttrs[3],
				Inputs: make([]Input, len(inputNodes)),
			}

			inputNodeAttrs := attributesFromNodes(ctx, inputNodes, []string{"class", "id", "placeholder", "name", "type", "value"})
			for jdx, inputAttrs := range inputNodeAttrs {
				id := inputAttrs[1]

				form.Inputs[jdx].Class = inputAttrs[0]
				form.Inputs[jdx].ID = id
				form.Inputs[jdx].Placeholder = strings.TrimSpace(inputAttrs[2])
				form.Inputs[jdx].Name = inputAttrs[3]
				form.Inputs[jdx].Type = inputAttrs[4]
				form.Inputs[jdx].Value = inputAttrs[5]

				if inputAttrs[4] != "hidden" {
					form.Inputs[jdx].Label = findLabelTextForInput(ctx, labelsByID[id], inputNodes[jdx], logger)
				}
			}

			result.Forms = append(result.Forms, form)
		}

		return nil
	}))
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

func runInteractions(ctx context.Context, wait int64, result *AnalyzeResult, logger *zerolog.Logger) error {
	err := runClipboardInteractions(ctx, wait, result, logger)
	if err != nil {
		return err
	}

	return nil
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

// Try to find label text. First, we'll look for a label node whose
// `for` attribute is equal to the `id` attribute of the input. Next,
// we'll check if the input has a label parent. If so, we'll take the
// text content of that parent.
func findLabelTextForInput(ctx context.Context, labelNode *cdp.Node, inputNode *cdp.Node, logger *zerolog.Logger) string {
	inputLog := logger.With().Str("input", inputNode.FullXPath()).Logger()

	if labelNode == nil {
		labelNode = findParentByTagName("LABEL", inputNode)
	}

	if labelNode == nil {
		inputLog.Debug().Msg("no label node found")

		return ""
	}

	var label string
	if err := chromedp.Run(ctx, chromedp.TextContent(labelNode.FullXPath(), &label, chromedp.BySearch)); err != nil {
		logger.Warn().Msgf("form label error: %v", err)
	}

	return strings.TrimSpace(label)
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
