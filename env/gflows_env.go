package env

import (
	"github.com/thoas/go-funk"

	"github.com/jbrunton/gflows/config"
	"github.com/jbrunton/gflows/io"
	"github.com/jbrunton/gflows/io/pkg"
	"github.com/spf13/afero"
)

type GFlowsEnv struct {
	deps      map[string]*GFlowsLib
	fs        *afero.Afero
	installer *GFlowsLibInstaller
	context   *config.GFlowsContext
	logger    *io.Logger
}

func NewGFlowsEnv(fs *afero.Afero, installer *GFlowsLibInstaller, context *config.GFlowsContext, logger *io.Logger) *GFlowsEnv {
	return &GFlowsEnv{
		deps:      make(map[string]*GFlowsLib),
		fs:        fs,
		installer: installer,
		context:   context,
		logger:    logger,
	}
}

func (env *GFlowsEnv) LoadDependency(path string) (*GFlowsLib, error) {
	lib := env.deps[path]
	if lib != nil {
		// already processed
		return lib, nil
	}

	lib, err := NewGFlowsLib(env.fs, env.installer, env.logger, path, env.context)
	if err != nil {
		return nil, err
	}
	err = lib.Setup()
	if err != nil {
		return nil, err
	}

	env.deps[path] = lib
	return lib, nil
}

// GetAllPackages - Returns all packages used by the configuration (including the local context)
func (env *GFlowsEnv) GetAllPackages() ([]pkg.GFlowsPackage, error) {
	deps, err := env.loadPackages(env.context.Config.GetAllDependencies())
	return append(deps, env.context), err
}

// GetWorkflowPackages - Returns all packages for a named workflow (including the local context)
func (env *GFlowsEnv) GetWorkflowPackages(workflowName string) ([]pkg.GFlowsPackage, error) {
	deps, err := env.loadPackages(env.context.Config.GetTemplateDeps(workflowName))
	return append(deps, env.context), err
}

// GetLibPaths - returns search paths for the given workflow (including libs and local dependency
// directories)
func (env *GFlowsEnv) GetLibPaths(workflowName string) ([]string, error) {
	libPaths := env.context.Config.GetTemplateLibs(workflowName)

	pkgs, err := env.GetWorkflowPackages(workflowName)
	if err != nil {
		return nil, err
	}

	for _, p := range pkgs {
		libPath := p.LibsDir()
		libInfo, err := pkg.GetLibInfo(libPath, env.fs)
		if err != nil {
			return nil, err
		}
		if libInfo.Exists {
			libPaths = append(libPaths, libPath)
		}
	}

	return env.context.ResolvePaths(libPaths), nil
}

func (env *GFlowsEnv) CleanUp() {
	for _, dep := range env.deps {
		dep.CleanUp()
	}
	env.deps = make(map[string]*GFlowsLib)
}

func (env *GFlowsEnv) loadPackages(paths []string) ([]pkg.GFlowsPackage, error) {
	for _, libPath := range paths {
		_, err := env.LoadDependency(libPath)
		if err != nil {
			return nil, err
		}
	}
	deps := funk.Map(funk.Values(env.deps), func(dep *GFlowsLib) pkg.GFlowsPackage {
		return dep
	}).([]pkg.GFlowsPackage)
	return deps, nil
}
