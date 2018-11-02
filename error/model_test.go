package error

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddDetails(t *testing.T) {
	assert := assert.New(t)
	err := NewRestError("", "", nil)
	err.AddDetail("foo")
	assert.Equal("foo", err.Details[0])
	ferr := FieldError{}
	err.AddDetail(ferr)
	assert.Equal(ferr, err.Details[1])
}
