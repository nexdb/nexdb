package rest

import (
	"encoding/json"
	"net/http"
)

// Response is a response from a handler.
type Response struct {
	status int
	body   []byte
	header http.Header
	err    error
}

// Status returns the response status code.
func (r *Response) Status() int {
	return r.status
}

// Body returns the response body.
func (r *Response) Body() []byte {
	return r.body
}

// Header returns the response header.
func (r *Response) Header() http.Header {
	return r.header
}

// Error returns the response error.
func (r *Response) Error() error {
	return r.err
}

// JsonResponse returns a response with a content-type of application/json.
func JsonResponse(opts ...Option) *Response {
	r := &Response{
		status: http.StatusOK,
		body:   []byte{},
		header: http.Header{
			"Content-Type": []string{"application/json"},
		},
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Option is a function that modifies a response.
type Option func(*Response)

// SetStatus sets the status code of the response.
func SetStatus(status int) Option {
	return func(r *Response) {
		r.status = status
	}
}

// SetBody will marshal the body to JSON if the content-type is application/json.
func SetBody(body interface{}) Option {
	return func(r *Response) {
		if r.header.Get("Content-Type") == "application/json" {
			r.body, _ = json.Marshal(body)
		}
	}
}

// SetWrap will wrap the body in a map with the given key, should be called after SetBody.
func SetWrap(key string) Option {
	return func(r *Response) {
		if r.header.Get("Content-Type") == "application/json" {
			r.body, _ = json.Marshal(map[string]interface{}{
				key: json.RawMessage(r.body),
			})
		}
	}
}

// WithError will set the error on the response.
func WithError(err error) Option {
	return func(r *Response) {
		r.err = err
	}
}
