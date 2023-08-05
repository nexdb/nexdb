package rest_test

import (
	errs "errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nexdb/nexdb/pkg/api/rest"
	"github.com/nexdb/nexdb/pkg/errors"

	"github.com/stretchr/testify/assert"
)

func TestJsonHandler(t *testing.T) {
	for _, tc := range []struct {
		name    string
		handler rest.JsonHandler
		verify  func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "default response should be 200 OK with no body and content-type application/json",
			handler: func(r *http.Request) *rest.Response {
				return rest.JsonResponse()
			},
			verify: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rr.Code)
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
			},
		},
		{
			name: "response with status code 400 should be 400 Bad Request with no body and content-type application/json",
			handler: func(r *http.Request) *rest.Response {
				return rest.JsonResponse(
					rest.SetStatus(http.StatusBadRequest),
				)
			},
			verify: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
			},
		},
		{
			name: "response with error should return expected status code, body and content-type application/json",
			handler: func(r *http.Request) *rest.Response {
				return rest.JsonResponse(
					rest.SetStatus(http.StatusBadRequest),
					rest.WithError(errors.New(errors.ErrCollectionNameIsEmpty)),
				)
			},
			verify: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
				assert.Equal(t, `{"error":"collection name is empty","code":1000}`, rr.Body.String())
			},
		},
		{
			name: "response with error and without set status code should return 500, body and content-type application/json",
			handler: func(r *http.Request) *rest.Response {
				return rest.JsonResponse(
					rest.WithError(errors.New(errors.ErrCollectionNameIsEmpty)),
				)
			},
			verify: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rr.Code)
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
				assert.Equal(t, `{"error":"collection name is empty","code":1000}`, rr.Body.String())
			},
		},
		{
			name: "response with a system error should return 500, body and content-type application/json",
			handler: func(r *http.Request) *rest.Response {
				return rest.JsonResponse(
					rest.WithError(errs.New("a sensitive error that should not be shown")),
				)
			},
			verify: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rr.Code)
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
				assert.Equal(t, `{"error":"system error"}`, rr.Body.String())
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)

			rr := httptest.NewRecorder()

			// Call the handler directly.
			tc.handler.ServeHTTP(rr, req)

			tc.verify(t, rr)
		})
	}
}
