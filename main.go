// Copyright IBM Corp. 2019, 2025
// SPDX-License-Identifier: MPL-2.0

package main

import (
	// TODO: update the path below to match your own repository

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/plugins"
	"github.com/stackshadow/nspawn2/nspawndriver"
)

func main() {
	// Serve the plugin
	plugins.Serve(factory)
}

// factory returns a new instance of a nomad driver plugin
func factory(log hclog.Logger) interface{} {
	return nspawndriver.NewPlugin(log)
}
