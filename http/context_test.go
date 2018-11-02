package http

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Kasita-Inc/gadget/stringutil"
	qerror "github.com/Kasita-Inc/quimby/error"
)

type TestModel struct {
	Name string `json:"name"`
}

type TestModel2 struct {
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
}

type TestModel3 struct {
	FirstName []string `json:"firstname"`
	LastName  string   `json:"lastname"`
}

type FakeBody struct {
	Content string
	Size    int
	Error   error
}

func (body FakeBody) Read(p []byte) (n int, err error) {
	content := []byte(body.Content)
	copy(p, content)
	return len(content), body.Error
}

func (body FakeBody) Close() error {
	return body.Error
}

/******************************************************
 *                      Tests                         *
 ******************************************************/

func TestSetError(t *testing.T) {
	assert := assert.New(t)
	context := Context{}
	expectedError := qerror.NewRestError(qerror.MethodNotAllowed, "", nil)
	context.SetError(expectedError, http.StatusMethodNotAllowed)
	assert.Equal(expectedError, context.Error)
	assert.Equal(http.StatusMethodNotAllowed, context.Status())
	assert.True(context.HasError())

	assert.False(context.SetResponse("model", http.StatusOK))
	assert.Equal(nil, context.Model)
	assert.Equal(http.StatusMethodNotAllowed, context.Status())
	assert.True(context.HasError())
}

func TestAddError(t *testing.T) {
	assert := assert.New(t)
	context := Context{}
	fieldError := qerror.FieldError{Code: qerror.CannotBeBlank, Field: "name"}
	context.AddError(fieldError)

	expectedError := &qerror.RestError{Code: qerror.ValidationError}
	expectedError.Details = append(expectedError.Details, fieldError)

	assert.Equal(expectedError, context.Error)
	assert.Equal(fieldError, context.Error.Details[0])
	assert.True(context.HasError())
}

func TestSetResponse(t *testing.T) {
	assert := assert.New(t)
	context := Context{}
	assert.True(context.SetResponse("model", http.StatusOK))
	assert.Equal("model", context.Model)
	assert.Equal(http.StatusOK, context.Status())
	assert.False(context.HasError())
}

func TestCreateContext(t *testing.T) {
	assert := assert.New(t)

	u, _ := url.Parse("http://127.0.0.1/")

	w := testResponseWriter{}
	r := http.Request{
		URL:        u,
		RequestURI: u.RequestURI(),
	}
	c := NewTestController("HTTP Test")
	c.Routes = append(c.Routes, "/")
	router := CreateRouter(&c)
	context := CreateContext(w, &r, router)
	context.Extended["foo"] = "bar"

	assert.False(context.HasError())
}

func TestCreateContextWithFailingAuth(t *testing.T) {
	assert := assert.New(t)

	u, _ := url.Parse("http://127.0.0.1/")

	w := testResponseWriter{}
	r := http.Request{
		URL:        u,
		RequestURI: u.RequestURI(),
	}
	c := NewNoAuthTestController("HTTP Test")
	c.Routes = append(c.Routes, "/")
	router := CreateRouter(&c)
	context := CreateContext(w, &r, router)

	assert.True(context.HasError())
	assert.Equal(qerror.AuthenticationFailed, context.Error.Code)
}

func TestCreateContextBadParameters(t *testing.T) {
	assert := assert.New(t)

	u, _ := url.Parse("http://127.0.0.1/?%zzzzz")

	w := testResponseWriter{}
	r := http.Request{
		URL:        u,
		RequestURI: u.RequestURI(),
	}
	c := NewTestController("HTTP Test")
	c.Routes = append(c.Routes, "/")
	router := CreateRouter(&c)
	context := CreateContext(w, &r, router)

	assert.True(context.HasError())
	assert.Equal(qerror.MalformedURL, context.Error.Code)
}

func TestCreateContextBadRoute(t *testing.T) {
	assert := assert.New(t)

	u, _ := url.Parse("http://127.0.0.1/foo")

	w := testResponseWriter{}
	r := http.Request{
		URL:        u,
		RequestURI: u.RequestURI(),
	}
	c := NewTestController("HTTP Test")
	router := CreateRouter(&c)
	context := CreateContext(w, &r, router)

	assert.True(context.HasError())
	assert.Equal(qerror.InvalidRoute, context.Error.Code)
	assert.Equal(http.StatusBadRequest, context.Status())
}

func TestCreateContextBadTemplate(t *testing.T) {
	assert := assert.New(t)

	u, _ := url.Parse("http://127.0.0.1/foo/bar")

	w := testResponseWriter{}
	r := http.Request{
		URL:        u,
		RequestURI: u.RequestURI(),
	}
	c := NewTestController("HTTP Test")
	c.Routes = append(c.Routes, "foo/{{id}}{{id2}}")
	router := CreateRouter(&c)
	router.AddController(&c)
	context := CreateContext(w, &r, router)

	assert.True(context.HasError())
	assert.Equal(qerror.InvalidRoute, context.Error.Code)
	assert.Equal(http.StatusInternalServerError, context.Status())
}

func TestCreateContextQueryStringNotInParameters(t *testing.T) {
	assert := assert.New(t)

	u, _ := url.Parse("http://127.0.0.1/foo/bar/?awef=qwer")

	w := testResponseWriter{}
	r := http.Request{
		URL:        u,
		RequestURI: u.RequestURI(),
	}
	c := NewTestController("HTTP Test")
	c.Routes = append(c.Routes, "foo/{{id}}")
	router := CreateRouter(&c)
	router.AddController(&c)

	context := CreateContext(w, &r, router)
	assert.False(context.HasError())

	actual, ok := context.URIParameters["id"]
	assert.True(ok, "Parameter should exist in context.")
	assert.Equal("bar", actual)
}

func TestRead(t *testing.T) {
	assert := assert.New(t)

	b := FakeBody{
		Content: "foo",
		Error:   io.EOF,
	}
	r := http.Request{
		ContentLength: int64(len(b.Content)),
		Body:          b,
	}
	context := Context{
		Request: &r,
	}
	body, err := context.Read()

	assert.Equal("foo", string(body))
	assert.Nil(err)
}

func TestReadEmpty(t *testing.T) {
	assert := assert.New(t)

	r := http.Request{}
	context := Context{
		Request: &r,
	}
	body, err := context.Read()

	assert.Nil(body)
	assert.EqualError(err, NewNoContentError("", "").Error())
}

func TestReadObject_withJSON(t *testing.T) {
	assert := assert.New(t)

	s := TestModel{
		Name: "foo",
	}
	j, _ := json.Marshal(s)
	b := FakeBody{
		Content: string(j),
		Error:   io.EOF,
	}
	r := http.Request{
		Header: map[string][]string{
			"Content-Type": {"application/json"},
		},
		ContentLength: int64(len(b.Content)),
		Body:          b,
	}
	context := Context{
		Request: &r,
	}
	model := TestModel{}
	err := context.ReadObject(&model)

	assert.Equal(s, model)
	assert.Nil(err)
}

func TestReadObject_withJSON_NoContentTypeHeader(t *testing.T) {
	assert := assert.New(t)

	s := TestModel{
		Name: "foo",
	}
	j, _ := json.Marshal(s)
	b := FakeBody{
		Content: string(j),
		Error:   io.EOF,
	}
	r := http.Request{
		ContentLength: int64(len(b.Content)),
		Body:          b,
	}
	context := Context{
		Request: &r,
	}
	model := TestModel{}
	err := context.ReadObject(&model)

	assert.Error(err)
}

func TestReadObject_withJSON_BadInput(t *testing.T) {
	assert := assert.New(t)

	b := FakeBody{
		Content: "asdf",
		Error:   io.EOF,
	}
	r := http.Request{
		ContentLength: int64(len(b.Content)),
		Body:          b,
	}
	context := Context{
		Request: &r,
	}
	body := &TestModel{}

	assert.False(context.HasError())

	err := context.ReadObject(body)

	assert.Error(err)
	assert.True(context.HasError())
	assert.Equal(http.StatusNotAcceptable, context.Status())
}

func TestReadObject_withJSON_Empty(t *testing.T) {
	assert := assert.New(t)

	r := http.Request{ContentLength: 0}
	context := Context{
		Request: &r,
	}
	s := ""
	err := context.ReadObject(s)

	assert.Equal("", s)
	assert.EqualError(err, NewNoContentError("", "").Error())
}

func TestReadObject_withForm(t *testing.T) {
	assert := assert.New(t)

	v := url.Values{}
	v.Add("name", "RoundyMcTrashcan")
	v.Add("grant", "auth_code")

	expected := TestModel{
		Name: "RoundyMcTrashcan",
	}

	b := FakeBody{
		Content: v.Encode(),
		Error:   io.EOF,
	}
	r := http.Request{
		Method: http.MethodPost,
		Header: map[string][]string{
			"Content-Type": {"application/x-www-form-urlencoded"},
		},
		ContentLength: int64(len(b.Content)),
		Body:          b,
	}
	context := Context{
		Request: &r,
	}
	model := TestModel{}
	err := context.ReadObject(&model)

	assert.Equal(expected, model)
	assert.Nil(err)
}

func TestReadObject_withForm_Casing(t *testing.T) {
	assert := assert.New(t)

	v := url.Values{}
	v.Add("first_name", "Roundy")
	v.Add("last_name", "McTrashcan")
	v.Add("grant", "auth_code")

	expected := TestModel2{
		FirstName: "Roundy",
		LastName:  "McTrashcan",
	}

	b := FakeBody{
		Content: v.Encode(),
		Error:   io.EOF,
	}
	r := http.Request{
		Method: http.MethodPost,
		Header: map[string][]string{
			"Content-Type": {"application/x-www-form-urlencoded"},
		},
		ContentLength: int64(len(b.Content)),
		Body:          b,
	}
	context := Context{
		Request: &r,
	}
	model := TestModel2{}
	err := context.ReadObject(&model)

	assert.Equal(expected, model)
	assert.Nil(err)
}

func TestReadObject_withForm_Arrays(t *testing.T) {
	assert := assert.New(t)

	v := "first_name[]=Roundy&first_name[]=Roundy2&first_name=Roundy3&grant=auth_code&last_name=McTrashcan"

	expected := TestModel3{
		FirstName: []string{"Roundy3", "Roundy", "Roundy2"},
		LastName:  "McTrashcan",
	}

	b := FakeBody{
		Content: v,
		Error:   io.EOF,
	}
	r := http.Request{
		Method: http.MethodPost,
		Header: map[string][]string{
			"Content-Type": {"application/x-www-form-urlencoded"},
		},
		ContentLength: int64(len(b.Content)),
		Body:          b,
	}
	context := Context{
		Request: &r,
	}
	model := TestModel3{}
	err := context.ReadObject(&model)

	assert.Equal(expected, model)
	assert.Nil(err)
}

func TestReadObject_withForm_EmptyQueryParam(t *testing.T) {
	assert := assert.New(t)

	v := url.Values{}
	v.Add("name", "")
	v.Add("grant", "auth_code")

	expected := TestModel{}

	b := FakeBody{
		Content: v.Encode(),
		Error:   io.EOF,
	}
	r := http.Request{
		Method: http.MethodPost,
		Header: map[string][]string{
			"Content-Type": {"application/x-www-form-urlencoded"},
		},
		ContentLength: int64(len(b.Content)),
		Body:          b,
	}
	context := Context{
		Request: &r,
	}
	model := TestModel{}
	err := context.ReadObject(&model)

	assert.Equal(expected, model)
	assert.Nil(err)
}

func TestWrite(t *testing.T) {
	assert := assert.New(t)

	writerBody := make([]byte, 20)
	w := testResponseWriter{Body: &writerBody}
	context := Context{
		Response: w,
	}

	context.Write("foo")
	assert.Equal("foo", stringutil.NullTerminatedString(writerBody))
}
