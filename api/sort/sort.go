package api

import (
	"strings"

	appErrors "github.com/ONSdigital/dis-migration-service/errors"
)

// SortParameterField is the field you are sorting by
type SortParameterField string

// SortParameterDirection is the direction which you are sorting by
type SortParameterDirection string

const (
	// SortParameterFieldJobNumber is the name of the field,
	// job number, in the sort parameter.
	SortParameterFieldJobNumber SortParameterField = "job_number"
	// SortParameterFieldLabel is the name of the field,
	// label, in the sort parameter.
	SortParameterFieldLabel SortParameterField = "label"
	// SortParameterDirectionAsc is the name of the direction,
	// ascending, in the sort parameter.
	SortParameterDirectionAsc SortParameterDirection = "asc"
	// SortParameterDirectionDesc is the name of the direction,
	// descending, in the sort parameter.
	SortParameterDirectionDesc SortParameterDirection = "desc"
)

// ParseSortParameters parses the sort parameters of the form "field:direction"
func ParseSortParameters(sortParam []string) (field SortParameterField, direction SortParameterDirection, err error) {
	validSortParameterFields := []SortParameterField{
		SortParameterFieldJobNumber,
		SortParameterFieldLabel,
	}

	validSortParameterDirections := []SortParameterDirection{
		SortParameterDirectionAsc,
		SortParameterDirectionDesc,
	}

	for _, s := range sortParam {
		for _, p := range strings.Split(s, ",") {
			part := strings.TrimSpace(p)
			parts := strings.Split(part, ":")

			field = SortParameterField(strings.TrimSpace(parts[0]))
			direction = SortParameterDirection(strings.TrimSpace(parts[1]))

			if !IsValidSortParameter(field, validSortParameterFields) {
				return "", "", appErrors.ErrSortFieldInvalid
			}

			if !IsValidSortParameter(direction, validSortParameterDirections) {
				return "", "", appErrors.ErrSortDirectionInvalid
			}
		}
	}

	return field, direction, nil
}

// IsValidSortParameter validates sort parameter fields,
// and checks they are correct.
func IsValidSortParameter[SortParameter comparable](parameter SortParameter, validParameters []SortParameter) bool {
	for _, v := range validParameters {
		if parameter == v {
			return true
		}
	}
	return false
}
