package content

import (
	"path/filepath"
	"testing"

	"github.com/jbrunton/gflows/fixtures"
	"github.com/stretchr/testify/assert"
)

func TestGetRepo(t *testing.T) {
	container, _, out := fixtures.NewTestContext("")
	fs := container.FileSystem()
	gitAdapter := container.GitAdapter().(*fixtures.TestGitAdapter)
	repoManager := NewRepoManager(gitAdapter, container.FileSystem(), container.Logger())
	gitAdapter.StubRepo("git@example.com:my/repo", &map[string]string{"example.txt": "foo bar"})

	repo, err := repoManager.GetRepo("git@example.com:my/repo")

	expectedOutput := "Cloning git@example.com:my/repo...\n"
	assert.Equal(t, expectedOutput, out.String())
	assert.NoError(t, err)

	assert.Equal(t, repo.Repository, "git@example.com:my/repo")
	exists, _ := fs.DirExists(repo.LocalDir)
	assert.True(t, exists)
	exists, _ = fs.Exists(filepath.Join(repo.LocalDir, "example.txt"))
	assert.True(t, exists)
}

func TestCleanupRepo(t *testing.T) {
	container, _, _ := fixtures.NewTestContext("")
	fs := container.FileSystem()
	gitAdapter := container.GitAdapter().(*fixtures.TestGitAdapter)
	repoManager := NewRepoManager(gitAdapter, container.FileSystem(), container.Logger())
	gitAdapter.StubRepo("git@example.com:my/repo", &map[string]string{})

	repo, err := repoManager.GetRepo("git@example.com:my/repo")
	assert.NoError(t, err)

	exists, _ := fs.DirExists(repo.LocalDir)
	assert.True(t, exists)

	repoManager.CleanUp()
	exists, _ = fs.DirExists(repo.LocalDir)
	assert.False(t, exists)
}
