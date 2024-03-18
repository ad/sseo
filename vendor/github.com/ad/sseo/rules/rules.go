package rules

import (
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

type Rule interface {
	Check() error
}

// Rules is a set of rules.
type Rules struct {
	Parsed     *goquery.Document
	StatusCode int
	URL        string
	rules      []Rule
}

// Check checks all the rules.
func (r *Rules) Check() []string {
	errors := []string{}

	for _, rule := range r.rules {
		if err := rule.Check(); err != nil {
			errors = append(errors, err.Error())
		}
	}

	return errors
}

func NewRulesWith(url string, opts ...func(*Rules)) (*Rules, error) {
	r := &Rules{
		URL: url,
	}

	// Request the HTML page.
	res, err := http.Get(url)
	if err != nil {
		return r, err
	}
	defer res.Body.Close()

	r.StatusCode = res.StatusCode

	parsed := &goquery.Document{
		Url:       res.Request.URL,
		Selection: &goquery.Selection{},
	}

	if res.StatusCode == 200 {
		// Load the HTML document
		parsed, err = goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			return r, err
		}
	}

	r.Parsed = parsed

	for _, opt := range opts {
		opt(r)
	}

	return r, nil
}

func (r *Rules) AddRule(rule Rule) {
	r.rules = append(r.rules, rule)
}
