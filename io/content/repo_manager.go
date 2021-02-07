package content

import (
	"github.com/jbrunton/gflows/io"
	"github.com/spf13/afero"
)

// GitRepo - a git repository
type GitRepo struct {
	Repository string
	LocalDir   string
}

type RepoManager struct {
	gitAdapter io.GitAdapter
	fs         *afero.Afero
	logger     *io.Logger
	repos      map[string]*GitRepo
}

func NewRepoManager(gitAdapter io.GitAdapter, fs *afero.Afero, logger *io.Logger) *RepoManager {
	return &RepoManager{
		gitAdapter: gitAdapter,
		fs:         fs,
		logger:     logger,
		repos:      make(map[string]*GitRepo),
	}
}

func (manager *RepoManager) GetRepo(url string) (*GitRepo, error) {
	repo := manager.repos[url]
	if repo != nil {
		// already processed
		return repo, nil
	}

	manager.logger.Printfln("Cloning %s...", url)

	tempDir, err := manager.fs.TempDir("", "")
	if err != nil {
		return nil, err
	}

	repo = &GitRepo{
		Repository: url,
		LocalDir:   tempDir,
	}
	manager.repos[url] = repo

	err = manager.gitAdapter.Clone(url, tempDir)

	return repo, err
}

func (manager *RepoManager) CleanUp() {
	for _, repo := range manager.repos {
		manager.logger.Debug("Removing temp directory", repo.LocalDir)
		manager.fs.RemoveAll(repo.LocalDir)
	}
	manager.repos = make(map[string]*GitRepo)
}
