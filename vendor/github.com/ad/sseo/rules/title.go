package rules

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// TitleRule is a rule that checks if the title is present and not too long.
type TitleRule struct {
	// Title is the title of the page.
	Title string

	Errors []string
}

func WithTitle(parsed *goquery.Document) Rule {
	titleText := ""
	title := parsed.Find("head title")

	if title.Text() != "" {
		titleText = strings.Trim(title.Text(), "\t\r\n ")
	}

	return &TitleRule{Title: titleText}
}

// Check checks if the title is present and not too long.
func (r *TitleRule) Check() error {
	if r.Title == "" {
		r.Errors = append(r.Errors, "title is missing")

		return fmt.Errorf("%s", strings.Join(r.Errors, ", "))
	}

	titleLength := len([]rune(r.Title))

	if titleLength > 80 {
		r.Errors = append(r.Errors, fmt.Sprintf("title is too long (%d > 80)", titleLength))
	}

	if titleLength < 30 {
		r.Errors = append(r.Errors, fmt.Sprintf("title is too short (%d < 30)", titleLength))
	}

	if len(r.Errors) > 0 {
		return fmt.Errorf("%s", strings.Join(r.Errors, ", "))
	}

	return nil
}
