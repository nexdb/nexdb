package rest

import (
	"net/http"

	"github.com/nexdb/nexdb/pkg/errors"
)

// JsonHandler is a handler that returns a response.
type JsonHandler func(r *http.Request) *Response

// ServeHTTP implements http.Handler.
func (h JsonHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := h(r)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for k, v := range resp.Header() {
		w.Header()[k] = v
	}

	respErr := resp.Error()
	if respErr != nil {
		if resp.Status() == http.StatusOK {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(resp.Status())
		}

		// if a system error, return a generic error message
		respErr, ok := respErr.(*errors.Error)
		if !ok {
			w.Write([]byte(`{"error":"system error"}`))
			return
		}

		w.Write([]byte(`{"error":"` + respErr.Error() + `","code":` + respErr.Code().ToString() + `}`))
		return
	}

	w.WriteHeader(resp.Status())
	w.Write(resp.Body())
}
