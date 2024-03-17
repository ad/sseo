package rules

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/temoto/robotstxt"
)

type RobotsTXTRule struct {
	Body       []byte
	StatusCode int

	Errors []string
}

func WithRobotsTXT(path string) Rule {
	rule := &RobotsTXTRule{}

	u, err := url.Parse(path)
	if err != nil {
		return rule
	}

	path = fmt.Sprintf("%s://%s/robots.txt", u.Scheme, u.Host)

	// Request the HTML page.
	res, err := http.Get(path)
	if err != nil {

		return rule
	}

	defer res.Body.Close()

	rule.StatusCode = res.StatusCode
	rule.Body, err = io.ReadAll(res.Body)
	if err != nil {
		return rule
	}

	return rule
}

// Check checks if the robots.txt is present and not too long.
func (r *RobotsTXTRule) Check() error {
	robots, err := robotstxt.FromStatusAndBytes(r.StatusCode, r.Body)
	if err != nil {
		return err
	}

	if robots == nil {
		r.Errors = append(r.Errors, "robots.txt is missing")
	}

	if !robots.TestAgent("/", "Googlebot") {
		r.Errors = append(r.Errors, "Google bot is not allowed")
	}

	if !robots.TestAgent("/", "Yandex") {
		r.Errors = append(r.Errors, "Yandex bot is not allowed")
	}

	if !robots.TestAgent("/", "*") {
		r.Errors = append(r.Errors, "site is closed for all")
	}

	if len(r.Errors) > 0 {
		return fmt.Errorf("%s", strings.Join(r.Errors, ", "))
	}

	return nil
}
