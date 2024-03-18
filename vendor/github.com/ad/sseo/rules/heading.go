package rules

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// HeadingRule is a rule that checks if the title is present and not too long.
type HeadingRule struct {
	Headings []Heading
	HTML     string

	Errors []string
}

type Heading struct {
	Level    int
	Position int
	Text     string
}

func WithHeading(parsed *goquery.Document) Rule {
	headings := []Heading{}

	html, _ := parsed.Html()

	for i := 1; i <= 6; i++ {
		parsed.Find(fmt.Sprintf("h%d", i)).Each(func(index int, item *goquery.Selection) {
			heading := Heading{
				Level:    i,
				Position: item.Index(),
				Text:     item.Text(),
			}

			headings = append(headings, heading)
		})
	}

	return &HeadingRule{Headings: headings, HTML: html}
}

// Check checks if the title is present and not too long.
func (r *HeadingRule) Check() error {
	if len(r.Headings) == 0 {
		r.Errors = append(r.Errors, "headings are missing")

		return fmt.Errorf("%s", strings.Join(r.Errors, ", "))
	}

	isPresentH1 := false
	h1Count := 0

	for _, heading := range r.Headings {
		if heading.Level == 1 {
			isPresentH1 = true
			h1Count++
		}
	}

	if !isPresentH1 {
		r.Errors = append(r.Errors, "h1 is missing")
	}

	if h1Count > 1 {
		r.Errors = append(r.Errors, "multiple h1 are present")
	}

	if isPresentH1 {
		// check heading order in html string
		for i := 1; i <= 6; i++ {
			prev := strings.Index(r.HTML, fmt.Sprintf("<h%d", i))
			current := strings.Index(r.HTML, fmt.Sprintf("<h%d", i+1))
			if current > -1 && prev > current {
				r.Errors = append(r.Errors, fmt.Sprintf("h%d is not in correct order (%d > %d)", i, prev, current))
			}
		}
	}

	if len(r.Errors) > 0 {
		return fmt.Errorf("%s", strings.Join(r.Errors, ", "))
	}

	return nil
}
