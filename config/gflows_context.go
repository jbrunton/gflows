package config

import (
	"os"
	"path/filepath"

	"github.com/jbrunton/gflows/io"
	"github.com/thoas/go-funk"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// GFlowsContext - current command context
type GFlowsContext struct {
	Dir          string
	ConfigPath   string
	GitHubDir    string
	WorkflowsDir string
	Config       *GFlowsConfig
	EnableColors bool
}

type ContextOpts struct {
	ConfigPath     string
	EnableColors   bool
	Engine         string
	AllowNoContext bool
}

func NewContext(fs *afero.Afero, logger *io.Logger, opts ContextOpts) (*GFlowsContext, error) {
	contextDir := filepath.Dir(opts.ConfigPath)

	config, err := LoadConfig(fs, logger, opts)
	if err != nil {
		return nil, err
	}

	githubDir := config.GithubDir
	if githubDir == "" {
		githubDir = ".github/"
	}
	if !filepath.IsAbs(githubDir) {
		githubDir = filepath.Join(filepath.Dir(contextDir), githubDir)
	}

	workflowsDir := filepath.Join(contextDir, "/workflows")

	context := &GFlowsContext{
		Config:       config,
		ConfigPath:   opts.ConfigPath,
		GitHubDir:    githubDir,
		WorkflowsDir: workflowsDir,
		Dir:          contextDir,
		EnableColors: opts.EnableColors,
	}

	return context, nil
}

// CreateContextOpts - creates ContextOpts from flags and environment variables
func CreateContextOpts(cmd *cobra.Command) ContextOpts {
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		panic(err)
	}

	if configPath == "" {
		configPath = os.Getenv("GFLOWS_CONFIG")
	}
	if configPath == "" {
		configPath = ".gflows/config.yml"
	}

	disableColors, err := cmd.Flags().GetBool("disable-colors")
	if err != nil {
		panic(err)
	}

	if os.Getenv("GFLOWS_DISABLE_COLORS") == "true" {
		disableColors = true
	}

	var engine string
	if cmd.Flags().Lookup("engine") != nil {
		engine, err = cmd.Flags().GetString("engine")
		if err != nil {
			panic(err)
		}
	}

	allowNoContext := funk.ContainsString([]string{"init", "version"}, cmd.Name())

	return ContextOpts{
		ConfigPath:     configPath,
		EnableColors:   !disableColors,
		Engine:         engine,
		AllowNoContext: allowNoContext,
	}
}

// ResolvePath - returns paths relative to the working directory (since paths in configs may be written relative to the
// context directory instead)
func (context *GFlowsContext) ResolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	if filepath.HasPrefix(path, "http://") || filepath.HasPrefix(path, "https://") {
		return path
	}
	return filepath.Join(context.Dir, path)
}

// ResolvePaths - returns an array of resolved paths
func (context *GFlowsContext) ResolvePaths(paths []string) []string {
	return funk.Map(paths, context.ResolvePath).([]string)
}
