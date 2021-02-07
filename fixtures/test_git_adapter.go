package fixtures

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
)

type TestGitRepository = map[string]string

type TestGitAdapter struct {
	fs    *afero.Afero
	repos map[string]*TestGitRepository
}

func NewTestGitAdapter(fs *afero.Afero) *TestGitAdapter {
	return &TestGitAdapter{
		fs:    fs,
		repos: make(map[string]*TestGitRepository),
	}
}

func (gitAdapter *TestGitAdapter) Clone(url string, dir string) error {
	repo := gitAdapter.repos[url]
	if repo == nil {
		return fmt.Errorf("Missing repo for %s", url)
	}
	for path, content := range *repo {
		err := gitAdapter.fs.WriteFile(filepath.Join(dir, path), []byte(content), 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func (gitAdapter *TestGitAdapter) StubRepo(url string, repo *TestGitRepository) {
	gitAdapter.repos[url] = repo
}
