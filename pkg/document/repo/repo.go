package repo

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/osfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/cache"
	"gopkg.in/src-d/go-git.v4/storage"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"

	"opendev.org/airship/airshipctl/pkg/log"
)

const (
	SSHAuth   = "ssh-key"
	SSHPass   = "ssh-pass"
	HTTPBasic = "http-basic"

	DefaultRemoteName = "origin"
)

var (
	ErrNoOpenRepo              = errors.New("No open repository is stored")
	ErrRemoteRefNotImplemented = errors.New("RemoteRef is not yet impletemented")
)

// Repository container holds Filesystem, spec and open repository object
// Abstracts git repository and allows for easy cloning, checkout and update of git repos
type Repository struct {
	Driver Adapter
	OptionsBuilder
	Name string
}

// NewRepositoryFromSpec create repository object, with real filesystem on disk
// basePath is used to calculate final path where to clone/open the repository
func NewRepositoryFromSpec(basePath string, spec *RepositorySpec) (*Repository, error) {
	err := spec.Validate()
	if err != nil {
		return nil, err
	}

	if spec.RemoteName == "" {
		spec.RemoteName = DefaultRemoteName
	}

	if spec.Checkout == nil {
		spec.Checkout = &Checkout{Branch: "master"}
	}
	dirName := nameFromURL(spec.URLString)
	fs := osfs.New(filepath.Join(basePath, dirName))

	s, err := storerFromFs(fs)
	if err != nil {
		return nil, err
	}

	// This can create
	return &Repository{
		Name:           dirName,
		Driver:         NewGitDriver(fs, s),
		OptionsBuilder: NewBuilder(spec),
	}, nil
}

func nameFromURL(urlString string) string {
	_, fileName := filepath.Split(urlString)
	return strings.TrimSuffix(fileName, ".git")
}

func storerFromFs(fs billy.Filesystem) (storage.Storer, error) {
	dot, err := fs.Chroot(".git")
	if err != nil {
		return nil, err
	}
	return filesystem.NewStorage(dot, cache.NewObjectLRUDefault()), nil
}

// Update fetches new refs, and checkout according to checkout options
func (repo *Repository) Update(force bool) error {
	log.Debugf("Updating repository %s", repo.Name)
	if !repo.Driver.IsOpen() {
		return ErrNoOpenRepo
	}
	auth, err := repo.ToAuth()
	if err != nil {
		return err
	}
	err = repo.Driver.Fetch(repo.ToFetchOptions(auth))
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("Failed to fetch refs for repository %v: %w", repo.Name, err)
	}
	return repo.Checkout(force)
}

func (repo *Repository) Checkout(enforce bool) error {
	log.Debugf("Attempting to checkout the repository %s", repo.Name)
	if !repo.Driver.IsOpen() {
		return ErrNoOpenRepo
	}
	co := repo.ToCheckoutOptions(enforce)
	tree, err := repo.Driver.Worktree()
	if err != nil {
		return fmt.Errorf("Cloud not get worktree from the repo, %w", err)
	}
	return tree.Checkout(co)
}

func (repo *Repository) Open() error {
	log.Debugf("Attempting to open repository %s", repo.Name)
	return repo.Driver.Open()
}

// Clone given repository
func (repo *Repository) Clone() error {
	log.Debugf("Attempting to clone the repository %s", repo.Name)
	auth, err := repo.ToAuth()
	if err != nil {
		return fmt.Errorf("Failed to build Auth options for repository %v: %w", repo.Name, err)
	}

	return repo.Driver.Clone(repo.ToCloneOptions(auth))
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// Download will clone and checkout repository based on auth and checkout fields of the Repository object
// If repository is already cloned, it will be opened and checked out to configured hash,branch,tag etc...
// no remotes will be modified in this case, also no refs will be updated.
// enforce parameter is used to simulate git reset --hard option.
// If you want to enforce state of the repository, please delete current git repository before downloading.
func (repo *Repository) Download(enforceCheckout bool) error {
	log.Debugf("Attempting to download the repository %s", repo.Name)

	if !repo.Driver.IsOpen() {
		err := repo.Clone()
		if err == git.ErrRepositoryAlreadyExists {
			openErr := repo.Open()
			if openErr != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	return repo.Checkout(enforceCheckout)
}

// RepositorySpec holds options how to authenticate, add remote and checkout repository
type RepositorySpec struct {
	// URLString for Repository,
	URLString string `json:"url"`
	// Auth holds authentication options against remote
	Auth *Auth `json:"auth,omitempty"`
	// Checkout Holds options to checkout repository
	Checkout *Checkout `json:"checkout,omitempty"`
	// RemoteName is a remote that will be added with url, and used to checkout
	RemoteName string `json:"remote-name,omitempty"`
}

var (
	ErrRepoSpecRequiresURL = errors.New("Repostory spec requires url")
)

func (spec *RepositorySpec) Validate() error {
	if spec.URLString == "" {
		return ErrRepoSpecRequiresURL
	}

	if spec.Auth != nil {
		err := spec.Auth.Validate()
		if err != nil {
			return err
		}
	}

	if spec.Checkout != nil {
		err := spec.Checkout.Validate()
		if err != nil {
			return err
		}
	}

	return nil
}

// Equal compares repository specs
func (repo *RepositorySpec) Equal(s *RepositorySpec) bool {
	if s == nil {
		return s == repo
	}

	return repo.URLString == s.URLString &&
		repo.RemoteName == s.RemoteName &&
		reflect.DeepEqual(s.Auth, repo.Auth) &&
		reflect.DeepEqual(s.Checkout, repo.Checkout)
}

// Checkout container holds information how to checkout repository
// Each field is mutually exclusive
type Checkout struct {
	// CommitHash is full hash of the commit that will be used to checkout
	CommitHash string `json:"commit-hash,omitempty"`
	// Branch is the branch name to checkout
	Branch string `json:"branch"`
	// Tag is the tag name to checkout
	Tag string `json:"tag"`
	// RemoteRef is not supported currently TODO
	// RemoteRef is used for remote checkouts such as gerrit change requests/github pull request
	// for example refs/changes/04/691202/5
	// TODO Add support for fetching remote refs
	RemoteRef string `json:"remote-ref"`
}

var (
	ErrMutuallyExclusiveCheckout = errors.New("Checkout is mutually execlusive, use either: commit-hash, branch, tag")
)

func (c *Checkout) Equal(s *Checkout) bool {
	if s == nil {
		return s == c
	}
	return c.CommitHash == s.CommitHash &&
		c.Branch == s.Branch &&
		c.Tag == s.Tag &&
		c.RemoteRef == s.RemoteRef
}

func (c *Checkout) Validate() error {
	possibleValues := []string{c.CommitHash, c.Branch, c.Tag, c.RemoteRef}
	var r []string
	for _, val := range possibleValues {
		if val != "" {
			r = append(r, val)
		}
	}
	if len(r) > 1 {
		return ErrMutuallyExclusiveCheckout
	}
	if c.RemoteRef != "" {
		return errors.New("RemoteRef is not yet impletemented")
	}
	return nil
}

// Auth struct describies method of authentication agaist given repository
type Auth struct {
	// Type of the authentication method to be used with given repository
	// supported types are "ssh-key", "ssh-pass", "http-basic"
	Type string `json:"type,omitempty"`
	//KeyPassword is a password decrypt ssh private key (used with ssh-key auth type)
	KeyPassword string `json:"key-pass,omitempty"`
	// KeyPath is path to private ssh key on disk (used with ssh-key auth type)
	KeyPath string `json:"ssh-key,omitempty"`
	//HTTPPassword is password for basic http authentication (used with http-basic auth type)
	HTTPPassword string `json:"http-pass,omitempty"`
	// SSHPassword is password for ssh password authnetication (used with ssh-pass)
	SSHPassword string `json:"ssh-pass,omitempty"`
	// Username to authenticate against git remote (used with any type)
	Username string `json:"username,omitempty"`
}

var (
	AllowedAuthTypes                  = []string{SSHAuth, SSHPass, HTTPBasic}
	ErrAuthTypeNotSupported           = errors.New("Invalid auth, allowed types: " + strings.Join(AllowedAuthTypes, ","))
	ErrMutuallyExclusiveAuthSSHKey    = errors.New("Can not use http-pass, ssh-pass with auth ssh-key")
	ErrMutuallyExclusiveAuthHTTPBasic = errors.New("Can not use ssh-pass, key-path or key-password with http-basic auth")
	ErrMutuallyExclusiveAuthSSHPass   = errors.New("Can not use http-pass, ssh-key,key-pass with auth ssh-pass")
)

func (auth *Auth) Equal(s *Auth) bool {
	if s == nil {
		return s == auth
	}
	return auth.Type == s.Type &&
		auth.KeyPassword == s.KeyPassword &&
		auth.KeyPath == s.KeyPath &&
		auth.SSHPassword == s.SSHPassword &&
		auth.Username == s.Username
}

func (auth *Auth) Validate() error {

	if !stringInSlice(auth.Type, AllowedAuthTypes) {
		return ErrAuthTypeNotSupported
	}

	switch auth.Type {
	case SSHAuth:
		if auth.HTTPPassword != "" || auth.SSHPassword != "" {
			return ErrMutuallyExclusiveAuthSSHKey
		}
	case HTTPBasic:
		if auth.SSHPassword != "" || auth.KeyPath != "" || auth.KeyPassword != "" {
			return ErrMutuallyExclusiveAuthHTTPBasic
		}
	case SSHPass:
		if auth.KeyPath != "" || auth.KeyPassword != "" || auth.HTTPPassword != "" {
			return ErrMutuallyExclusiveAuthSSHPass
		}
	}
	return nil
}
