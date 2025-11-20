package api

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"

	appErrors "github.com/ONSdigital/dis-migration-service/errors"
)

// PaginatedHandler defines a handler function signature for paginated endpoints
type PaginatedHandler func(w http.ResponseWriter, r *http.Request, limit int, offset int) (items interface{}, totalCount int, err error)

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Items            interface{} `json:"items"`
	PaginationFields             // embedded, flattening fields into JSON
}

// PaginationFields represents the generic fields used to describe
// pagination in an API response
type PaginationFields struct {
	Count      int `json:"count"`
	Limit      int `json:"limit"`
	Offset     int `json:"offset"`
	TotalCount int `json:"total_count"`
}

// Paginator handles pagination logic for API endpoints
type Paginator struct {
	DefaultLimit    int
	DefaultOffset   int
	DefaultMaxLimit int
}

// NewPaginator creates a new Paginator with specified default values
func NewPaginator(defaultLimit, defaultOffset, defaultMaxLimit int) *Paginator {
	return &Paginator{
		DefaultLimit:    defaultLimit,
		DefaultOffset:   defaultOffset,
		DefaultMaxLimit: defaultMaxLimit,
	}
}

func (p *Paginator) getPaginationParameters(r *http.Request) (limit, offset int, errs []error) {
	var err error

	limitParameter := r.URL.Query().Get(QueryParameterLimit)
	offsetParameter := r.URL.Query().Get(QueryParameterOffset)

	offset = p.DefaultOffset
	limit = p.DefaultLimit

	if limitParameter != "" {
		limit, err = strconv.Atoi(limitParameter)
		if err != nil || limit < 0 {
			errs = append(errs, appErrors.ErrLimitInvalid)
		}
	}

	if limit > p.DefaultMaxLimit {
		errs = append(errs, appErrors.ErrLimitExceeded)
	}

	if offsetParameter != "" {
		offset, err = strconv.Atoi(offsetParameter)
		if err != nil || offset < 0 {
			errs = append(errs, appErrors.ErrOffsetInvalid)
		}
	}

	return limit, offset, errs
}

func generatePaginatedResponse(list interface{}, limit, offset, totalCount int) PaginatedResponse {
	if listLength(list) == 0 {
		list = []interface{}{}
	}

	return PaginatedResponse{
		Items: list,
		PaginationFields: PaginationFields{
			Count:      listLength(list),
			Offset:     offset,
			Limit:      limit,
			TotalCount: totalCount,
		},
	}
}

func listLength(list interface{}) int {
	l := reflect.ValueOf(list)
	return l.Len()
}

// Paginate is a middleware function that handles pagination for API endpoints
func (p *Paginator) Paginate(paginatedHandler PaginatedHandler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		limit, offset, errs := p.getPaginationParameters(r)
		if len(errs) > 0 {
			handleError(ctx, w, r, errs...)
			return
		}

		items, totalCount, err := paginatedHandler(w, r, limit, offset)
		if err != nil {
			handleError(ctx, w, r, err)
			return
		}

		paginatedResults := generatePaginatedResponse(items, limit, offset, totalCount)

		responseBytes, err := json.Marshal(paginatedResults)
		if err != nil {
			handleError(ctx, w, r, err)
			return
		}

		handleSuccess(ctx, w, r, http.StatusOK, responseBytes)
	}
}
