package rules

import (
	"fmt"
	"strings"
)

// StatusRule is a rule that checks if the status is allowed.
type StatusRule struct {
	// Status is the Status of the page.
	Status int

	Allowed []int

	Errors []string
}

func WithStatus(status int, allowed []int) Rule {
	return &StatusRule{Status: status, Allowed: allowed}
}

// Check checks if the status is allowed.
func (r *StatusRule) Check() error {
	if !contains(r.Allowed, r.Status) {
		r.Errors = append(r.Errors, fmt.Sprintf("status is not allowed: %d", r.Status))
	}

	if len(r.Errors) > 0 {
		return fmt.Errorf("%s", strings.Join(r.Errors, ", "))
	}

	return nil
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
