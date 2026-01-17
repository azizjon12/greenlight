package data

import (
	"slices"
	"strings"

	"github.com/azizjon12/greenlight/internal/validator"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

// Check that the client-provided Sort field matches one of the entries in our safelist.
// If so, extract the column name from the Sort by stripping "-" if exists
func (f Filters) sortColumn() string {
	if slices.Contains(f.SortSafelist, f.Sort) {
		return strings.TrimPrefix(f.Sort, "-")
	}

	panic("unsafe sort parameter: " + f.Sort)
}

// Return the sort direction (ASC or DESC) depending on the prefix character of the Sort field
func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}

	return "ASC"
}

func (f Filters) limit() int {
	return f.PageSize
}

func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

func ValidateFilters(v *validator.Validator, f Filters) {
	// Check that page and page_size parameters contain sensible values
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10_000_000, "page", "must be a maximum of 10 million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")

	// Check that the sort parameter matches a value in the safelist
	v.Check(validator.PermittedValue(f.Sort, f.SortSafelist...), "sort", "invalid sort value")
}
