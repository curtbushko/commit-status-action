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

type ghRepositoryClient interface {
	CreateStatus(context.Context, string, string, string, *github.RepoStatus) (*github.RepoStatus, *github.Response, error)
}

type getInputFunc func(string) string

type ghClient struct {
	client               ghRepositoryClient
	input                input
	maxConnectionRetries uint64
}

func main() {
	ctx := context.Background()
	client, err := newGHClient(ctx, uint64(5), actions.GetInput)
	if err != nil {
		actions.Fatalf(err.Error())
	}
	err = client.createStatus(ctx)
	if err != nil {
		actions.Fatalf(err.Error())
	}
}

// newGHClient creates a new github client for creating a github repo status.
func newGHClient(ctx context.Context, maxConnectionRetries uint64, getInputFunc getInputFunc) (ghClient, error) {
	in, err := getInputs(getInputFunc)
	if err != nil {
		return ghClient{}, err
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: in.token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc).Repositories

	return ghClient{
		client:               client,
		input:                in,
		maxConnectionRetries: maxConnectionRetries,
	}, nil
}

// createStatus creates a new github repo status.
func (gh *ghClient) createStatus(ctx context.Context) error {
	status := &github.RepoStatus{
		State:       &gh.input.state,
		Context:     &gh.input.context,
		Description: &gh.input.description,
		TargetURL:   &gh.input.detailsURL,
	}

	// Do a fibonacci backoff 1s -> 1s -> 2s -> 3s -> 5s -> 8s
	var err error
	err = retry.Do(ctx, retry.WithMaxRetries(gh.maxConnectionRetries, retry.NewFibonacci(1*time.Second)), func(ctx context.Context) error {
		status, _, err = gh.client.CreateStatus(context.Background(), gh.input.owner, gh.input.repository, gh.input.sha, status)
		if err != nil {
			actions.Errorf("Error creating status %v. Owner: %s, SHA: %s, Repo %s: %s", gh.input.state, gh.input.owner, gh.input.sha, gh.input.repository, err.Error())
			// This marks the error as retryable
			return retry.RetryableError(err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	// We are going to access the status ID so make sure it is valid before proceeding.
	if status.ID == nil {
		return errors.New("status ID returned is nil")
	}

	commitURL := fmt.Sprintf("https://github.com/%s/%s/commits/%s", gh.input.owner, gh.input.repository, gh.input.sha)
	actions.Infof("Updated status: \nID: %d \nState: %s \nURL: %s ", *status.ID, gh.input.state, commitURL)
	return nil
}

// getInputs loads in all the inputs from the action and returns them as a struct.
func getInputs(getInput getInputFunc) (input, error) {
	// Read in Input
	in := input{
		token:       getInput("token"),
		state:       getInput("state"),
		context:     getInput("context"),
		description: getInput("description"),
		owner:       getInput("owner"),
		repository:  getInput("repository"),
		sha:         getInput("sha"),
		detailsURL:  getInput("details_url"),
	}

	// Convert State to a repo status
	var err error
	in.state, err = convertActionStateToRepoStatusState(in.state)
	if err != nil {
		return input{}, err
	}

	// Set Input Defaults
	in, err = setInputDefaults(in)
	if err != nil {
		return input{}, err
	}

	// Validate inputs before proceeding
	err = validateRequiredInputs(in)
	if err != nil {
		return input{}, err
	}

	return in, err
}

// setInputDefaults sets the default values for inputs that are not required.
func setInputDefaults(in input) (input, error) {
	// Set Defaults
	if in.owner == "" {
		owner, err := getOwner()
		if err != nil {
			return input{}, err
		}
		in.owner = owner
	}

	if in.repository == "" {
		repo, err := getRepository()
		if err != nil {
			return input{}, err
		}
		in.repository = repo
	}
	in.repository = removeOwnerFromRepository(in.repository, in.owner)

	if in.sha == "" {
		sha, err := getSHA()
		if err != nil {
			return input{}, err
		}
		in.sha = sha
	}

	return in, nil
}

// validateRequiredInputs validates that all required inputs are set.
func validateRequiredInputs(in input) error {
	// Accumulate errors
	var errs *multierror.Error

	if in.token == "" {
		errs = multierror.Append(errs, errors.New(tokenRequiredErr))
	}

	if in.state == "" {
		errs = multierror.Append(errs, errors.New(stateRequiredErr))
	}

	if errs != nil {
		errs.ErrorFormat = func(errs []error) string {
			var errStr []string
			for _, e := range errs {
				errStr = append(errStr, e.Error())
			}
			return strings.Join(errStr, ", ")
		}
		return errs
	}

	return nil
}

// getOwner gets github.repository_owner from the Github API.
func getOwner() (string, error) {
	owner := os.Getenv("GITHUB_OWNER")
	if owner == "" {
		return "", fmt.Errorf(ownerEnvNotSetErr)
	}
	return owner, nil
}

// getRepository gets github.repository_owner from the Github API.
func getRepository() (string, error) {
	repo := os.Getenv("GITHUB_REPOSITORY")
	if repo == "" {
		return "", fmt.Errorf(repositoryEnvNotSetErr)
	}
	return repo, nil
}

// getSHA gets github.repository_owner from the Github API.
func getSHA() (string, error) {
	sha := os.Getenv("GITHUB_SHA")
	if sha == "" {
		return "", fmt.Errorf(shaEnvNotSetErr)
	}
	return sha, nil
}

// convertActionStateToRepoStatusState validates that the state is a correct value. If the state is
// 'cancel', 'cancelled'  or 'skipped' then return 'error' as the state.
// 'Cancelled' can be a valid state if a workflow is cancelled.
func convertActionStateToRepoStatusState(actionState string) (string, error) {
	switch actionState {
	// success, error, failure or pending
	case "error", "failure", "pending", "success":
		return actionState, nil
	case "cancel", "cancelled", "skipped":
		return "error", nil
	default:
		return "", fmt.Errorf("state value not supported: %s", actionState)
	}
}

// removeOwnerFromRepository removes the owner from the repository string.
func removeOwnerFromRepository(repo, owner string) string {
	return strings.ReplaceAll(repo, fmt.Sprintf("%s/", owner), "")
}
