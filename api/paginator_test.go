package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type testItem struct {
	ID string `json:"id"`
}

func TestPaginatorPaginate(t *testing.T) {
	Convey("Given a Paginator and a PaginatedHandler that returns a slice of testItem", t, func() {
		p := NewPaginator(10, 0, 100)
		handler := func(w http.ResponseWriter, r *http.Request, limit, offset int) (interface{}, int, error) {
			items := []testItem{
				{ID: "item1"},
				{ID: "item2"},
			}
			return items, len(items), nil
		}
		paginate := p.Paginate(handler)

		Convey("When a valid request is made", func() {
			req := httptest.NewRequest(http.MethodGet, "/test?limit=2&offset=0", http.NoBody)
			resp := httptest.NewRecorder()
			paginate(resp, req)

			Convey("Then the response should contain the paginated items", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)
				var paginatedResponse struct {
					Items      []testItem `json:"items"`
					Count      int        `json:"count"`
					Limit      int        `json:"limit"`
					Offset     int        `json:"offset"`
					TotalCount int        `json:"total_count"`
				}
				err := json.Unmarshal(resp.Body.Bytes(), &paginatedResponse)
				So(err, ShouldBeNil)
				So(len(paginatedResponse.Items), ShouldEqual, 2)
				So(paginatedResponse.Count, ShouldEqual, 2)
				So(paginatedResponse.Limit, ShouldEqual, 2)
				So(paginatedResponse.Offset, ShouldEqual, 0)
				So(paginatedResponse.TotalCount, ShouldEqual, 2)
			})
		})

		Convey("When an invalid limit is provided", func() {
			req := httptest.NewRequest(http.MethodGet, "/test?limit=invalid", http.NoBody)
			resp := httptest.NewRecorder()
			paginate(resp, req)

			Convey("Then a bad request is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Convey("When an invalid offset is provided", func() {
			req := httptest.NewRequest(http.MethodGet, "/test?offset=invalid", http.NoBody)
			resp := httptest.NewRecorder()
			paginate(resp, req)

			Convey("Then a bad request is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Convey("When the limit exceeds the max limit", func() {
			req := httptest.NewRequest(http.MethodGet, "/test?limit=1000", http.NoBody)
			resp := httptest.NewRecorder()
			paginate(resp, req)

			Convey("Then a bad request is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Convey("When no limit or offset is provided", func() {
			req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			resp := httptest.NewRecorder()
			paginate(resp, req)

			Convey("Then the default limit and offset are used", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)
				var paginatedResponse struct {
					Items      []testItem `json:"items"`
					Count      int        `json:"count"`
					Limit      int        `json:"limit"`
					Offset     int        `json:"offset"`
					TotalCount int        `json:"total_count"`
				}
				err := json.Unmarshal(resp.Body.Bytes(), &paginatedResponse)
				So(err, ShouldBeNil)
				So(paginatedResponse.Limit, ShouldEqual, 10)
				So(paginatedResponse.Offset, ShouldEqual, 0)
			})
		})
	})

	Convey("Given a Paginator and a PaginatedHandler that returns an error", t, func() {
		p := NewPaginator(10, 0, 100)
		handler := func(w http.ResponseWriter, r *http.Request, limit, offset int) (interface{}, int, error) {
			return nil, 0, http.ErrServerClosed
		}
		paginate := p.Paginate(handler)

		Convey("When a request is made", func() {
			req := httptest.NewRequest(http.MethodGet, "/test?limit=2&offset=0", http.NoBody)
			resp := httptest.NewRecorder()
			paginate(resp, req)

			Convey("Then an internal server error is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})

	Convey("Given a Paginator and a PaginatedHandler that returns an empty slice", t, func() {
		p := NewPaginator(10, 0, 100)
		handler := func(w http.ResponseWriter, r *http.Request, limit, offset int) (interface{}, int, error) {
			return []testItem{}, 0, nil
		}
		paginate := p.Paginate(handler)

		Convey("When a request is made", func() {
			req := httptest.NewRequest(http.MethodGet, "/test?limit=2&offset=0", http.NoBody)
			resp := httptest.NewRecorder()
			paginate(resp, req)

			Convey("Then the response should contain an empty items slice", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)
				var paginatedResponse struct {
					Items      []testItem `json:"items"`
					Count      int        `json:"count"`
					Limit      int        `json:"limit"`
					Offset     int        `json:"offset"`
					TotalCount int        `json:"total_count"`
				}
				err := json.Unmarshal(resp.Body.Bytes(), &paginatedResponse)
				So(err, ShouldBeNil)
				So(len(paginatedResponse.Items), ShouldEqual, 0)
				So(paginatedResponse.Count, ShouldEqual, 0)
				So(paginatedResponse.TotalCount, ShouldEqual, 0)
			})
		})
	})
}

func TestListLength(t *testing.T) {
	Convey("Given a slice of testItem", t, func() {
		items := []testItem{{ID: "item1"}, {ID: "item2"}}
		Convey("When listLength is called", func() {
			length := listLength(items)
			Convey("Then it should return the correct length", func() {
				So(length, ShouldEqual, 2)
			})
		})
	})
}
