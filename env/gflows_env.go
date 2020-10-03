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

func (env *GFlowsEnv) GetPackages() ([]pkg.GFlowsPackage, error) {
	for _, libPath := range env.context.Config.GetAllDependencies() {
		_, err := env.LoadDependency(libPath)
		if err != nil {
			return nil, err
		}
	}
	deps := funk.Map(funk.Values(env.deps), func(dep *GFlowsLib) pkg.GFlowsPackage {
		return dep
	}).([]pkg.GFlowsPackage)
	return append(deps, env.context), nil
}

// GetLibPaths - returns search paths for the given workflow (including libs and local dependency
// directories)
func (env *GFlowsEnv) GetLibPaths(workflowName string) ([]string, error) {
	libPaths := env.context.Config.GetTemplateLibs(workflowName)
	depPaths := env.context.Config.GetTemplateDeps(workflowName)
	for _, depPath := range depPaths {
		dep, err := env.LoadDependency(depPath)
		if err != nil {
			return nil, err
		}
		libPaths = append(libPaths, dep.LibsDir())
	}
	return env.context.ResolvePaths(libPaths), nil
}

func (env *GFlowsEnv) CleanUp() {
	for _, dep := range env.deps {
		dep.CleanUp()
	}
	env.deps = make(map[string]*GFlowsLib)
}
