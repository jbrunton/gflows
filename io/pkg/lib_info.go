package pkg

import (
	"github.com/spf13/afero"
)

type LibInfo struct {
	Path     string
	IsRemote bool
	Exists   bool
	IsDir    bool
}

func GetLibInfo(libPath string, fs *afero.Afero) (*LibInfo, error) {
	libInfo := &LibInfo{
		Path:     libPath,
		IsRemote: IsRemotePath(libPath),
	}

	if libInfo.IsRemote {
		libInfo.Exists = true // assume it exists, as we can't know for sure without making a request
		return libInfo, nil
	}

	exists, err := fs.Exists(libPath)
	if err != nil {
		return nil, err
	}
	libInfo.Exists = exists
	if !libInfo.Exists {
		return libInfo, nil
	}

	isDir, err := fs.IsDir(libPath)
	if err != nil {
		return nil, err
	}
	libInfo.IsDir = isDir

	return libInfo, nil
}
