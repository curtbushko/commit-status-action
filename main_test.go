// Copyright (c) Curt Bushko.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"errors"
	"fmt"
	"os"
	"testing"

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
			value, err := getAndValidateState(c.actual)
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
			err := getRequiredInputs(c.inputs)
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
			os.Setenv("GITHUB_OWNER", c.owner)
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
			os.Setenv("GITHUB_REPOSITORY", c.repo)
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
			os.Setenv("GITHUB_SHA", c.sha)
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
