package rules

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// SitemapRule is the sitemap.xml of the page.
type SitemapRule struct {
	Body       []byte
	StatusCode int

	Errors []string
}

func WithSitemap(path string) Rule {
	rule := &SitemapRule{}

	u, err := url.Parse(path)
	if err != nil {
		return rule
	}

	path = fmt.Sprintf("%s://%s/sitemap.xml", u.Scheme, u.Host)

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
func (r *SitemapRule) Check() error {
	if string(r.Body) == "" || r.StatusCode != 200 {
		r.Errors = append(r.Errors, "sitemap.xml is missing")
	}

	if len(r.Errors) > 0 {
		return fmt.Errorf("%s", strings.Join(r.Errors, ", "))
	}

	return nil
}
