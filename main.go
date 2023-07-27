// Copyright (c) Curt Bushko.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v53/github"
	"github.com/hashicorp/go-multierror"
	actions "github.com/sethvargo/go-githubactions"
	"github.com/sethvargo/go-retry"
	"golang.org/x/oauth2"
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
	detailsURL  string
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
		detailsURL:  actions.GetInput("details_url"),
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

	if in.repository == "" {
		repo, err := getRepository()
		if err != nil {
			actions.Fatalf(err.Error())
		}
		in.repository = repo
	}

	if in.sha == "" {
		sha, err := getSHA()
		if err != nil {
			actions.Fatalf(err.Error())
		}
		in.sha = sha
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: in.token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	status := &github.RepoStatus{
		State:       &in.state,
		Context:     &in.context,
		Description: &in.description,
		TargetURL:   &in.detailsURL,
	}

	// Do a fibonacci backoff 1s -> 1s -> 2s -> 3s -> 5s -> 8s
	if err := retry.Do(ctx, retry.WithMaxRetries(5, retry.NewFibonacci(1*time.Second)), func(ctx context.Context) error {
		status, _, err = client.Repositories.CreateStatus(context.Background(), in.owner, in.repository, in.sha, status)
		if err != nil {
			actions.Errorf(err.Error())
			// This marks the error as retryable
			return retry.RetryableError(err)
		}
		return nil
	}); err != nil {
		actions.Fatalf(err.Error())
	}

	commitURL := fmt.Sprintf("https://github.com/%s/%s/commits/%s", in.owner, in.repository, in.sha)
	actions.Infof("Updated status: \nID: %d \nState: %s \nURL: %s ", *status.ID, in.state, commitURL)
}

// getRequiredInputs checks the required inputs and returns an error
// if they are not set.
func getRequiredInputs(in input) error {
	var err *multierror.Error
	if in.token == "" {
		err = multierror.Append(err, errors.New(tokenRequiredErr))
	}

	if in.state == "" {
		err = multierror.Append(err, errors.New(stateRequiredErr))
	}

	if err != nil {
		err.ErrorFormat = func(errs []error) string {
			var errStr []string
			for _, e := range errs {
				errStr = append(errStr, e.Error())
			}
			return strings.Join(errStr, ", ")
		}
		return err
	}
	return nil
}

// getRepositoryOwner gets github.repository_owner from the Github API.
func getOwner() (string, error) {
	owner := os.Getenv("GITHUB_OWNER")
	if owner == "" {
		return "", fmt.Errorf(ownerEnvNotSetErr)
	}
	return owner, nil
}

// getRepositoryOwner gets github.repository_owner from the Github API.
func getRepository() (string, error) {
	repo := os.Getenv("GITHUB_REPOSITORY")
	if repo == "" {
		return "", fmt.Errorf(repositoryEnvNotSetErr)
	}
	return repo, nil
}

// getRepositoryOwner gets github.repository_owner from the Github API.
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
