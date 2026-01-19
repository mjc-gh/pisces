package engine

import (
	"context"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/brianvoe/gofakeit/v7/source"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/rs/zerolog"
)

var defaultFormData formData = &fakeItData{
	gofakeit.NewFaker(source.NewCrypto(), true),
}

type InputValueType int

const (
	InputUnknown InputValueType = iota
	InputEmail
	InputPhone
	InputPassword
	InputName
	InputFirstName
	InputMiddleName
	InputLastName
)

const INPUT_HIDDEN = "hidden"

type formData interface {
	Email() string
	Phone() string
	Password() string
	Name() string
	FirstName() string
	MiddleName() string
	LastName() string
}

// scanForForms will look for form elements and add them to the forms slice
// pointer.
func scanForForms(ctx context.Context, wait int64, forms *[]Form, logger *zerolog.Logger) error {
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

		formNodeAttrs := attributesFromNodes(ctx, formNodes, []string{"action", "method", "class", "id"})
		for idx, formNode := range formNodes {
			var inputNodes []*cdp.Node

			if err := queryWithDeadline(ctx, wait, func(ctx context.Context) error {
				return chromedp.Nodes("input, textarea, select", &inputNodes, chromedp.ByQueryAll, chromedp.FromNode(formNode)).Do(ctx)
			}); err != nil {
				return err
			}

			formAttrs := formNodeAttrs[idx]
			method := strings.ToUpper(formAttrs[1])
			form := Form{
				Action: formAttrs[0],
				Method: method,
				Class:  formAttrs[2],
				ID:     formAttrs[3],
				Inputs: make([]Input, len(inputNodes)),
				xpath:  formNode.FullXPath(),
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
				form.Inputs[jdx].xpath = inputNodes[jdx].FullXPath()

				if inputAttrs[4] != INPUT_HIDDEN {
					form.Inputs[jdx].Label = findLabelTextForInput(ctx, labelsByID[id], inputNodes[jdx], logger)
				}
			}

			*forms = append(*forms, form)
		}

		return nil
	}))
}

// Complete form fields for emails, passwords, names, etc. This function will
// also return the last input that is completed (which is likely then used to
// submit the form).
func completeForm(ctx context.Context, form *Form, logger *zerolog.Logger) string {
	var lastInputXPath string

	for _, input := range form.Inputs {
		var value string

		switch input.Type {
		case "email":
			value = defaultFormData.Email()
		case "password":
			value = defaultFormData.Password()
		case "text":
			switch detectFormInputType(&input) {
			case InputEmail:
				value = defaultFormData.Email()
			case InputPassword:
				value = defaultFormData.Password()
			case InputPhone:
				value = defaultFormData.Phone()
			case InputName:
				value = defaultFormData.Name()
			case InputFirstName:
				value = defaultFormData.FirstName()
			case InputMiddleName:
				value = defaultFormData.MiddleName()
			case InputLastName:
				value = defaultFormData.LastName()
			case InputUnknown:
				logger.Debug().Msgf("unknown text input value type %s", input)

				continue
			}
		}

		// Unknown input type
		if value == "" {
			continue
		}

		if err := completeFormInput(ctx, input.xpath, value); err != nil {
			logger.Warn().Msgf("complete form input error: %v", err)
		}

		lastInputXPath = input.xpath
	}

	return lastInputXPath
}

func analyzeForm(ctx context.Context, form *Form, visit *Visit, logger *zerolog.Logger) (bool, error) {
	visit.RequestedURL = form.Action

	inputXPath := completeForm(ctx, form, logger)
	if inputXPath == "" {
		logger.Info().Msgf("form %s has no inputs to interact with", form)

		if len(form.Inputs) > 0 && !form.AllInputsHidden() {
			inputXPath = form.Inputs[0].xpath
		}
	}

	if inputXPath == "" {
		return false, nil
	}

	if err := chromedp.Run(
		ctx,
		chromedp.Submit(inputXPath, chromedp.BySearch),
		chromedp.Sleep(2*time.Second),
		// chromedp.WaitReady("body"),
		chromedp.Location(&visit.Location),
		chromedp.Sleep(2*time.Second),
		chromedp.OuterHTML("html", &visit.Body),
	); err != nil {
		return false, err
	}

	return true, nil
}

func completeFormInput(ctx context.Context, xpath, value string) error {
	if err := chromedp.Run(
		ctx,
		chromedp.Focus(xpath, chromedp.BySearch),
		chromedp.SendKeys(xpath, value, chromedp.BySearch),
	); err != nil {
		return err
	}

	return nil
}

// Look at labels and other attributes for context clues on what the
// input is for.
func detectFormInputType(input *Input) InputValueType {
	strs := []string{
		strings.ToLower(input.Label),
		strings.ToLower(input.Placeholder),
		strings.ToLower(input.ID),
		strings.ToLower(input.Class),
	}

	for _, str := range strs {
		switch {
		case strings.Contains(str, "email"):
			return InputEmail
		case strings.Contains(str, "password"):
			return InputPassword
		case strings.Contains(str, "phone"):
			return InputPhone
		case strings.Contains(str, "first") && strings.Contains(str, "name"):
			return InputFirstName
		case strings.Contains(str, "middle") && strings.Contains(str, "name"):
			return InputMiddleName
		case strings.Contains(str, "last") && strings.Contains(str, "name"):
			return InputLastName
		case strings.Contains(str, "name"):
			return InputName
		}
	}

	return InputUnknown
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

// Score is used to rank forms during form submission interaction. This
// interaction mode has a maximum submission limit, and if there are more forms
// than the limit allows, the highest-ranked forms are submitted first.
//
// Scoring criteria:
//   - Password fields: +50 points each (strongly indicates login/signup form)
//   - Email inputs: +20 points each
//   - Text inputs: +10 points each (up to 3, then -5 for each additional)
//   - Other non-hidden inputs: +5 points each
//   - Bonus: +15 points if form has 2-4 non-hidden inputs (typical login form)
func (f *Form) Score() int {
	score := 0
	nonHiddenCount := 0
	textInputCount := 0

	for _, input := range f.Inputs {
		switch input.Type {
		case INPUT_HIDDEN:
			// Hidden inputs don't contribute to score
			continue
		case "password":
			score += 50
		case "email":
			score += 20
		case "text":
			textInputCount++
			if textInputCount <= 3 {
				score += 10
			} else {
				// Penalize forms with many text inputs (likely not login forms)
				score -= 5
			}
		default:
			// Other non-hidden types (checkbox, radio, submit, etc.)
			score += 5
		}
		nonHiddenCount++
	}

	// Bonus for forms with 2-4 non-hidden inputs (typical login/signup forms)
	if nonHiddenCount >= 2 && nonHiddenCount <= 4 {
		score += 15
	}

	return score
}

func (f *Form) AllInputsHidden() bool {
	for _, input := range f.Inputs {
		if input.Type != INPUT_HIDDEN {
			return false
		}
	}

	return true
}

func (f *Form) String() string {
	if f.ID != "" {
		return f.ID
	} else if f.Class != "" {
		return f.Class
	}

	return f.xpath
}

func (i *Input) String() string {
	if i.ID != "" {
		return i.ID
	} else if i.Class != "" {
		return i.Class
	}

	return i.xpath
}

type fakeItData struct {
	source *gofakeit.Faker
}

func (f *fakeItData) Name() string       { return f.source.Name() }
func (f *fakeItData) FirstName() string  { return f.source.FirstName() }
func (f *fakeItData) MiddleName() string { return f.source.MiddleName() }
func (f *fakeItData) LastName() string   { return f.source.LastName() }
func (f *fakeItData) Phone() string      { return f.source.Phone() }
func (f *fakeItData) Email() string      { return f.source.Email() }

func (f *fakeItData) Password() string {
	return f.source.Password(true, false, false, false, false, 8)
}
