// Copyright (c) Curt Bushko.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	actions "github.com/sethvargo/go-githubactions"
)

type input struct {
	token           string
	state           string
	description     string
	product         string
	releaseMetadata string
	repo            string
	org             string
	securityScan    string
	sha             string
	version         string
}

func main() {

}
