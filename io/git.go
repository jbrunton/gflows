package io

import (
	"os"

	"github.com/go-git/go-git/v5"
)

type GitAdapter interface {
	Clone(repo string, dir string) error
}

type GoGitAdapter struct{}

func NewGoGitAdapter() *GoGitAdapter {
	return &GoGitAdapter{}
}

func (adapter *GoGitAdapter) Clone(url string, dir string) error {
	_, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})
	return err
}
