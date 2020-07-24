package runner

import "github.com/stretchr/testify/assert"

type Assertions interface {
	NoError(err error, msgAndArgs ...interface{})
	EqualError(theError error, errString string, msgAndArgs ...interface{})
	True(value bool, msgAndArgs ...interface{})
	Equal(expected, actual interface{}, msgAndArgs ...interface{})
}

type TestifyAssertions struct {
	t assert.TestingT
}

func NewTestifyAssertions(t assert.TestingT) Assertions {
	return &TestifyAssertions{t: t}
}

func (a *TestifyAssertions) NoError(err error, msgAndArgs ...interface{}) {
	assert.NoError(a.t, err, msgAndArgs...)
}

func (a *TestifyAssertions) EqualError(theError error, errString string, msgAndArgs ...interface{}) {
	assert.EqualError(a.t, theError, errString, msgAndArgs...)
}

func (a *TestifyAssertions) True(value bool, msgAndArgs ...interface{}) {
	assert.True(a.t, value, msgAndArgs...)
}
func (a *TestifyAssertions) Equal(expected, actual interface{}, msgAndArgs ...interface{}) {
	assert.Equal(a.t, expected, actual, msgAndArgs...)
}
