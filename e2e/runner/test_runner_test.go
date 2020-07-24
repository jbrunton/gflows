package runner

import (
	"errors"
	"testing"

	"github.com/jbrunton/gflows/io"
	"github.com/stretchr/testify/mock"
)

type mockAssertions struct {
	mock.Mock
}

func (a *mockAssertions) NoError(err error, msgAndArgs ...interface{}) {
	a.Called(append([]interface{}{err}, msgAndArgs...)...)
}

func (a *mockAssertions) EqualError(theError error, errString string, msgAndArgs ...interface{}) {
	a.Called(append([]interface{}{theError, errString}, msgAndArgs...)...)
}

func (a *mockAssertions) True(value bool, msgAndArgs ...interface{}) {
	a.Called(append([]interface{}{value}, msgAndArgs...)...)
}
func (a *mockAssertions) Equal(expected, actual interface{}, msgAndArgs ...interface{}) {
	a.Called(append([]interface{}{expected, actual}, msgAndArgs...)...)
}

func TestRunnerUpToDate(t *testing.T) {
	osFs := io.CreateOsFs()
	assertions := &mockAssertions{}
	runner := NewTestRunner(osFs, "./tests/test-runner-up-to-date.yml", true, assertions)

	assertions.On(
		"NoError",
		nil,
		"Unexpected error (%s)", "./tests/test-runner-up-to-date.yml")
	assertions.On(
		"Equal",
		"Checking test ... OK\nWorkflows up to date\n",
		"Checking test ... OK\nWorkflows up to date\n",
		"Unexpected output (%s)", "./tests/test-runner-up-to-date.yml")

	runner.Run()

	assertions.AssertNumberOfCalls(t, "Errorf", 0)
}

func TestRunnerOutOfDate(t *testing.T) {
	osFs := io.CreateOsFs()
	assertions := &mockAssertions{}
	runner := NewTestRunner(osFs, "./tests/test-runner-out-of-date.yml", true, assertions)

	assertions.On(
		"EqualError",
		errors.New("workflow validation failed"),
		"workflow validation failed",
		"Unexpected error (%s)", "./tests/test-runner-out-of-date.yml")
	assertions.On(
		"Equal",
		"Checking test ... FAILED\n  Content is out of date for \"test\" (.github/workflows/test.yml)\n  ► Run \"gflows workflow update\" to update\n",
		"Checking test ... FAILED\n  Content is out of date for \"test\" (.github/workflows/test.yml)\n  ► Run \"gflows workflow update\" to update\n",
		"Unexpected output (%s)", "./tests/test-runner-out-of-date.yml")

	runner.Run()

	assertions.AssertNumberOfCalls(t, "Errorf", 0)
}
