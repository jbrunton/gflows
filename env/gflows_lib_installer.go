package env

import (
	"path"

	"github.com/jbrunton/gflows/io"
	"github.com/jbrunton/gflows/io/content"
	"github.com/jbrunton/gflows/io/pkg"
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

func (installer *GFlowsLibInstaller) install(lib *GFlowsLib) ([]*LibFileInfo, error) {
	manifest, err := installer.loadManifest(lib.ManifestPath)
	if err != nil {
		return nil, err
	}

	rootPath, err := pkg.ParentPath(lib.ManifestPath)
	if err != nil {
		return nil, err
	}

	files := []*LibFileInfo{}
	for _, relPath := range manifest.Libs {
		fileInfo, err := installer.copyFile(lib, rootPath, relPath)
		if err != nil {
			return nil, err
		}
		files = append(files, fileInfo)
	}
	return files, nil
}

func (installer *GFlowsLibInstaller) loadManifest(manifestPath string) (*GFlowsLibManifest, error) {
	manifestContent, err := installer.reader.ReadContent(manifestPath)
	if err != nil {
		return nil, err
	}
	return ParseManifest(manifestContent)
}

func (installer *GFlowsLibInstaller) copyFile(lib *GFlowsLib, rootPath string, relPath string) (*LibFileInfo, error) {
	sourcePath, err := pkg.JoinRelativePath(rootPath, relPath)
	if err != nil {
		return nil, err
	}
	if pkg.IsRemotePath(rootPath) {
		installer.logger.Debugf("Downloading %s\n", sourcePath)
	} else {
		installer.logger.Debugf("Copying %s\n", sourcePath)
	}
	destPath, err := pkg.JoinRelativePath(lib.LocalDir, relPath)
	if err != nil {
		return nil, err
	}
	sourceContent, err := installer.reader.ReadContent(sourcePath)
	if err != nil {
		return nil, err
	}
	info := &LibFileInfo{
		SourcePath:  sourcePath,
		LocalPath:   destPath,
		Description: path.Join(lib.ManifestName, relPath),
	}
	err = installer.writer.SafelyWriteFile(destPath, sourceContent)
	return info, err
}
