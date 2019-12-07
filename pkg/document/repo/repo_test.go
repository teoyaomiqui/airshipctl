package repo

import (
	"errors"
	"testing"

	"sigs.k8s.io/yaml"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/src-d/go-billy.v4/memfs"
	fixtures "gopkg.in/src-d/go-git-fixtures.v3"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

const (
	validateTestName         = "ToCheckout"
	validateFailuresTestName = "Validate"
	toAuthTestName           = "ToAuth"
	toAuthNilTestName        = "ToAuthNil"
)

var (
	ErrTest        = errors.New("my error")
	StringTestData = `test-data:
  no-auth:
    url: https://github.com/src-d/go-git.git
    checkout:
      tag: v3.0.0
  ssh-key-auth:
    url: git@github.com:src-d/go-git.git
    auth:
      type: ssh-key
      ssh-key: "testdata/test-key.pem"
      username: git
    checkout:
      branch: master
  http-basic-auth:
    url: /home/ubuntu/some-gitrepo
    auth:
      type: http-basic
      http-pass: "qwerty123"
      username: deployer
    checkout:
      commit-hash: 01c4f7f32beb9851ae8f119a6b8e497d2b1e2bb8
  empty-checkout:
    url: /home/ubuntu/some-gitrepo
    auth:
      type: http-basic
      http-pass: "qwerty123"
      username: deployer
  wrong-type-auth:
    url: /home/ubuntu/some-gitrepo
    auth:
      type: wrong-type
      http-pass: "qwerty123"
      username: deployer
    checkout:
      commit-hash: 01c4f7f32beb9851ae8f119a6b8e497d2b1e2bb8
  mutually-exclusive-auth-opts:
    url: /home/ubuntu/some-gitrepo
    auth:
      type: http-basic
      ssh-key: "/path-to-key"
      username: deployer
  mutually-exclusive-checkout-opts:
    url: /home/ubuntu/some-gitrepo
    checkout:
      commit-hash: 01c4f7f32beb9851ae8f119a6b8e497d2b1e2bb8
      branch: master	 
  mutually-exclusive-auth-opts-ssh-key:
    url: /home/ubuntu/some-gitrepo
    auth:
      type: ssh-key
      http-pass: "qwerty123"
      ssh-key: "/path-to-key"
      username: deployer
    checkout:
      commit-hash: 01c4f7f32beb9851ae8f119a6b8e497d2b1e2bb8
  mutually-exclusive-auth-opts-ssh-pass:
    url: /home/ubuntu/some-gitrepo
    auth:
      type: ssh-pass
      ssh-pass: "qwerty123"
      http-pass: "qwerty123"
      ssh-key: "/path-to-key"
      username: deployer
    checkout:
      commit-hash: 01c4f7f32beb9851ae8f119a6b8e497d2b1e2bb8`

	TestCaseMap = map[string]*TestCase{
		validateTestName: {
			expectError:  false,
			dataMapEntry: []string{"http-basic-auth", "ssh-key-auth", "no-auth", "empty-checkout"},
			expectedNil:  false,
		},
		validateFailuresTestName: {
			expectError: true,
			dataMapEntry: []string{"wrong-type-auth",
				"mutually-exclusive-auth-opts",
				"mutually-exclusive-checkout-opts",
				"mutually-exclusive-auth-opts-ssh-key",
				"mutually-exclusive-auth-opts-ssh-pass"},
			expectedNil: true,
		},
		toAuthTestName: {
			expectError:  false,
			dataMapEntry: []string{"ssh-key-auth", "http-basic-auth"},

			expectedNil: false,
		},
		toAuthNilTestName: {
			expectError:  false,
			dataMapEntry: []string{"no-auth"},
			expectedNil:  true,
		},
	}
)

type TestCase struct {
	expectError bool
	// this maps to TestData map in TestRepos struct
	dataMapEntry []string
	expectedNil  bool
}

type TestRepos struct {
	TestData map[string]*RepositorySpec `json:"test-data"`
}

func TestDownload(t *testing.T) {

	err := fixtures.Init()
	require.NoError(t, err)
	fx := fixtures.Basic().One()
	spec := &RepositorySpec{
		Checkout: &Checkout{
			Branch: "master",
		},
		URLString: fx.DotGit().Root(),
	}

	fs := memfs.New()
	s := memory.NewStorage()

	repo, err := NewRepositoryFromSpec(".", spec)
	require.NoError(t, err)
	repo.Driver.SetFilesystem(fs)
	repo.Driver.SetStorer(s)

	err = repo.Download(false)
	assert.NoError(t, err)

	// This should try to open the repo because it is already downloaded
	repoOpen, err := NewRepositoryFromSpec(".", spec)
	require.NoError(t, err)
	repoOpen.Driver.SetFilesystem(fs)
	repoOpen.Driver.SetStorer(s)
	err = repoOpen.Download(false)
	assert.NoError(t, err)
	ref, err := repo.Driver.Head()
	require.NoError(t, err)
	assert.NotNil(t, ref.String())
}

func TestToCheckout(t *testing.T) {
	data := &TestRepos{}
	err := yaml.Unmarshal([]byte(StringTestData), data)
	require.NoError(t, err)

	testCase := TestCaseMap[validateTestName]

	for _, name := range testCase.dataMapEntry {
		t.Logf("Testing Data Entry %s \n", name)
		repo, err := NewRepositoryFromSpec(".", data.TestData[name])
		require.NoError(t, err)
		require.NotNil(t, repo)
		co := repo.ToCheckoutOptions(false)
		if testCase.expectedNil {
			assert.Nil(t, co)
		} else {
			assert.NotNil(t, co)
			assert.NoError(t, co.Validate())
		}
	}
}

func TestToAuth(t *testing.T) {
	data := &TestRepos{}
	err := yaml.Unmarshal([]byte(StringTestData), data)
	require.NoError(t, err)

	for _, testCaseName := range []string{toAuthTestName, toAuthNilTestName} {
		testCase := TestCaseMap[testCaseName]
		for _, name := range testCase.dataMapEntry {
			t.Logf("Testing Data Entry %s \n", name)
			repo, err := NewRepositoryFromSpec(".", data.TestData[name])
			require.NoError(t, err)
			auth, authErr := repo.ToAuth()
			if testCase.expectError {
				assert.Error(t, authErr)
			} else {
				assert.NoError(t, authErr)
			}
			if testCase.expectedNil {
				assert.Nil(t, auth)
			} else {
				assert.NotNil(t, auth)
			}
		}
	}
}

func TestNewRepositoryFromSpec(t *testing.T) {
	data := &TestRepos{}
	err := yaml.Unmarshal([]byte(StringTestData), data)
	require.NoError(t, err)

	for _, testCaseName := range []string{validateTestName, validateFailuresTestName} {
		testCase := TestCaseMap[testCaseName]
		for _, name := range testCase.dataMapEntry {
			t.Logf("Testing Data Entry %s \n", name)
			repo, err := NewRepositoryFromSpec(".", data.TestData[name])
			if testCase.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if testCase.expectedNil {
				assert.Nil(t, repo)
			} else {
				assert.NotNil(t, repo)
			}
		}
	}
}

func TestUpdate(t *testing.T) {
	err := fixtures.Init()
	require.NoError(t, err)
	fx := fixtures.Basic().One()

	checkout := &Checkout{Branch: "master"}
	spec := &RepositorySpec{
		Checkout:  checkout,
		URLString: fx.DotGit().Root(),
	}

	repo, err := NewRepositoryFromSpec(".", spec)
	require.NoError(t, err)
	driver := &GitDriver{
		Filesystem: memfs.New(),
		Storer:     memory.NewStorage(),
	}
	// Set inmemory fs instead of real one
	repo.Driver = driver
	require.NoError(t, err)

	// Clone repo into memory fs
	err = repo.Clone()
	require.NoError(t, err)
	// Get hash of the HEAD
	ref, err := repo.Driver.Head()
	require.NoError(t, err)
	headHash := ref.Hash()

	// calculate previous commit hash
	prevCommitHash, err := repo.Driver.ResolveRevision("HEAD~1")
	require.NoError(t, err)
	require.NotEqual(t, prevCommitHash.String(), headHash.String())
	spec.Checkout = &Checkout{CommitHash: prevCommitHash.String()}
	// Checkout previous commit
	err = repo.Checkout(true)
	require.NoError(t, err)

	// Set checkout back to master
	spec.Checkout = checkout
	err = repo.Checkout(true)
	assert.NoError(t, err)
	// update
	err = repo.Update(true)

	currentHash, err := repo.Driver.Head()
	assert.NoError(t, err)
	// Make sure that current has is same as master hash
	assert.Equal(t, headHash.String(), currentHash.Hash().String())

	repo.Driver.Close()
	updateError := repo.Update(true)
	assert.Error(t, updateError)

}

func TestOpen(t *testing.T) {
	err := fixtures.Init()
	require.NoError(t, err)
	fx := fixtures.Basic().One()

	checkout := &Checkout{Branch: "master"}
	spec := &RepositorySpec{
		Checkout:  checkout,
		URLString: fx.DotGit().Root(),
	}

	repo, err := NewRepositoryFromSpec(".", spec)
	require.NoError(t, err)
	driver := &GitDriver{
		Filesystem: memfs.New(),
		Storer:     memory.NewStorage(),
	}
	repo.Driver = driver

	err = repo.Clone()
	assert.NotNil(t, repo.Driver)
	require.NoError(t, err)

	// This should try to open the repo
	repoOpen, err := NewRepositoryFromSpec(".", spec)
	err = repoOpen.Open()
	ref, err := repo.Driver.Head()
	assert.NoError(t, err)
	assert.NotNil(t, ref.String())
}

func TestCheckout(t *testing.T) {
	err := fixtures.Init()
	require.NoError(t, err)
	fx := fixtures.Basic().One()

	checkout := &Checkout{Branch: "master"}
	spec := &RepositorySpec{
		Checkout:  checkout,
		URLString: fx.DotGit().Root(),
	}

	repo, err := NewRepositoryFromSpec(".", spec)
	require.NoError(t, err)
	err = repo.Checkout(true)
	assert.Error(t, err)
}

func TestURLtoName(t *testing.T) {
	tests := []struct {
		input          string
		expectedOutput string
	}{
		{
			input:          "https://github.com/kubernetes/kubectl.git",
			expectedOutput: "kubectl",
		},
		{
			input:          "git@github.com:kubernetes/kubectl.git",
			expectedOutput: "kubectl",
		},
		{
			input:          "https://github.com/kubernetes/kube.somepath.ctl.git",
			expectedOutput: "kube.somepath.ctl",
		},
		{
			input:          "https://github.com/kubernetes/kubectl",
			expectedOutput: "kubectl",
		},
		{
			input:          "git@github.com:kubernetes/kubectl",
			expectedOutput: "kubectl",
		},
		{
			input:          "github.com:kubernetes/kubectl.git",
			expectedOutput: "kubectl",
		},
		{
			input:          "/kubernetes/kubectl.git",
			expectedOutput: "kubectl",
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expectedOutput, nameFromURL(test.input))
	}
}

func TestEqual(t *testing.T) {
	t.Run("spec-equal", func(t *testing.T) {
		testSpec1 := &RepositorySpec{}
		testSpec2 := &RepositorySpec{}
		testSpec2.URLString = "Different"
		assert.True(t, testSpec1.Equal(testSpec1))
		assert.False(t, testSpec1.Equal(testSpec2))
		assert.False(t, testSpec1.Equal(nil))
	})
	t.Run("auth-equal", func(t *testing.T) {
		testSpec1 := &Auth{}
		testSpec2 := &Auth{}
		testSpec2.Type = "ssh-key"
		assert.True(t, testSpec1.Equal(testSpec1))
		assert.False(t, testSpec1.Equal(testSpec2))
		assert.False(t, testSpec1.Equal(nil))
	})
	t.Run("checkout-equal", func(t *testing.T) {
		testSpec1 := &Checkout{}
		testSpec2 := &Checkout{}
		testSpec2.Branch = "Master"
		assert.True(t, testSpec1.Equal(testSpec1))
		assert.False(t, testSpec1.Equal(testSpec2))
		assert.False(t, testSpec1.Equal(nil))
	})
}
