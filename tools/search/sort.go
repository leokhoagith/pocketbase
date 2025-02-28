package search

import (
	"fmt"
	"strings"
)

// sort field directions
const (
	SortAsc  string = "ASC"
	SortDesc string = "DESC"
)

// SortField defines a single search sort field.
type SortField struct {
	Name      string `json:"name"`
	Direction string `json:"direction"`
}

// BuildExpr resolves the sort field into a valid db sort expression.
func (s *SortField) BuildExpr(fieldResolver FieldResolver) (string, error) {
	name, params, err := fieldResolver.Resolve(s.Name)

	// invalidate empty fields and non-column identifiers
	if err != nil || len(params) > 0 || name == "" || strings.ToLower(name) == "null" {
		return "", fmt.Errorf("Invalid sort field %q.", s.Name)
	}

	return fmt.Sprintf("%s %s", name, s.Direction), nil
}

// ParseSortFromString parses the provided string expression
// into a slice of SortFields.
//
// Example:
//	fields := search.ParseSortFromString("-name,+created")
func ParseSortFromString(str string) []SortField {
	result := []SortField{}

	data := strings.Split(str, ",")

	for _, field := range data {
		// trim whitespaces
		field = strings.TrimSpace(field)

		var dir string
		if strings.HasPrefix(field, "-") {
			dir = SortDesc
			field = strings.TrimPrefix(field, "-")
		} else {
			dir = SortAsc
			field = strings.TrimPrefix(field, "+")
		}

		result = append(result, SortField{field, dir})
	}

	return result
}
