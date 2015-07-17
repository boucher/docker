// +build experimental

package main

var (
	experimentalCommands = []command{
		{"checkpoint", "Checkpoint one or more running containers"},
		{"restore", "Restore one or more checkpointed containers"},
	}
)
