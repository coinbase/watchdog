package git

import (
	"io"
	"os"
)

// Client is an interface which describes a git client.
type Client interface {
	// NewFile creates a new file in git workspace.
	NewFile(name string, body []byte) error

	// ReadFile reads a file from a git workspace.
	ReadFile(path string) ([]byte, error)

	// Clone the repository from url.
	Clone(url string, progress io.Writer) (err error)

	// CreateBranch creates a local branch.
	CreateBranch(name string) error

	// DiffCommits compares 2 commits and returns the difference.
	DiffCommits(commitAHash, commitBHash string, files ...string) (bool, string, error)

	// Checkout to a git branch
	Checkout(branch string, create, force bool) error

	// RemoveBranch removes a local branch.
	RemoveBranch(name string) error

	// RemoveRemoteBranch removes a remote branch.
	RemoveRemoteBranch(name string) error

	// PullMaster pulls and checks out to master branch.
	PullMaster() error

	// ReadDir reads the files/dirs in git workspace.
	ReadDir(path string) ([]os.FileInfo, error)

	// OpenFile opens a file from a git workspace.
	OpenFile(path string) (io.ReadCloser, error)

	// Add works like "git add" to add files to a commit.
	Add(path string) error

	// Commit makes a new commit.
	Commit(msg string) (string, string, error)

	// Push branches to remote repo.
	Push(branches ...string) error

	// Clean returns true if the workspace is clean, similar to "git status"
	Clean() (bool, string, error)
}
