package rest_test

import (
	"net/http"
	"testing"

	"github.com/nexdb/nexdb/pkg/api/rest"

	"github.com/stretchr/testify/assert"
)

func TestResponse(t *testing.T) {
	for _, tc := range []struct {
		name     string
		response *rest.Response
		verify   func(*testing.T, *rest.Response)
	}{
		{
			name:     "default response should be 200 OK with no body and content-type application/json",
			response: rest.JsonResponse(),
			verify: func(t *testing.T, resp *rest.Response) {
				assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
				assert.Equal(t, http.StatusOK, resp.Status())
				assert.Equal(t, []byte{}, resp.Body())
			},
		},
		{
			name: "response with status code 400 should be 400 Bad Request with no body and content-type application/json",
			response: rest.JsonResponse(
				rest.SetStatus(http.StatusBadRequest),
			),
			verify: func(t *testing.T, resp *rest.Response) {
				assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
				assert.Equal(t, http.StatusBadRequest, resp.Status())
				assert.Equal(t, []byte{}, resp.Body())
			},
		},
		{
			name: "response with body should have that body and content-type application/json",
			response: rest.JsonResponse(
				rest.SetBody(map[string]interface{}{"message": "ok"}),
			),
			verify: func(t *testing.T, resp *rest.Response) {
				assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
				assert.Equal(t, http.StatusOK, resp.Status())
				assert.Equal(t, []byte(`{"message":"ok"}`), resp.Body())
			},
		},
		{
			name: "response with body and wrap should have that body wrapped and content-type application/json",
			response: rest.JsonResponse(
				rest.SetBody(map[string]interface{}{"message": "ok"}),
				rest.SetWrap("data"),
			),
			verify: func(t *testing.T, resp *rest.Response) {
				assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
				assert.Equal(t, http.StatusOK, resp.Status())
				assert.Equal(t, []byte(`{"data":{"message":"ok"}}`), resp.Body())
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.verify(t, tc.response)
		})
	}
}
