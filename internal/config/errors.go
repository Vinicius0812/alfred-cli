package config

import (
	"fmt"
	"strings"
)

type ValidationError struct {
	File    string
	Path    string
	Message string
}

func (e ValidationError) Error() string {
	if e.Path == "" {
		return fmt.Sprintf("%s: %s", e.File, e.Message)
	}
	return fmt.Sprintf("%s: %s: %s", e.File, e.Path, e.Message)
}

type ErrorList []ValidationError

func (l ErrorList) Error() string {
	if len(l) == 0 {
		return ""
	}

	parts := make([]string, 0, len(l))
	for _, err := range l {
		parts = append(parts, err.Error())
	}
	return strings.Join(parts, "\n")
}
