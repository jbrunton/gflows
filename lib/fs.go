package lib

import "github.com/spf13/afero"

// CreateOsFs - creates an OS Afero instance
func CreateOsFs() *afero.Afero {
	fs := afero.NewOsFs()
	return &afero.Afero{Fs: fs}
}

// CreateMemFs - creates an in-memory Afero instance for testing
func CreateMemFs() *afero.Afero {
	fs := afero.NewMemMapFs()
	return &afero.Afero{Fs: fs}
}
