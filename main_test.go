// Copyright (c) Curt Bushko.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/google/go-github/v53/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAndValidateState(t *testing.T) {
	cases := []struct {
		name          string
		actual        string
		expectedValue string
		expectError   bool
	}{
		{
			name:          "success",
			actual:        "success",
			expectedValue: "success",
			expectError:   false,
		},
		{
			name:          "fail_with_invalid_state",
			actual:        "foo",
			expectedValue: "",
			expectError:   true,
		},
		{
			name:          "change_cancelled_to_error",
			actual:        "cancelled",
			expectedValue: "error",
			expectError:   false,
		},
		{
			name:          "change_skipped_to_pending",
			actual:        "skipped",
			expectedValue: "error",
			expectError:   false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			value, err := convertActionStateToRepoStatusState(c.actual)
			assert.Equal(t, c.expectedValue, value)
			if err != nil {
				assert.Equal(t, c.expectError, true)
			}
		})
	}
}

func TestRequiredInputs(t *testing.T) {
	cases := []struct {
		name   string
		inputs input
		expErr string
	}{
		{
			name:   "no_inputs_returns_all_required_fields",
			inputs: input{},
			expErr: fmt.Sprintf("%s, %s", tokenRequiredErr, stateRequiredErr),
		},
		{
			name: "token_input_returns_required_fields",
			inputs: input{
				token: "foo",
			},
			expErr: stateRequiredErr,
		},
		{
			name: "token_and_state_inputs_returns_no_errors",
			inputs: input{
				token: "foo",
				state: "bar",
			},
			expErr: "",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validateRequiredInputs(c.inputs)
			if c.expErr != "" {
				require.EqualError(t, err, c.expErr)
			}
		})
	}
}

func TestSetInputDefaults(t *testing.T) {
	cases := []struct {
		name   string
		inputs input
		expErr string
	}{
		{
			name:   "no_inputs_returns_all_required_fields",
			inputs: input{},
			expErr: fmt.Sprintf("%s, %s", tokenRequiredErr, stateRequiredErr),
		},
		{
			name: "token_input_returns_required_fields",
			inputs: input{
				token: "foo",
			},
			expErr: stateRequiredErr,
		},
		{
			name: "token_and_state_inputs_returns_no_errors",
			inputs: input{
				token: "foo",
				state: "bar",
			},
			expErr: "",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validateRequiredInputs(c.inputs)
			if c.expErr != "" {
				require.EqualError(t, err, c.expErr)
			}
		})
	}
}

func TestGetOwnerEnvironmentVariable(t *testing.T) {
	cases := []struct {
		name        string
		actual      string
		owner       string
		expectedErr error
	}{
		{
			name:        "set_owner",
			actual:      "foo",
			owner:       "foo",
			expectedErr: nil,
		},
		{
			name:        "owner_not_set_in_environment",
			actual:      "",
			owner:       "",
			expectedErr: errors.New(ownerEnvNotSetErr),
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Always set the environment variable because it will persist
			// across cases
			err := os.Setenv("GITHUB_OWNER", c.owner)
			require.NoError(t, err)
			got, err := getOwner()
			require.Equal(t, c.owner, got)
			if c.expectedErr != nil {
				require.Equal(t, c.expectedErr, err)
			}
		})
	}
}

func TestGetRepositoryEnvironmentVariable(t *testing.T) {
	cases := []struct {
		name        string
		actual      string
		repo        string
		expectedErr error
	}{
		{
			name:        "set_repository",
			actual:      "foo",
			repo:        "foo",
			expectedErr: nil,
		},
		{
			name:        "repository_not_set_in_environment",
			actual:      "",
			repo:        "",
			expectedErr: errors.New(repositoryEnvNotSetErr),
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Always set the environment variable because it will persist
			// across cases
			err := os.Setenv("GITHUB_REPOSITORY", c.repo)
			require.NoError(t, err)
			got, err := getRepository()
			require.Equal(t, c.repo, got)
			if c.expectedErr != nil {
				require.Equal(t, c.expectedErr, err)
			}
		})
	}
}

func TestGetSHAEnvironmentVariable(t *testing.T) {
	cases := []struct {
		name        string
		actual      string
		sha         string
		expectedErr error
	}{
		{
			name:        "set_sha",
			actual:      "123456",
			sha:         "123456",
			expectedErr: nil,
		},
		{
			name:        "sha_not_set_in_environment",
			actual:      "",
			sha:         "",
			expectedErr: errors.New(shaEnvNotSetErr),
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Always set the environment variable because it will persist
			// across cases
			err := os.Setenv("GITHUB_SHA", c.sha)
			require.NoError(t, err)
			got, err := getSHA()
			require.Equal(t, c.sha, got)
			if c.expectedErr != nil {
				require.Equal(t, c.expectedErr, err)
			}
		})
	}
}

func TestRemoveOwnerFromRepo(t *testing.T) {
	cases := []struct {
		name     string
		repo     string
		owner    string
		expected string
	}{
		{
			name:     "owner is in repo",
			repo:     "foo/bar",
			owner:    "foo",
			expected: "bar",
		},
		{
			name:     "owner is not in repo",
			repo:     "bar",
			owner:    "foo",
			expected: "bar",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := removeOwnerFromRepository(c.repo, c.owner)
			require.Equal(t, c.expected, got)
		})
	}
}

func TestGetInputs(t *testing.T) {
	cases := []struct {
		name          string
		inputs        input
		inputEnvOwner string
		inputEnvRepo  string
		inputEnvSHA   string
		expected      input
		expectError   string
	}{
		{
			name: "all_inputs_set",
			inputs: input{
				token:       "some-token",
				state:       "success",
				context:     "some-context",
				description: "some-description",
				owner:       "some-owner",
				repository:  "some-owner/some-repo",
				detailsURL:  "some-url",
				sha:         "some-sha",
			},
			expected: input{
				token:       "some-token",
				state:       "success",
				context:     "some-context",
				description: "some-description",
				owner:       "some-owner",
				repository:  "some-repo",
				detailsURL:  "some-url",
				sha:         "some-sha",
			},
		},
		{
			name: "state_converted",
			inputs: input{
				token:       "some-token",
				state:       "cancelled",
				context:     "some-context",
				description: "some-description",
				owner:       "some-owner",
				repository:  "some-repo",
				detailsURL:  "some-url",
				sha:         "some-sha",
			},
			expected: input{
				token:       "some-token",
				state:       "error",
				context:     "some-context",
				description: "some-description",
				owner:       "some-owner",
				repository:  "some-repo",
				detailsURL:  "some-url",
				sha:         "some-sha",
			},
		},
		{
			name: "defaults set",
			inputs: input{
				token:       "some-token",
				state:       "cancelled",
				context:     "some-context",
				description: "some-description",
				owner:       "",
				repository:  "",
				detailsURL:  "some-url",
				sha:         "",
			},
			expected: input{
				token:       "some-token",
				state:       "error",
				context:     "some-context",
				description: "some-description",
				owner:       "env-owner",
				repository:  "env-repo",
				detailsURL:  "some-url",
				sha:         "env-sha",
			},
			inputEnvOwner: "env-owner",
			inputEnvRepo:  "env-repo",
			inputEnvSHA:   "env-sha",
		},
		{
			name: "error_invalid_state",
			inputs: input{
				token:       "some-token",
				state:       "some-state",
				context:     "some-context",
				description: "some-description",
				owner:       "some-owner",
				repository:  "some-repo",
				detailsURL:  "some-url",
				sha:         "some-sha",
			},
			expected:    input{},
			expectError: "value not supported",
		},
		{
			name: "error_missing_token",
			inputs: input{
				token:       "",
				state:       "success",
				context:     "some-context",
				description: "some-description",
				owner:       "some-owner",
				repository:  "some-repo",
				detailsURL:  "some-url",
				sha:         "some-sha",
			},
			expected:    input{},
			expectError: tokenRequiredErr,
		},
		{
			name: "error-missing-env-owner",
			inputs: input{
				token:       "some-token",
				state:       "cancelled",
				context:     "some-context",
				description: "some-description",
				owner:       "",
				repository:  "",
				detailsURL:  "some-url",
				sha:         "",
			},
			expected:      input{},
			inputEnvOwner: "",
			inputEnvRepo:  "env-repo",
			inputEnvSHA:   "env-sha",
			expectError:   ownerEnvNotSetErr,
		},
		{
			name: "error-missing-env-repo",
			inputs: input{
				token:       "some-token",
				state:       "cancelled",
				context:     "some-context",
				description: "some-description",
				owner:       "",
				repository:  "",
				detailsURL:  "some-url",
				sha:         "",
			},
			expected:      input{},
			inputEnvOwner: "env-owner",
			inputEnvRepo:  "",
			inputEnvSHA:   "env-sha",
			expectError:   repositoryEnvNotSetErr,
		},
		{
			name: "error-missing-env-sha",
			inputs: input{
				token:       "some-token",
				state:       "cancelled",
				context:     "some-context",
				description: "some-description",
				owner:       "",
				repository:  "",
				detailsURL:  "some-url",
				sha:         "",
			},
			expected:      input{},
			inputEnvOwner: "env-owner",
			inputEnvRepo:  "env-repo",
			inputEnvSHA:   "",
			expectError:   shaEnvNotSetErr,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := os.Setenv("GITHUB_REPOSITORY", c.inputEnvRepo)
			require.NoError(t, err)
			err = os.Setenv("GITHUB_OWNER", c.inputEnvOwner)
			require.NoError(t, err)
			err = os.Setenv("GITHUB_SHA", c.inputEnvSHA)
			require.NoError(t, err)

			got, err := getInputs(mockGetInput{c.inputs}.GetInput)
			if c.expectError != "" {
				require.Contains(t, err.Error(), c.expectError)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, c.expected, got)
		})
	}
}

func TestCreateStatus(t *testing.T) {
	id := int64(24601)
	cases := []struct {
		name         string
		inputs       input
		ghRepoClient ghRepositoryClient
		expectError  string
	}{
		{
			name: "success",
			inputs: input{
				token:       "some-token",
				state:       "success",
				context:     "some-context",
				description: "some-description",
				owner:       "some-owner",
				repository:  "some-repo",
				detailsURL:  "some-url",
				sha:         "some-sha",
			},
			ghRepoClient: mockghRepositoryClient{
				status:      &github.RepoStatus{ID: &id},
				returnError: false,
				t:           t,
				in: input{
					token:       "some-token",
					state:       "success",
					context:     "some-context",
					description: "some-description",
					owner:       "some-owner",
					repository:  "some-repo",
					detailsURL:  "some-url",
					sha:         "some-sha",
				},
			},
		},
		{
			name: "error-nil-status-id",
			inputs: input{
				token:       "some-token",
				state:       "success",
				context:     "some-context",
				description: "some-description",
				owner:       "some-owner",
				repository:  "some-repo",
				detailsURL:  "some-url",
				sha:         "some-sha",
			},
			ghRepoClient: mockghRepositoryClient{
				status:      &github.RepoStatus{},
				returnError: false,
				t:           t,
				in: input{
					token:       "some-token",
					state:       "success",
					context:     "some-context",
					description: "some-description",
					owner:       "some-owner",
					repository:  "some-repo",
					detailsURL:  "some-url",
					sha:         "some-sha",
				},
			},
			expectError: "status ID",
		},
		{
			name: "error-from-client",
			inputs: input{
				token:       "some-token",
				state:       "success",
				context:     "some-context",
				description: "some-description",
				owner:       "some-owner",
				repository:  "some-repo",
				detailsURL:  "some-url",
				sha:         "some-sha",
			},
			ghRepoClient: mockghRepositoryClient{status: &github.RepoStatus{}, returnError: true, t: t},
			expectError:  "some-error",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx := context.Background()
			gh := ghClient{c.ghRepoClient, c.inputs, uint64(0)}
			err := gh.createStatus(ctx)
			if c.expectError != "" {
				require.Contains(t, err.Error(), c.expectError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

type mockGetInput struct {
	in input
}

func (m mockGetInput) GetInput(i string) string {
	switch i {
	case "token":
		return m.in.token
	case "state":
		return m.in.state
	case "owner":
		return m.in.owner
	case "repository":
		return m.in.repository
	case "sha":
		return m.in.sha
	case "context":
		return m.in.context
	case "description":
		return m.in.description
	case "details_url":
		return m.in.detailsURL
	default:
		return ""
	}
}

type mockghRepositoryClient struct {
	status      *github.RepoStatus
	returnError bool
	t           *testing.T

	in input
}

func (m mockghRepositoryClient) CreateStatus(_ context.Context, owner, repo, ref string, status *github.RepoStatus) (*github.RepoStatus, *github.Response, error) {
	if m.returnError {
		return nil, nil, errors.New("some-error")
	}

	assert.Equal(m.t, m.in.owner, owner)
	assert.Equal(m.t, m.in.repository, repo)
	assert.Equal(m.t, m.in.sha, ref)

	// Verify the Status
	assert.Equal(m.t, m.in.state, *status.State)
	assert.Equal(m.t, m.in.context, *status.Context)
	assert.Equal(m.t, m.in.description, *status.Description)
	assert.Equal(m.t, m.in.detailsURL, *status.TargetURL)

	return m.status, nil, nil
}
