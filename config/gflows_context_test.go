package config

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/jbrunton/gflows/io"
	"github.com/jbrunton/gflows/io/pkg"
	"github.com/spf13/cobra"

	"github.com/stretchr/testify/assert"
)

func newTestContext() *GFlowsContext {
	opts := ContextOpts{
		ConfigPath:     ".gflows/config.yml",
		Engine:         "ytt",
		AllowNoContext: true,
	}
	fs := io.CreateMemFs()
	out := new(bytes.Buffer)
	logger := io.NewLogger(out, false, false)
	context, err := NewContext(fs, logger, opts)
	if err != nil {
		panic(err)
	}
	return context
}

func newTestCommand(run func(*cobra.Command, []string)) *cobra.Command {
	cmd := &cobra.Command{
		Use: "test",
		Run: run,
	}
	cmd.Flags().String("config", "", "")
	cmd.Flags().Bool("disable-colors", false, "")
	cmd.Flags().Bool("debug", false, "")
	return cmd
}

func TestResolvePath(t *testing.T) {
	context := newTestContext()
	assert.Equal(t, ".gflows/foo", context.ResolvePath("foo"))
	assert.Equal(t, ".gflows/foo", context.ResolvePath(".gflows/foo"))
	assert.Equal(t, "foo", context.ResolvePath("../foo"))
	assert.Equal(t, "http://example.com", context.ResolvePath("http://example.com"))
	assert.Equal(t, "https://example.com", context.ResolvePath("https://example.com"))
	assert.Equal(t, ".gflows/http/lib", context.ResolvePath("http/lib"))
	assert.Equal(t, "/foo", context.ResolvePath("/foo"))
}

func TestResolvePaths(t *testing.T) {
	context := newTestContext()
	assert.Equal(t, []string{".gflows/foo"}, context.ResolvePaths([]string{"foo"}))
}

func TestCreateContextOpts(t *testing.T) {
	scenarios := []struct {
		description  string
		setup        func(cmd *cobra.Command)
		expectedOpts ContextOpts
	}{
		{
			description: "default values",
			setup:       func(cmd *cobra.Command) {},
			expectedOpts: ContextOpts{
				ConfigPath:   ".gflows/config.yml",
				Engine:       "",
				EnableColors: true,
			},
		},
		{
			description: "override config",
			setup: func(cmd *cobra.Command) {
				cmd.SetArgs([]string{"test", "--config", "/my/config.yml"})
			},
			expectedOpts: ContextOpts{
				ConfigPath:   "/my/config.yml",
				Engine:       "",
				EnableColors: true,
			},
		},
		{
			description: "disable colors",
			setup: func(cmd *cobra.Command) {
				cmd.SetArgs([]string{"test", "--disable-colors"})
			},
			expectedOpts: ContextOpts{
				ConfigPath:   ".gflows/config.yml",
				Engine:       "",
				EnableColors: false,
			},
		},
		{
			description: "specify engine",
			setup: func(cmd *cobra.Command) {
				cmd.SetArgs([]string{"test", "--engine", "ytt"})
			},
			expectedOpts: ContextOpts{
				ConfigPath:   ".gflows/config.yml",
				Engine:       "ytt",
				EnableColors: true,
			},
		},
	}

	for _, scenario := range scenarios {
		cmd := newTestCommand(func(cmd *cobra.Command, args []string) {
			opts := CreateContextOpts(cmd)
			assert.Equal(t, scenario.expectedOpts, opts, "Unexpected opts for scenario %q", scenario.description)
		})
		scenario.setup(cmd)
		cmd.Execute()
	}
}

func TestGetPathInfo(t *testing.T) {
	context := newTestContext()

	info, err := context.GetPathInfo(".gflows/workflows/foo.jsonnet")

	assert.NoError(t, err)
	assert.Equal(t, &pkg.PathInfo{
		LocalPath:   ".gflows/workflows/foo.jsonnet",
		SourcePath:  ".gflows/workflows/foo.jsonnet",
		Description: ".gflows/workflows/foo.jsonnet",
	}, info)
}

func TestGetPathInfoErrors(t *testing.T) {
	context := newTestContext()
	_, err := context.GetPathInfo(".")
	assert.Regexp(t, fmt.Sprintf("^Expected . to be a subdirectory of .gflows"), err)
}
