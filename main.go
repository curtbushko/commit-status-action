// Copyright (c) Curt Bushko.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-multierror"
	actions "github.com/sethvargo/go-githubactions"
)

const tokenRequiredErr = "token is a required field"
const stateRequiredErr = "state is a required field"
const ownerEnvNotSetErr = "GITHUB_OWNER environment variable not set"
const repositoryEnvNotSetErr = "GITHUB_REPOSITORY environment variable not set"
const shaEnvNotSetErr = "GITHUB_SHA environment variable not set"

type input struct {
	token       string
	state       string
	context     string
	description string
	owner       string
	repository  string
	sha         string
	detailsUrl  string
}

func main() {

	in := input{
		token:       actions.GetInput("token"),
		state:       actions.GetInput("state"),
		context:     actions.GetInput("context"),
		description: actions.GetInput("description"),
		owner:       actions.GetInput("owner"),
		repository:  actions.GetInput("repository"),
		sha:         actions.GetInput("sha"),
		detailsUrl:  actions.GetInput("details_url"),
	}

	err := getRequiredInputs(in)
	if err != nil {
		actions.Fatalf(err.Error())
	}

	if in.owner == "" {
		owner, err := getOwner()
		if err != nil {
			actions.Fatalf(err.Error())
		}
		in.owner = owner
	}

	fmt.Println(in)
}

// getRequiredInputs checks the required inputs and returns an error
// if they are not set
func getRequiredInputs(in input) error {
	var err *multierror.Error
	if in.token == "" {
		err = multierror.Append(err, fmt.Errorf(tokenRequiredErr))
	}

	if in.state == "" {
		err = multierror.Append(err, fmt.Errorf(stateRequiredErr))
	}

	return err.ErrorOrNil()
}

// getRepositoryOwner gets github.repository_owner from the Github API
func getOwner() (string, error) {
	owner := os.Getenv("GITHUB_OWNER")
	if owner == "" {
		return "", fmt.Errorf(ownerEnvNotSetErr)
	}
	return owner, nil
}

// getRepositoryOwner gets github.repository_owner from the Github API
func getRepository() (string, error) {
	repo := os.Getenv("GITHUB_REPOSITORY")
	if repo == "" {
		return "", fmt.Errorf(repositoryEnvNotSetErr)
	}
	return repo, nil
}

// getRepositoryOwner gets github.repository_owner from the Github API
func getSHA() (string, error) {
	sha := os.Getenv("GITHUB_SHA")
	if sha == "" {
		return "", fmt.Errorf(shaEnvNotSetErr)
	}
	return sha, nil
}

// getAndValidateState validates that the state is a correct value. If the state is
// 'cancel', 'cancelled'  or 'skipped' then return 'error' as the state.
// 'Cancelled' can be a valid state if a workflow is cancelled.
func getAndValidateState(s string) (string, error) {
	switch s {
	// success, error, failure or pending
	case "error", "failure", "pending", "success":
		return s, nil
	case "cancel", "cancelled", "skipped":
		return "error", nil
	default:
		return "", fmt.Errorf("state value not supported: %s", s)
	}
}
