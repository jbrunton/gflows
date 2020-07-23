package config

import (
	"os"
	"path/filepath"

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

func NewContext(fs *afero.Afero, configPath string, enableColors bool) (*GFlowsContext, error) {
	contextDir := filepath.Dir(configPath)

	config, err := LoadConfig(fs, configPath)
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
		ConfigPath:   configPath,
		GitHubDir:    githubDir,
		WorkflowsDir: workflowsDir,
		Dir:          contextDir,
		EnableColors: enableColors,
	}

	return context, nil
}

// GetContext - returns the current command context
func GetContext(fs *afero.Afero, cmd *cobra.Command) (*GFlowsContext, error) {
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

	return NewContext(fs, configPath, !disableColors)
}

// ResolvePath - returns paths relative to the working directory (since paths in configs may be written relative to the
// context directory instead)
func (context *GFlowsContext) ResolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(context.Dir, path)
}

// ResolvePaths - returns an array of resolved paths
func (context *GFlowsContext) ResolvePaths(paths []string) []string {
	return funk.Map(paths, context.ResolvePath).([]string)
}
