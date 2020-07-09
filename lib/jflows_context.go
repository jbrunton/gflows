package lib

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// JFlowsContext - current command context
type JFlowsContext struct {
	Dir          string
	ConfigPath   string
	GitHubDir    string
	WorkflowsDir string
	Config       *JFlowsConfig
}

var contextCache map[*cobra.Command]*JFlowsContext

// NewContext - returns a context for the given config
func NewContext(fs *afero.Afero, configPath string) (*JFlowsContext, error) {
	contextDir := filepath.Dir(configPath)

	config, err := GetContextConfig(fs, configPath)
	if err != nil {
		return nil, err
	}

	githubDir := config.Workflows.GitHubDir
	if githubDir == "" {
		githubDir = ".github/"
	}
	if !filepath.IsAbs(githubDir) {
		githubDir = filepath.Join(filepath.Dir(contextDir), githubDir)
	}

	workflowsDir := filepath.Join(contextDir, "/workflows")

	context := &JFlowsContext{
		Config:       config,
		ConfigPath:   configPath,
		GitHubDir:    githubDir,
		WorkflowsDir: workflowsDir,
		Dir:          contextDir,
	}

	return context, nil
}

// GetContext - returns the current command context
func GetContext(fs *afero.Afero, cmd *cobra.Command) (*JFlowsContext, error) {
	context := contextCache[cmd]
	if context != nil {
		return context, nil
	}

	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		panic(err)
	}

	if configPath == "" {
		configPath = os.Getenv("JFLOWS_CONFIG")
	}
	if configPath == "" {
		configPath = ".jflows/config.yml"
	}

	return NewContext(fs, configPath)
}

// LoadServiceManifest - finds and returns the JFlowsService for the given service
func (context *JFlowsContext) LoadServiceManifest(name string) (JFlowsService, error) {
	serviceContext := context.Config.Services[name]

	// TODO: use afero
	data, err := ioutil.ReadFile(serviceContext.Manifest)

	if err != nil {
		return JFlowsService{}, err
	}

	service := JFlowsService{}
	err = yaml.Unmarshal(data, &service)
	if err != nil {
		panic(err)
	}

	return service, nil
}

func init() {
	contextCache = make(map[*cobra.Command]*JFlowsContext)
}
