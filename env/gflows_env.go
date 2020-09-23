package env

import (
	"github.com/jbrunton/gflows/config"
	"github.com/jbrunton/gflows/io"
	"github.com/jbrunton/gflows/io/content"
	"github.com/spf13/afero"
)

type GFlowsEnv struct {
	libs       map[string]*GFlowsLib
	fs         *afero.Afero
	downloader *content.Downloader
	context    *config.GFlowsContext
	logger     *io.Logger
}

func NewGFlowsEnv(fs *afero.Afero, downloader *content.Downloader, context *config.GFlowsContext, logger *io.Logger) *GFlowsEnv {
	return &GFlowsEnv{
		libs:       make(map[string]*GFlowsLib),
		fs:         fs,
		downloader: downloader,
		context:    context,
		logger:     logger,
	}
}

func (env *GFlowsEnv) CleanUp() {
	for _, lib := range env.libs {
		lib.CleanUp()
	}
	env.libs = make(map[string]*GFlowsLib)
}

func (env *GFlowsEnv) LoadLib(libUrl string) (string, error) {
	lib := env.libs[libUrl]
	if lib != nil {
		// already processed
		return lib.LocalDir, nil
	}

	lib = NewGFlowsLib(env.fs, env.downloader, env.logger, libUrl, env.context)
	err := lib.Setup()
	if err != nil {
		return "", err
	}

	env.libs[libUrl] = lib
	return lib.LocalDir, nil
}
