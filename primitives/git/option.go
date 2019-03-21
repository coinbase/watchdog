package git

import (
	"github.com/pkg/errors"
	ssh2 "golang.org/x/crypto/ssh"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

// ErrUninitializedAuth is returned if the auth object was not set.
var ErrUninitializedAuth = errors.New("uninitialized auth")

// Option is a functional parameter.
type Option func(*Git) error

// WithRSAKey is a option to configure git client with RSA key.
func WithRSAKey(user, password string, privateKeyBody []byte) Option {
	return func(g *Git) error {
		sshAuth, err := ssh.NewPublicKeys(user, privateKeyBody, password)
		if err != nil {
			return errors.Wrap(err, "unable to create a new ssh auth object")
		}

		g.auth = sshAuth
		return nil
	}
}

// WithIgnoreKnownHosts is a functional option used to disable ssh known hosts option.
func WithIgnoreKnownHosts(ignore bool) Option {
	return func(g *Git) error {
		if !ignore {
			return nil
		}

		if g.auth == nil {
			return ErrUninitializedAuth
		}

		g.auth.(*ssh.PublicKeys).HostKeyCallback = ssh2.InsecureIgnoreHostKey()
		return nil
	}
}

// WithGitUserEmail configures user and email for git commits.
func WithGitUserEmail(user, email string) Option {
	return func(g *Git) error {
		g.user = user
		g.email = email

		return nil
	}
}
