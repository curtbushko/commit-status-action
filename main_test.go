// Copyright (c) Curt Bushko.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
			name:          "fail with invalid state",
			actual:        "foo",
			expectedValue: "",
			expectError:   true,
		},
		{
			name:          "change cancelled to error",
			actual:        "cancelled",
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
