package git

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/storage"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

var (
	// ErrDirtyBranch is returned if the branch is dirty.
	ErrDirtyBranch = errors.New("dirty branch")

	// ErrInvalidBranch is returned if the branch does not exist.
	ErrInvalidBranch = errors.New("invalid branch")
)

// New returns a new instance of Git object.
func New(opts ...Option) (Client, error) {
	g := &Git{
		// use mem fs by default, this could be overridden by a new functional parameter
		fs:      memfs.New(),
		storage: memory.NewStorage(),
	}

	for _, opt := range opts {
		if opt != nil {
			if err := opt(g); err != nil {
				return nil, errors.Wrap(err, "invalid option used")
			}
		}
	}

	return g, nil
}

// Git represents an abstraction over git.
type Git struct {
	auth    transport.AuthMethod
	fs      billy.Filesystem
	storage storage.Storer

	repository *git.Repository
	worktree   *git.Worktree

	user  string
	email string
}

// NewFile creates a new file in git workspace
func (g *Git) NewFile(name string, body []byte) error {
	fh, err := g.fs.Create(name)
	if err != nil {
		return errors.Wrapf(err, "unable to create a new file %s", name)
	}
	defer fh.Close()

	_, err = fh.Write(body)
	if err != nil {
		return errors.Wrapf(err, "unable to write body to a file %s", name)
	}

	return nil
}

// ReadFile reads a file from git workspace
func (g *Git) ReadFile(path string) ([]byte, error) {
	f, err := g.fs.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to open a file %s", path)
	}

	body, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read a file %s", path)
	}

	return body, nil
}

// Clone clones repository from url.
func (g *Git) Clone(url string, progress io.Writer) (err error) {

	g.repository, err = git.Clone(g.storage, g.fs, &git.CloneOptions{
		Auth:     g.auth,
		URL:      url,
		Progress: progress,
	})
	if err != nil {
		return errors.Wrapf(err, "unable to clone directory %s", url)
	}

	g.worktree, err = g.repository.Worktree()
	if err != nil {
		return errors.Wrapf(err, "unable to get a worktree for URL %s", url)
	}

	return nil
}

// CreateBranch creates a new git branch.
func (g *Git) CreateBranch(name string) error {
	headRef, err := g.repository.Head()
	if err != nil {
		return errors.Wrapf(err, "unable to get a headRef while creating a new branch %s", name)
	}

	refName := plumbing.ReferenceName(name)
	ref := plumbing.NewHashReference(refName, headRef.Hash())
	err = g.repository.Storer.SetReference(ref)
	if err != nil {
		return errors.Wrapf(err, "unable to set reference while create a new branch %s", name)
	}

	return nil
}

// DiffCommits commits check the difference between 2 commits.
// Optionally files can be passed to this function and it will check if the difference was found between commits
// for specified files. similar to "git diff commitA commitB foo/bar.txt"
// The function returns 3 values: commitAreTheSame bool, patch string, err error
func (g *Git) DiffCommits(commitAHash, commitBHash string, files ...string) (bool, string, error) {
	commitA, err := g.repository.CommitObject(plumbing.NewHash(commitAHash))
	if err != nil {
		return false, "", errors.Wrapf(err, "unable to build a commit object from hash %s", commitAHash)
	}

	commitB, err := g.repository.CommitObject(plumbing.NewHash(commitBHash))
	if err != nil {
		return false, "", errors.Wrapf(err, "unable to build a commit object from hash %s", commitBHash)
	}

	patch, err := commitA.Patch(commitB)
	if err != nil {
		return false, "", errors.Wrapf(err, "unable to make a new patch between commits %s and %s", commitAHash, commitBHash)
	}

	// if the patch is empty, there are no changes between commits, we can exit
	if patch.String() == "" {
		return false, "", nil
	}

	// look for specified files in this change, if the files are not present in this change
	// we consider there was no changes for the given files
	filesFoundInChange, err := g.diffCommitsFilesModified(patch, files...)
	return filesFoundInChange, patch.String(), err
}

func (g *Git) diffCommitsFilesModified(patch *object.Patch, files ...string) (bool, error) {
	// if user did not pass any files, we consider the whole patch between commits
	// meaning we return true because there is a difference between commits, otherwise we should've
	// exited early.
	if len(files) == 0 {
		return true, nil
	}

	// if the user passed the files, we will look for them in this patch
	for _, file := range files {
		_, err := g.fs.Stat(file)
		if err != nil {
			return false, err
		}

		for _, diffPatch := range patch.FilePatches() {
			from, to := diffPatch.Files()
			if (from != nil && from.Path() == file) || (to != nil && to.Path() == file) {
				return true, nil
			}
		}
	}

	return false, nil
}

// Checkout to a new branch.
func (g *Git) Checkout(branch string, create, force bool) error {
	b := plumbing.ReferenceName(branch)
	err := g.worktree.Checkout(&git.CheckoutOptions{
		Create: create,
		Force:  force,
		Branch: b,
	})
	if err != nil {
		return errors.Wrapf(err, "unable to checkout to branch %s", branch)
	}

	return nil
}

// RemoveBranch removes a git branch.
func (g *Git) RemoveBranch(name string) error {
	err := g.PullMaster()
	if err != nil {
		return errors.Wrapf(err, "unable to remove branch %s", name)
	}

	headRef, err := g.repository.Head()
	if err != nil {
		return errors.Wrapf(err, "unable to remove branch %s while getting a headRef", name)
	}

	ref := plumbing.NewHashReference(plumbing.ReferenceName(name), headRef.Hash())
	err = g.repository.Storer.RemoveReference(ref.Name())
	if err != nil {
		return errors.Wrapf(err, "unable to remove branch %s", name)
	}

	return nil
}

func (g *Git) validateBranch(name string) error {
	if !strings.HasPrefix(name, "refs/heads/") {
		return ErrInvalidBranch
	}
	return nil
}

// RemoveRemoteBranch removes the branch from remote git repo.
func (g *Git) RemoveRemoteBranch(name string) error {
	if err := g.validateBranch(name); err != nil {
		return errors.Wrapf(err, "unable to remove remote branch %s", name)
	}

	remote, err := g.repository.Remote("origin")
	if err != nil {
		return errors.Wrapf(err, "unable to remove remote branch %s while getting a remote object: %s", name, err)
	}

	err = remote.Push(&git.PushOptions{
		Auth: g.auth,
		RefSpecs: []config.RefSpec{
			config.RefSpec(":" + name),
		},
		Progress: os.Stdout,
	})

	if err != nil {
		return errors.Wrapf(err, "unable to remove remote branch %s", name)
	}

	return nil
}

// PullMaster pulls the latest changes from master branch.
func (g *Git) PullMaster() error {
	err := g.Checkout("refs/heads/master", false, false)
	if err != nil {
		return errors.Wrap(err, "unable to pull, checkout failed")
	}

	clean, _, err := g.Clean()
	if err != nil {
		return errors.Wrap(err, "unable to pull, clean failed")
	}

	if !clean {
		return ErrDirtyBranch
	}

	err = g.worktree.Pull(&git.PullOptions{
		Auth:       g.auth,
		RemoteName: "origin",
	})

	if err != nil && err != git.NoErrAlreadyUpToDate {
		return errors.Wrap(err, "unable to pull")
	}

	return nil
}

// ReadDir reads the directory and returns a list of FileInfos.
func (g *Git) ReadDir(path string) ([]os.FileInfo, error) {
	return g.fs.ReadDir(path)
}

// OpenFile opens a new file from git workspace. The caller is responsible for closing it.
func (g *Git) OpenFile(path string) (io.ReadCloser, error) {
	return g.fs.Open(path)
}

// Add similar to git add.
func (g *Git) Add(path string) error {
	_, err := g.worktree.Add(path)
	if err != nil {
		return errors.Wrapf(err, "unable to add %s", path)
	}

	return nil
}

// Commit makes a new git commit.
func (g *Git) Commit(msg string) (string, string, error) {
	commit, err := g.worktree.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  g.user,
			Email: g.email,
			When:  time.Now(),
		},
	})
	if err != nil {
		return "", "", errors.Wrap(err, "unable to commit")
	}

	obj, err := g.repository.CommitObject(commit)
	if err != nil {
		return "", "", errors.Wrap(err, "unable to commit object")
	}

	return obj.String(), obj.Hash.String(), nil
}

// Push to remote repo.
func (g *Git) Push(branches ...string) error {
	if len(branches) == 0 {
		return errors.New("empty branch")
	}

	var refSpecs []config.RefSpec
	for _, branch := range branches {
		refSpecs = append(refSpecs, config.RefSpec(fmt.Sprintf("%s:%s", branch, branch)))
	}

	err := g.repository.Push(&git.PushOptions{
		Auth:     g.auth,
		Progress: os.Stdout,
		RefSpecs: refSpecs,
	})
	if err != nil {
		return errors.Wrapf(err, "unable to push to branches %s", branches)
	}

	return nil
}

// Clean returns  true if the branch is clean (e.g. git status)
func (g *Git) Clean() (bool, string, error) {
	s, err := g.worktree.Status()
	if err != nil {
		return false, "", errors.Wrap(err, "unable to check worktree status")
	}

	return s.IsClean(), s.String(), nil
}
