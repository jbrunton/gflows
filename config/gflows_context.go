package config

import (
	"fmt"
	"os"
	"path/filepath"

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

	fmt.Println("disableColors:", disableColors)

	if os.Getenv("GFLOWS_DISABLE_COLORS") == "true" {
		disableColors = true
	}

	return NewContext(fs, configPath, !disableColors)
}

func (context *GFlowsContext) EvalJPaths(workflowName string) []string {
	var paths []string
	configJPaths := context.Config.GetTemplateArrayProperty(workflowName, func(config *GFlowsTemplateConfig) []string {
		return config.Jsonnet.JPath
	})

	for _, path := range configJPaths {
		if filepath.IsAbs(path) {
			paths = append(paths, path)
		} else {
			paths = append(paths, filepath.Join(context.Dir, path))
		}
	}

	return paths
}
