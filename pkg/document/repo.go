package document

import (
	"errors"
	"fmt"
	"path/filepath"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
	"gopkg.in/src-d/go-git.v4/plumbing/cache"
	"gopkg.in/src-d/go-git.v4/storage"
/* 	
	"gopkg.in/src-d/go-git.v4/storage/filesystem" */
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/osfs"
)

const (
	SshAuth = "ssh-key"
	SshPass = "ssh-pass"
	HttpBasic = "http-basic"
	NoneTypeAuth = "None"

	DefaultRemoteName = "origin"
)

type Repository struct {

	Filesystem billy.Filesystem
	StoredRepo *git.Repository

	Spec *RepositorySpec `json:"spec,omitempty"`
}

type RepositorySpec struct {

	// URL for Repositor,
	UrlString string `json:"url"`

	// Username is the username for authentication to the repository .
	// +optional
	Username string `json:"username,omitempty"`

	// Clone To Name  Should always be relative to the setting of Manifest TargetPath.
	// Defines where ths repo will be cloned to locally.
	TargetPath string `json:"target-path"`

	CommitHash string `json:"commitHash,omitempty"`
	
	Auth Auth `json:"auth"`
	
}

// Auth struct describies method of authentication agaist given repository
type Auth struct {

	// Type of the authentication method to be used with given repository
	// supported types are "ssh-key", "ssh-pass", "http-basic", "None"
	Type string `json:"type,omitempty"`
	KeyPassword string `json:"keyPass,omitempty"`
	KeyPath string `json:"sshKey,omitempty"`
	HttpPassword string `json:"httpPassword,omitempty"`
	SshPassword string `json:"sshPassword,omitempty"`
	Username string `json:username,omitempty`
}

func NewRepositoryFromSpec(basePath string, spec *RepositorySpec) *Repository {
	return &Repository {
		Filesystem: osfs.New(filepath.Join(basePath, spec.TargetPath)), 
		Spec: spec,
	}
}

func storerFromFs(fs billy.Filesystem) (storage.Storer, error){
	dot, err := fs.Chroot(".git")
	if err != nil {
		return nil, err
	}
	return filesystem.NewStorage(dot, cache.NewObjectLRUDefault()), nil
}

// Clone given repository, and remoteName argument will be used to name remote 
func (repo *Repository) Clone(remoteName string) (error) {

	s, err := storerFromFs(repo.Filesystem)
	if err != nil {
		return err
	}

	if remoteName == "" {
		remoteName = "origin"
	}
	auth, err:= repo.deriveAuth()
	if err != nil {
		return err
	}

	cloneOpts := &git.CloneOptions{ 
		Auth: auth,
		URL: repo.Spec.UrlString,
		RemoteName: remoteName,
	}

	repo.StoredRepo, err = git.Clone(s, repo.Filesystem, cloneOpts)
	return err
}

/* func (repo *Repository) Validate() error {
	allowedAuthTypes := []string{SshAuth,SshPass,HttpBasic}

	auth := repo.Spec.Auth
	if ! stringInSlice(auth.Type, allowedAuthTypes) {
		return errors.New("Invalid auth type, must be one of " + strings.Join(allowedAuthTypes, ","))
	}
	u, err := url.Parse(repo.Spec.UrlString)
	if err != nil {
		return err
	}
	switch u.Scheme {
	case "ssh": 
		if stringInSlice(auth.Type, []string{HttpBasic}){
			return errors.New("If url scheme is ssh, auth type can not be " + auth.Type)
		}
	case "https", "http":
		if stringInSlice(auth.Type, []string{SshAuth,SshPass}){
			return errors.New("If url scheme is http, auth type can not be " + auth.Type)
		}
	}

	return nil

} */

func stringInSlice(a string, list []string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}


func (repo *Repository) deriveAuth() (transport.AuthMethod, error) {

	switch repo.Spec.Auth.Type {
	case SshAuth: 
		return ssh.NewPublicKeysFromFile(repo.Spec.Auth.Username, repo.Spec.Auth.KeyPath, repo.Spec.Auth.KeyPassword)
	case SshPass:
		return &ssh.Password{User: repo.Spec.Auth.Username, Password: repo.Spec.Auth.HttpPassword}, nil
	case HttpBasic:
		return &http.BasicAuth{Username: repo.Spec.Auth.Username, Password: repo.Spec.Auth.HttpPassword}, nil
	case NoneTypeAuth:
		return nil, nil
	default:
		return nil, errors.New("Type not implemented: " + repo.Spec.Auth.Type)
	}
}


func (repo *Repository) Update() error {
	return nil
}

func (repo *Repository) Checkout(enforce bool) error {

	tree, err := repo.StoredRepo.Worktree()
	if err != nil {
		fmt.Printf("we got an error during worktree fetching: %v\n", err)
	}

	opts := &git.CheckoutOptions{
		Hash: plumbing.NewHash(repo.Spec.CommitHash),
		Force: enforce,
	}

	return tree.Checkout(opts)
}

func (repo *Repository) Open() error {
	s, err := storerFromFs(repo.Filesystem)
	if err != nil {
		return err
	}

	repo.StoredRepo, err = git.Open(s, repo.Filesystem)
	
	return err
}

func (repo *Repository) Download(enforce bool) error {
	if repo.StoredRepo == nil {
		err := repo.Clone(DefaultRemoteName)
		if err == git.ErrRepositoryAlreadyExists {
			fmt.Printf("Repository already exists at given path %s remote url %s\n", repo.Filesystem.Root(), repo.Spec.UrlString)
			openErr := repo.Open()
			if openErr != nil {
				fmt.Printf("Error during opening git repository, remote url %s, error: %v\n", repo.Spec.UrlString, err)
				return err
			}
		} else if err != nil {
			fmt.Printf("Error during cloning git repository, remote url %s, error: %v\n", repo.Spec.UrlString, err)
			return err
		}
	}
	fmt.Printf("Error during cloning git repository, remote url %s, error: %v\n", repo.Spec.UrlString)
	return repo.Checkout(enforce)
	

}