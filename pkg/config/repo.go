package config

import (
	"errors"
	"reflect"
	"strings"
	
	"sigs.k8s.io/yaml"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

const (
	SSHAuth   = "ssh-key"
	SSHPass   = "ssh-pass"
	HTTPBasic = "http-basic"
)

// RepoCheckout methods
var (
	ErrMutuallyExclusiveCheckout = errors.New("Checkout is mutually execlusive, use either: commit-hash, branch, tag")
)

func (c *RepoCheckout) Equal(s *RepoCheckout) bool {
	if s == nil {
		return s == c
	}
	return c.CommitHash == s.CommitHash &&
		c.Branch == s.Branch &&
		c.Tag == s.Tag &&
		c.RemoteRef == s.RemoteRef
}

func (r *RepoCheckout) String() string {
	yaml, err := yaml.Marshal(&r)
	if err != nil {
		return ""
	}
	return string(yaml)
}

func (c *RepoCheckout) Validate() error {
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

// RepoAuth methods
var (
	AllowedAuthTypes                  = []string{SSHAuth, SSHPass, HTTPBasic}
	ErrAuthTypeNotSupported           = errors.New("Invalid auth, allowed types: " + strings.Join(AllowedAuthTypes, ","))
	ErrMutuallyExclusiveAuthSSHKey    = errors.New("Can not use http-pass, ssh-pass with auth ssh-key")
	ErrMutuallyExclusiveAuthHTTPBasic = errors.New("Can not use ssh-pass, key-path or key-password with http-basic auth")
	ErrMutuallyExclusiveAuthSSHPass   = errors.New("Can not use http-pass, ssh-key,key-pass with auth ssh-pass")
)

func (auth *RepoAuth) Equal(s *RepoAuth) bool {
	if s == nil {
		return s == auth
	}
	return auth.Type == s.Type &&
		auth.KeyPassword == s.KeyPassword &&
		auth.KeyPath == s.KeyPath &&
		auth.SSHPassword == s.SSHPassword &&
		auth.Username == s.Username
}

func (r *RepoAuth) String() string {
	yaml, err := yaml.Marshal(&r)
	if err != nil {
		return ""
	}
	return string(yaml)
}

func (auth *RepoAuth) Validate() error {

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

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// Repository functions
var (
	ErrNoOpenRepo              = errors.New("No open repository is stored")
	
ErrRemoteRefNotImplemented = errors.New("RemoteRef is not yet impletemented")
	ErrRepoSpecRequiresURL     = errors.New("Repostory spec requires url")
)
// Repository functions
// Equal compares repository specs
func (repo *Repository) Equal(s *Repository) bool {
	if s == nil {
		return s == repo
	}

	return repo.URLString == s.URLString &&
		reflect.DeepEqual(s.Auth, repo.Auth) &&
		reflect.DeepEqual(s.CheckoutOptions, repo.CheckoutOptions)
}

func (r *Repository) String() string {
	yaml, err := yaml.Marshal(&r)
	if err != nil {
		return ""
	}
	return string(yaml)
}

func (spec *Repository) Validate() error {
	if spec.URLString == "" {
		return ErrRepoSpecRequiresURL
	}

	if spec.Auth != nil {
		err := spec.Auth.Validate()
		if err != nil {
			return err
		}
	}

	if spec.CheckoutOptions != nil {
		err := spec.CheckoutOptions.Validate()
		if err != nil {
			return err
		}
	}

	return nil
}

func (repo *Repository) ToAuth() (transport.AuthMethod, error) {
	if repo.Auth == nil {
		return nil, nil
	}
	switch repo.Auth.Type {
	case SSHAuth:
		return ssh.NewPublicKeysFromFile(repo.Auth.Username, repo.Auth.KeyPath, repo.Auth.KeyPassword)
	case SSHPass:
		return &ssh.Password{User: repo.Auth.Username, Password: repo.Auth.HTTPPassword}, nil
	case HTTPBasic:
		return &http.BasicAuth{Username: repo.Auth.Username, Password: repo.Auth.HTTPPassword}, nil
	default:
		return nil, errors.New("Type not implemented: " + repo.Auth.Type)
	}
}

func (repo *Repository) ToCheckoutOptions(force bool) *git.CheckoutOptions {
	co := &git.CheckoutOptions{
		Force: force,
	}
	switch {
	case repo.CheckoutOptions == nil:
	case repo.CheckoutOptions.Branch != "":
		co.Branch = plumbing.NewBranchReferenceName(repo.CheckoutOptions.Branch)
	case repo.CheckoutOptions.Tag != "":
		co.Branch = plumbing.NewTagReferenceName(repo.CheckoutOptions.Tag)
	case repo.CheckoutOptions.CommitHash != "":
		co.Hash = plumbing.NewHash(repo.CheckoutOptions.CommitHash)
	}
	return co
}

func (repo *Repository) ToCloneOptions(auth transport.AuthMethod) *git.CloneOptions {
	return &git.CloneOptions{
		Auth: auth,
		URL:  repo.URLString,
	}
}

func (repo *Repository) ToFetchOptions(auth transport.AuthMethod) *git.FetchOptions {
	return &git.FetchOptions{Auth: auth}
}
