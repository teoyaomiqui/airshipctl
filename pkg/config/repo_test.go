package config

import (
	"errors"
	"testing"

	"sigs.k8s.io/yaml"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	validateTestName         = "ToCheckout"
	validateFailuresTestName = "Validate"
	toAuthTestName           = "ToAuth"
	toAuthNilTestName        = "ToAuthNil"
	ToFetchOptionsTestName   = "ToFetchOptions"
	toAuthNilError			 = "toAuthNilError"
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
  ssh-pass:
    url: /home/ubuntu/some-gitrepo
    auth:
      type: ssh-pass
      ssh-pass: "qwerty123"
      username: deployer
    checkout:
      commit-hash: 01c4f7f32beb9851ae8f119a6b8e497d2b1e2bb8
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
			expectedNil: false,
		},
		toAuthTestName: {
			expectError: false,
			dataMapEntry: []string{"ssh-key-auth",
				"http-basic-auth",
				"ssh-pass"},

			expectedNil: false,
		},
		toAuthNilError: {
			expectError:  true,
			dataMapEntry: []string{"wrong-type-auth"},
			expectedNil:  true,
		},
		toAuthNilTestName: {
			expectError:  false,
			dataMapEntry: []string{"no-auth"},
			expectedNil:  true,
		},
		ToFetchOptionsTestName: {
			expectError:  false,
			dataMapEntry: []string{"no-auth"},
			expectedNil:  false,
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
	TestData map[string]*Repository `json:"test-data"`
}

func TestToCheckout(t *testing.T) {
	data := &TestRepos{}
	err := yaml.Unmarshal([]byte(StringTestData), data)
	require.NoError(t, err)

	testCase := TestCaseMap[validateTestName]

	for _, name := range testCase.dataMapEntry {
		t.Logf("Testing Data Entry %s \n", name)
		repo := data.TestData[name]
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

	for _, testCaseName := range []string{toAuthTestName, toAuthNilTestName,toAuthNilError} {
		testCase := TestCaseMap[testCaseName]
		for _, name := range testCase.dataMapEntry {
			t.Logf("Testing Data Entry %s \n", name)
			repo := data.TestData[name]
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

func TestValidateRepository(t *testing.T) {
	data := &TestRepos{}
	err := yaml.Unmarshal([]byte(StringTestData), data)
	require.NoError(t, err)

	for _, testCaseName := range []string{validateTestName, validateFailuresTestName} {
		testCase := TestCaseMap[testCaseName]
		for _, name := range testCase.dataMapEntry {
			t.Logf("Testing Data Entry %s \n", name)
			repo := data.TestData[name]
			err := repo.Validate()
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

func TestToFetchOptions(t *testing.T) {
	data := &TestRepos{}
	err := yaml.Unmarshal([]byte(StringTestData), data)
	require.NoError(t, err)

	testCase := TestCaseMap[ToFetchOptionsTestName]

	for _, name := range testCase.dataMapEntry {
		t.Logf("Testing Data Entry %s \n", name)
		repo := data.TestData[name]
		require.NotNil(t, repo)
		assert.NotNil(t, repo.ToFetchOptions(nil))
	}
}

func TestToCloneOptions(t *testing.T) {
	data := &TestRepos{}
	err := yaml.Unmarshal([]byte(StringTestData), data)
	require.NoError(t, err)

	testCase := TestCaseMap[ToFetchOptionsTestName]

	for _, name := range testCase.dataMapEntry {
		t.Logf("Testing Data Entry %s \n", name)
		repo := data.TestData[name]
		require.NotNil(t, repo)
		assert.NotNil(t, repo.ToCloneOptions(nil))
	}
}
