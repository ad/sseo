package rules

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// DescriptionRule is a rule that checks if the description is present and not too long.
type DescriptionRule struct {
	// Description is the description of the page.
	Description string

	Errors []string
}

func WithDescription(parsed *goquery.Document) Rule {
	descriptionText := ""
	description := parsed.Find("meta[name='description']")

	if descriptionContent, ok := description.Attr("content"); ok {
		descriptionText = strings.Trim(descriptionContent, "\t\r\n ")
	}

	return &DescriptionRule{Description: descriptionText}
}

// Check checks if the description is present and not too long.
func (r *DescriptionRule) Check() error {
	if r.Description == "" {
		r.Errors = append(r.Errors, "description is missing")

		return fmt.Errorf("%s", strings.Join(r.Errors, ", "))
	}

	descriptionLength := len([]rune(r.Description))

	if descriptionLength > 300 {
		r.Errors = append(r.Errors, fmt.Sprintf("description is too long (%d > 300)", descriptionLength))
	}

	if descriptionLength < 70 {
		r.Errors = append(r.Errors, fmt.Sprintf("description is too short (%d < 70)", descriptionLength))
	}

	if len(r.Errors) > 0 {
		return fmt.Errorf("%s", strings.Join(r.Errors, ", "))
	}

	return nil
}
