package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"

	goio "io"

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
	Libs         map[string]string
}

type LibManifest struct {
	Files []string
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
		Libs:         make(map[string]string),
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

func (context *GFlowsContext) PushGFlowsLib(fs *afero.Afero, libUrl string) (string, error) {
	fmt.Printf("Processing gflowslib: %s\n", libUrl)
	libDir := context.Libs[libUrl]
	if libDir != "" {
		// already processed
		fmt.Println("Already processed lib")
		return libDir, nil
	}

	filename := filepath.Base(libUrl)
	libDir, err := fs.TempDir("", filename)
	fmt.Printf("created tmpdir %s\n", libDir)
	if err != nil {
		panic(err)
	}
	//defer fs.RemoveAll(tmpDir)

	rootUrl, err := url.Parse(libUrl)
	rootUrl.Path = path.Dir(rootUrl.Path)
	fmt.Println("libUrl:", libUrl)
	//fmt.Println("filepath.Dir(libUrl):", filepath.Dir(libUrl))
	fmt.Println("rootUrl:", rootUrl)
	if err != nil {
		return "", err
	}

	manifestPath := filepath.Join(libDir, filename)
	err = DownloadFile(libUrl, manifestPath, fs)
	if err != nil {
		return "", err
	}
	fmt.Printf("Downloaded lib manifest to %s\n", manifestPath)

	manifestContent, err := fs.ReadFile(manifestPath)
	if err != nil {
		return "", err
	}

	manifest := LibManifest{}
	json.Unmarshal(manifestContent, &manifest)

	for _, relPath := range manifest.Files {
		//url := filepath.Join(manifestRoot, relPath)
		url, _ := url.Parse(rootUrl.String())
		if err != nil {
			return "", err
		}
		//fmt.Println("rootUrl:", rootUrl)
		//fmt.Println("relUrl:", relUrl)
		//url := rootUrl.ResolveReference(relUrl)
		url.Path = path.Join(url.Path, relPath)
		fmt.Println("url:", url)
		dest := filepath.Join(libDir, relPath)
		fmt.Printf("Downloading %s to %s\n", url, dest)
		err = DownloadFile(url.String(), dest, fs)
		if err != nil {
			panic(err) // TODO: handle this
		}
		//files = append(files, filepath.Join(tmpDir, file))
		//TODO: download
	}

	context.Libs[libUrl] = libDir

	return libDir, nil
}

// ResolvePaths - returns an array of resolved paths
func (context *GFlowsContext) ResolvePaths(paths []string) []string {
	return funk.Map(paths, context.ResolvePath).([]string)
}

func DownloadFile(url string, path string, fs *afero.Afero) error {
	// Create the file
	dir := filepath.Dir(path)
	if _, err := fs.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			err = fs.MkdirAll(dir, os.ModePerm)
			if err != nil {
				return err
			}
		}
	}

	out, err := os.Create(path)
	if err != nil {
		fmt.Println("Error creating file")
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = goio.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	fmt.Println("Downloaded", url, "to", path)

	return nil
}
