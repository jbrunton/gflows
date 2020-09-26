package env

import (
	"github.com/jbrunton/gflows/io"
	"github.com/jbrunton/gflows/io/content"
	"github.com/spf13/afero"
)

type GFlowsLibInstaller struct {
	fs     *afero.Afero
	reader *content.Reader
	writer *content.Writer
	logger *io.Logger
}

func NewGFlowsLibInstaller(fs *afero.Afero, reader *content.Reader, writer *content.Writer, logger *io.Logger) *GFlowsLibInstaller {
	return &GFlowsLibInstaller{
		fs:     fs,
		reader: reader,
		writer: writer,
		logger: logger,
	}
}

func (installer *GFlowsLibInstaller) install(manifestPath string, installDir string) error {
	manifest, err := installer.loadManifest(manifestPath)
	if err != nil {
		return err
	}

	rootPath, err := content.ParentPath(manifestPath)
	if err != nil {
		return err
	}

	for _, relPath := range manifest.Libs {
		err := installer.copyFile(rootPath, installDir, relPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func (installer *GFlowsLibInstaller) loadManifest(manifestPath string) (*GFlowsLibManifest, error) {
	manifestContent, err := installer.reader.ReadContent(manifestPath)
	if err != nil {
		return nil, err
	}
	return ParseManifest(manifestContent)
}

func (installer *GFlowsLibInstaller) copyFile(rootPath string, installDir string, relPath string) error {
	sourcePath, err := content.JoinRelativePath(rootPath, relPath)
	if err != nil {
		return err
	}
	if content.IsRemotePath(rootPath) {
		installer.logger.Debugf("Downloading %s\n", sourcePath)
	} else {
		installer.logger.Debugf("Copying %s\n", sourcePath)
	}
	destPath, err := content.JoinRelativePath(installDir, relPath)
	if err != nil {
		return err
	}
	sourceContent, err := installer.reader.ReadContent(sourcePath)
	if err != nil {
		return err
	}
	return installer.writer.SafelyWriteFile(destPath, sourceContent)
}
