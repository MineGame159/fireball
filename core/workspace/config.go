package workspace

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type Config struct {
	Name      string
	Src       string
	Namespace ConfigNamespace

	LinkLibraries []string
}

type ConfigNamespace []string

var identifierPattern = regexp.MustCompile("^[a-zA-Z_][a-zA-Z_0-9]*$")

func (c *ConfigNamespace) MarshalText() (text []byte, err error) {
	return []byte(c.String()), nil
}

func (c *ConfigNamespace) UnmarshalText(text []byte) error {
	parts := strings.Split(string(text), ".")

	for _, part := range parts {
		if identifierPattern.MatchString(part) {
			*c = append(*c, part)
		} else {
			return errors.New(fmt.Sprintf("config: Namespace: '%s' is not a valid namespace part", part))
		}
	}

	return nil
}

func (c *ConfigNamespace) String() string {
	sb := strings.Builder{}

	for i, part := range *c {
		if i > 0 {
			sb.WriteRune('.')
		}

		sb.WriteString(part)
	}

	return sb.String()
}
