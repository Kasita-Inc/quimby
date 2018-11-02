package httptest

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
)

/******************************************************
 *          Supporting code for tests                 *
 ******************************************************/

func createRequest(method string) (request *http.Request) {
	host := "http://127.0.0.1/"
	request = httptest.NewRequest(method, host, nil)
	u, _ := url.Parse("http://127.0.0.1/")
	request.URL = u
	request.RequestURI = u.RequestURI()
	request.Header.Add("Content-Type", "application/json")
	return request

}

// CreateTestRequest instantiates an http.Request that can be used for testing
func CreateTestRequest() *http.Request {
	return createRequest(http.MethodGet)
}

// CreateTestRequestBody creates a test http.Request with the body set
func CreateTestRequestBody(method string, data interface{}) (request *http.Request) {
	request = createRequest(method)
	body, _ := json.Marshal(data)
	request.Body = ioutil.NopCloser(bytes.NewReader(body))
	request.ContentLength = int64(len(body))
	return request
}
