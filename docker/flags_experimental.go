// +build experimental

package main

import "sort"

func init() {
	experimentalCommands := []command{
		{"checkpoint", "Checkpoint one or more running containers"},
		{"network", "Network management"},
		{"restore", "Restore one or more checkpointed containers"},
	}

	dockerCommands = append(dockerCommands, experimentalCommands...)

	//Sorting logic required here to pass Command Sort Test.
	sort.Sort(byName(dockerCommands))
}
