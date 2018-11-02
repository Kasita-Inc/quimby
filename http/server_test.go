package http

import (
	"encoding/json"
	"math"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Kasita-Inc/gadget/stringutil"
	qerror "github.com/Kasita-Inc/quimby/error"
)

/******************************************************
 *          Supporting code for tests                 *
 ******************************************************/
type testResponseWriter struct {
	Status *int
	Body   *[]byte
}

func (rw testResponseWriter) Header() http.Header {
	return http.Header{}
}

func (rw testResponseWriter) Write(b []byte) (int, error) {
	copy(*rw.Body, b)
	return 1, nil
}

func (rw testResponseWriter) WriteHeader(i int) {
	*rw.Status = i
}

type assertions struct {
}

var Assert = assertions{}

func (a *assertions) StringValueIn(key string, expectedValue string,
	m map[string]string, t *testing.T) {
	value, ok := m[key]
	if !ok {
		t.Errorf("Key '%s' not present in map.", key)
	}
	if expectedValue != value {
		t.Errorf("Unexpected value in map. Expected:'%s', Actual:'%s'",
			expectedValue, value)
	}
}

func (a *assertions) StringValueNotIn(key string, m map[string]string, t *testing.T) {
	_, ok := m[key]
	if !ok {
		t.Errorf("key '%s' should not be present in map.", key)
	}
}

/******************************************************
 *                      Tests                         *
 ******************************************************/

func TestServeHTTP(t *testing.T) {
	assert := assert.New(t)

	u, _ := url.Parse("http://127.0.0.1/")

	writerBody := make([]byte, 20)
	writerStatus := 0

	w := testResponseWriter{Body: &writerBody, Status: &writerStatus}
	r := http.Request{
		URL:        u,
		RequestURI: u.RequestURI(),
	}

	controller := NewTestController("HTTP Test")
	controller.Routes = append(controller.Routes, "/")
	server := CreateRESTServer(":8080", &controller)
	server.Router.AddController(&controller)

	methods := []string{
		http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodDelete, http.MethodOptions,
	}
	for _, method := range methods {
		controller.MethodCalled = ""
		r.Method = method
		server.ServeHTTP(w, &r)
		assert.Equal(method, controller.MethodCalled)
	}

	r.Method = "foo"
	controller.MethodCalled = ""
	server.ServeHTTP(&w, &r)
	assert.Equal(http.StatusMethodNotAllowed, writerStatus)
}

func TestCompleteRequestResponse(t *testing.T) {
	assert := assert.New(t)

	writerBody := make([]byte, 200)
	writerStatus := 0

	w := testResponseWriter{Body: &writerBody, Status: &writerStatus}
	r, err := http.NewRequest(http.MethodGet, "http://localhost", strings.NewReader(""))
	assert.NoError(err)

	controller := NewTestController("HTTP Test")
	server := CreateRESTServer(":8080", &controller)

	context := Context{Request: r, Response: w}
	context.SetResponse("OK", http.StatusOK)
	server.CompleteRequest(&context)
	assert.Equal("\"OK\"", stringutil.NullTerminatedString(writerBody))
	assert.Equal(http.StatusOK, writerStatus)
}

func TestCompleteRequestError(t *testing.T) {
	assert := assert.New(t)

	writerBody := make([]byte, 200)
	writerStatus := 0

	w := testResponseWriter{Body: &writerBody, Status: &writerStatus}
	r, e := http.NewRequest(http.MethodGet, "http://localhost", strings.NewReader(""))
	assert.NoError(e)

	controller := NewTestController("HTTP Test")
	server := CreateRESTServer(":8080", &controller)
	context := Context{Request: r, Response: w}
	err := qerror.NewRestError("testing", "", nil)
	context.SetError(err, http.StatusTeapot)
	server.CompleteRequest(&context)
	b, _ := json.Marshal(err)
	assert.Equal(string(b), stringutil.NullTerminatedString(writerBody))
	assert.Equal(http.StatusTeapot, writerStatus)
}

func TestCompleteRequestErrorCannotMarshall(t *testing.T) {
	assert := assert.New(t)

	writerBody := make([]byte, 200)
	writerStatus := 0

	w := testResponseWriter{Body: &writerBody, Status: &writerStatus}
	r, e := http.NewRequest(http.MethodGet, "http://localhost", strings.NewReader(""))
	assert.NoError(e)

	controller := NewTestController("HTTP Test")
	server := CreateRESTServer(":8080", &controller)
	context := Context{Request: r, Response: w}
	exp, _ := json.Marshal(qerror.NewRestError("system-error", "", nil))

	err := qerror.NewRestError("testing", "", nil)
	err.Details = append(err.Details, math.Inf(1))
	context.SetError(err, http.StatusTeapot)
	server.CompleteRequest(&context)
	assert.Equal(string(exp), stringutil.NullTerminatedString(writerBody))
	assert.Equal(http.StatusInternalServerError, writerStatus)
}

func TestCompleteRequestResponseCannotMarshal(t *testing.T) {
	assert := assert.New(t)

	writerBody := make([]byte, 200)
	writerStatus := 0

	w := testResponseWriter{Body: &writerBody, Status: &writerStatus}
	r, e := http.NewRequest(http.MethodGet, "http://localhost", strings.NewReader(""))
	assert.NoError(e)

	controller := NewTestController("HTTP Test")
	server := CreateRESTServer(":8080", &controller)
	context := Context{Request: r, Response: w}
	exp, _ := json.Marshal(qerror.NewRestError("system-error", "", nil))
	context.SetResponse(math.Inf(1), http.StatusAlreadyReported)
	server.CompleteRequest(&context)
	assert.Equal(string(exp), stringutil.NullTerminatedString(writerBody))
	assert.Equal(http.StatusInternalServerError, writerStatus)
}
