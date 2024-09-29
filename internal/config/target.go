package config

import (
	"errors"
	"fmt"
	"net/url"
)

var (
	hostNotDefinedError        = errors.New("host is not defined on target configuration")
	invalidMethodErrorOnTarget = errors.New("the method defined on target configuration is invalid")
)

// Target defines the structure for target endpoint configuration.
type Target struct {
	Host   string `json:"host"`
	Path   string `json:"path"`
	Method string `json:"method"`
}

func (t *Target) validateTarget() error {
	if len(t.Host) == 0 {
		return fmt.Errorf("validation error: %w (target: %v)", hostNotDefinedError, t)
	}

	if _, err := url.ParseRequestURI(t.Host); err != nil {
		return fmt.Errorf("validation error: %w (target: %v)", err, t)
	}

	if len(t.Method) > 0 && !isValidHTTPMethod(t.Method) {
		return fmt.Errorf("validation error: %w (target: %v)", invalidMethodErrorOnTarget, t)
	}

	return nil
}
