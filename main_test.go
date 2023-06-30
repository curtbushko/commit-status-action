// Copyright (c) Curt Bushko.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"errors"
	"testing"

	"github.com/hashicorp/go-multierror"
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
		errors []error
	}{
		{
			name:   "no_inputs_returns_all_required_fields",
			inputs: input{},
			errors: []error{
				errors.New(tokenRequiredErr),
				errors.New(stateRequiredErr),
			},
		},
		{
			name: "token_input_returns_required_fields",
			inputs: input{
				token: "foo",
			},
			errors: []error{
				errors.New(stateRequiredErr),
			},
		},
		{
			name: "token_and_state_inputs_returns_no_errors",
			inputs: input{
				token: "foo",
				state: "bar",
			},
			errors: nil,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			multiErr := &multierror.Error{Errors: c.errors}
			err := getRequiredInputs(c.inputs)
			require.Equal(t, multiErr.ErrorOrNil(), err)
		})
	}
}
