package engine

import (
	"context"
	"embed"
	"errors"
	"time"

	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/rs/zerolog"
)

//go:embed js/*.js
var jsFS embed.FS

type AnalyzeResult struct {
	ClipboardTexts []string `json:"clipboard_texts"`
	Forms          []Form   `json:"forms"`
	Head           Head     `json:"head"`
	Links          []Link   `json:"links"`
	VisibleText    string   `json:"visible_text"`
	*Visit
}

type Input struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
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
	result.Head = Head{}

	wait := int64(task.IntParam("wait", 50))

	if err = extractVisibleText(ctx, &result); err != nil {
		logger.Warn().Msgf("visible text error: %v", err)
	}

	if err = runFormAnalysis(ctx, wait, &result); err != nil {
		logger.Warn().Msgf("form analysis error: %v", err)
	}

	if err = runLinkAnalysis(ctx, wait, &result); err != nil {
		logger.Warn().Msgf("href analysis error: %v", err)
	}

	if err = runHeadAnalysis(ctx, wait, &result); err != nil {
		logger.Warn().Msgf("head analysis error: %v", err)
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

func runFormAnalysis(ctx context.Context, wait int64, result *AnalyzeResult) error {
	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var formNodes []*cdp.Node

		if err := queryWithDeadline(ctx, wait, func(ctx context.Context) error {
			return chromedp.Nodes("form", &formNodes, chromedp.ByQueryAll).Do(ctx)
		}); err != nil {
			return err
		}

		result.Forms = make([]Form, 0, len(formNodes))

		formNodeAttrs := attributesFromNodes(ctx, formNodes, []string{"action", "method", "class", "id"})
		for idx, formNode := range formNodes {
			var inputNodes []*cdp.Node
			if err := queryWithDeadline(ctx, wait, func(ctx context.Context) error {
				return chromedp.Nodes("input, textarea, select", &inputNodes, chromedp.FromNode(formNode)).Do(ctx)
			}); err != nil {
				return err
			}

			formAttrs := formNodeAttrs[idx]
			form := Form{
				Action: formAttrs[0],
				Method: formAttrs[1],
				Class:  formAttrs[2],
				ID:     formAttrs[3],
				Inputs: make([]Input, len(inputNodes)),
			}

			inputNodeAttrs := attributesFromNodes(ctx, inputNodes, []string{"name", "type", "value"})
			for jdx, inputAttrs := range inputNodeAttrs {
				form.Inputs[jdx].Name = inputAttrs[0]
				form.Inputs[jdx].Type = inputAttrs[1]
				form.Inputs[jdx].Value = inputAttrs[2]
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

		err := queryWithDeadline(ctx, 20, func(ctx context.Context) error {
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
