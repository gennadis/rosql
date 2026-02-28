package query

import (
	"fmt"
	"regexp"
	"strings"
)

var forbidden = []string{
	"insert",
	"update",
	"delete",
	"drop",
	"alter",
	"create",
	"grant",
	"revoke",
	"truncate",
	"copy",
	"call",
	"do",
}

var forbiddenPatterns = func() []*regexp.Regexp {
	res := make([]*regexp.Regexp, 0, len(forbidden))
	for _, word := range forbidden {
		res = append(res,
			regexp.MustCompile(`\b`+regexp.QuoteMeta(word)+`\b`),
		)
	}
	return res
}()

func ValidateReadOnly(q string) error {
	clean := removeComments(q)
	lower := strings.ToLower(strings.TrimSpace(clean))

	// simple multi statement guard
	if strings.Count(lower, ";") > 1 {
		return fmt.Errorf("multiple statements are not allowed")
	}

	fields := strings.Fields(lower)
	if len(fields) == 0 {
		return fmt.Errorf("empty query")
	}

	// only SELECT and WITH allowed
	if fields[0] != "select" && fields[0] != "with" {
		return fmt.Errorf("only SELECT queries allowed but found: %s", fields[0])
	}

	// forbidden keyword detection
	for _, r := range forbiddenPatterns {
		if match := r.FindString(lower); match != "" {
			return fmt.Errorf("forbidden keyword detected: %s", match)
		}
	}

	return nil
}

func removeComments(q string) string {
	lines := strings.Split(q, "\n")

	for i := range lines {
		if before, _, found := strings.Cut(lines[i], "--"); found {
			lines[i] = before
		}
	}

	return strings.Join(lines, "\n")
}
