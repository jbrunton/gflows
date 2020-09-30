package env

import (
	"github.com/thoas/go-funk"

	"github.com/jbrunton/gflows/config"
	"github.com/jbrunton/gflows/io"
	"github.com/jbrunton/gflows/io/pkg"
	"github.com/spf13/afero"
)

type GFlowsEnv struct {
	libs      map[string]*GFlowsLib
	fs        *afero.Afero
	installer *GFlowsLibInstaller
	context   *config.GFlowsContext
	logger    *io.Logger
}

func NewGFlowsEnv(fs *afero.Afero, installer *GFlowsLibInstaller, context *config.GFlowsContext, logger *io.Logger) *GFlowsEnv {
	return &GFlowsEnv{
		libs:      make(map[string]*GFlowsLib),
		fs:        fs,
		installer: installer,
		context:   context,
		logger:    logger,
	}
}

func (env *GFlowsEnv) LoadLib(libUrl string) (*GFlowsLib, error) {
	lib := env.libs[libUrl]
	if lib != nil {
		// already processed
		return lib, nil
	}

	lib = NewGFlowsLib(env.fs, env.installer, env.logger, libUrl, env.context)
	err := lib.Setup()
	if err != nil {
		return nil, err
	}

	env.libs[libUrl] = lib
	return lib, nil
}

func (env *GFlowsEnv) GetPackages() []pkg.GFlowsPackage {
	libs := funk.Map(funk.Values(env.libs), func(lib *GFlowsLib) pkg.GFlowsPackage {
		return lib
	}).([]pkg.GFlowsPackage)
	return append(libs, env.context)
}

func (env *GFlowsEnv) CleanUp() {
	for _, lib := range env.libs {
		lib.CleanUp()
	}
	env.libs = make(map[string]*GFlowsLib)
}
