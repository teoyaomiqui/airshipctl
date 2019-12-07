package repo

import (
	"errors"

	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"gopkg.in/src-d/go-git.v4/storage"
)

// Adapter is abstraction to version control
type Adapter interface {
	Open() error
	Clone(co *git.CloneOptions) error
	Fetch(fo *git.FetchOptions) error
	Worktree() (*git.Worktree, error)
	Head() (*plumbing.Reference, error)
	ResolveRevision(plumbing.Revision) (*plumbing.Hash, error)
	IsOpen() bool
	SetFilesystem(billy.Filesystem)
	SetStorer(s storage.Storer)
	Close()
}

// GitDriver implements repository interface
type GitDriver struct {
	*git.Repository
	Filesystem billy.Filesystem
	Storer     storage.Storer
}

func NewGitDriver(fs billy.Filesystem, s storage.Storer) *GitDriver {
	return &GitDriver{Storer: s, Filesystem: fs}
}

// Open implements repository interface
func (g *GitDriver) Open() error {
	r, err := git.Open(g.Storer, g.Filesystem)
	if err != nil {
		return err
	}
	g.Repository = r
	return nil
}

func (g *GitDriver) IsOpen() bool {
	if g.Repository == nil {
		return false
	}
	return true
}

// Close sets repository to nil, IsOpen() function will return false now
func (g *GitDriver) Close() {
	g.Repository = nil
}

// Clone implements repository interface
func (g *GitDriver) Clone(co *git.CloneOptions) error {
	r, err := git.Clone(g.Storer, g.Filesystem, co)
	if err != nil {
		return err
	}
	g.Repository = r
	return nil
}

func (g *GitDriver) SetFilesystem(fs billy.Filesystem) {
	g.Filesystem = fs
}

func (g *GitDriver) SetStorer(s storage.Storer) {
	g.Storer = s
}

type OptionsBuilder interface {
	ToAuth() (transport.AuthMethod, error)
	ToCloneOptions(auth transport.AuthMethod) *git.CloneOptions
	ToCheckoutOptions(force bool) *git.CheckoutOptions
	ToFetchOptions(auth transport.AuthMethod) *git.FetchOptions
}

type Builder struct {
	*RepositorySpec
}

func NewBuilder(rs *RepositorySpec) *Builder {
	return &Builder{RepositorySpec: rs}
}

func (b *Builder) ToAuth() (transport.AuthMethod, error) {
	if b.Auth == nil {
		return nil, nil
	}
	switch b.Auth.Type {
	case SSHAuth:
		return ssh.NewPublicKeysFromFile(b.Auth.Username, b.Auth.KeyPath, b.Auth.KeyPassword)
	case SSHPass:
		return &ssh.Password{User: b.Auth.Username, Password: b.Auth.HTTPPassword}, nil
	case HTTPBasic:
		return &http.BasicAuth{Username: b.Auth.Username, Password: b.Auth.HTTPPassword}, nil
	default:
		return nil, errors.New("Type not implemented: " + b.Auth.Type)
	}
}

func (b *Builder) ToCheckoutOptions(force bool) *git.CheckoutOptions {
	co := &git.CheckoutOptions{
		Force: force,
	}
	switch {
	case b.Checkout.Branch != "":
		co.Branch = plumbing.NewBranchReferenceName(b.Checkout.Branch)
	case b.Checkout.Tag != "":
		co.Branch = plumbing.NewTagReferenceName(b.Checkout.Tag)
	case b.Checkout.CommitHash != "":
		co.Hash = plumbing.NewHash(b.Checkout.CommitHash)
	}
	return co
}

func (b *Builder) ToCloneOptions(auth transport.AuthMethod) *git.CloneOptions {
	return &git.CloneOptions{
		Auth:       auth,
		URL:        b.URLString,
		RemoteName: b.RemoteName,
	}
}

func (b *Builder) ToFetchOptions(auth transport.AuthMethod) *git.FetchOptions {
	return &git.FetchOptions{Auth: auth}
}
