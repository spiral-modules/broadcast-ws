package ws

import (
	"bytes"
	"io"
	"net/http"
)

type accessValidator struct {
	buffer *bytes.Buffer
	header http.Header
	status int
}

func newValidator() *accessValidator {
	return &accessValidator{
		buffer: bytes.NewBuffer(nil),
		header: make(http.Header),
	}
}

// copy all content to parent response writer.
func (w *accessValidator) copy(rw http.ResponseWriter) {
	rw.WriteHeader(w.status)

	for k, v := range w.header {
		for _, vv := range v {
			rw.Header().Add(k, vv)
		}
	}

	io.Copy(rw, w.buffer)
}

// Header returns the header map that will be sent by WriteHeader.
func (w *accessValidator) Header() http.Header {
	return w.header
}

// Write writes the data to the connection as part of an HTTP reply.
func (w *accessValidator) Write(p []byte) (int, error) {
	return w.buffer.Write(p)
}

// WriteHeader sends an HTTP response header with the provided status code.
func (w *accessValidator) WriteHeader(statusCode int) {
	w.status = statusCode
}

// IsOK returns true if response contained 200 status code.
func (w *accessValidator) IsOK() bool {
	return w.status == 200
}

// Body returns response body to rely to user.
func (w *accessValidator) Body() []byte {
	return w.buffer.Bytes()
}
