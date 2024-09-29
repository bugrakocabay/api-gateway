package config

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var (
	pathNotDefinedError       = errors.New("path is not defined on route configuration")
	invalidPathFormatError    = errors.New("the path format is invalid")
	targetNotDefinedError     = errors.New("target is not defined on route configuration")
	methodNotDefinedError     = errors.New("method is not defined on route configuration")
	invalidMethodError        = errors.New("the method defined on route configuration is invalid")
	duplicateQueryParamsError = errors.New("query parameters must be unique")
	invalidLimitError         = errors.New("limit should be a positive integer on route configuration")
)

// Route defines the structure for route configuration.
type Route struct {
	Path        string   `json:"path"`
	Target      *Target  `json:"target"`
	Method      string   `json:"method"`
	QueryParams []string `json:"query_params"`
	Limit       int64    `json:"limit"`
}

func (r *Route) validateRoute() error {
	if len(r.Path) == 0 {
		return fmt.Errorf("validation error: %w (route: %v)", pathNotDefinedError, r)
	}

	if !isValidPathFormat(r.Path) {
		return fmt.Errorf("validation error: %w (path: %s)", invalidPathFormatError, r.Path)
	}

	if r.Target == nil {
		return fmt.Errorf("validation error: %w (route: %v)", targetNotDefinedError, r)
	}

	if len(r.Method) == 0 {
		return fmt.Errorf("validation error: %w (route: %v)", methodNotDefinedError, r)
	}

	if !isValidHTTPMethod(r.Method) {
		return fmt.Errorf("validation error: %w (method: %s)", invalidMethodError, r.Method)
	}

	if !isDuplicateQueryParams(r.QueryParams) {
		return fmt.Errorf("validation error: %w (query_params: %v)", duplicateQueryParamsError, r.QueryParams)
	}

	if r.Limit < 0 {
		return fmt.Errorf("validation error: %w (limit: %d)", invalidLimitError, r.Limit)
	}

	return nil
}

// isValidHTTPMethod validates whether a string is a valid HTTP method
func isValidHTTPMethod(method string) bool {
	method = strings.ToUpper(method)

	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete,
		http.MethodPatch, http.MethodOptions, http.MethodHead, http.MethodConnect, http.MethodTrace:
		return true
	default:
		return false
	}
}

func isValidPathFormat(path string) bool {
	return strings.HasPrefix(path, "/") && !strings.ContainsAny(path, "?#")
}

func isDuplicateQueryParams(queryParams []string) bool {
	seen := make(map[string]struct{})
	for _, param := range queryParams {
		if _, exists := seen[param]; exists {
			return false
		}
		seen[param] = struct{}{}
	}
	return true
}
