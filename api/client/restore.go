// +build experimental

package client

import (
	"fmt"

	Cli "github.com/docker/docker/cli"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/runconfig"
)

// CmdRestore restores the process in a checkpointed container
//
// Usage: docker restore CONTAINER
func (cli *DockerCli) CmdRestore(args ...string) error {
	cmd := Cli.Subcmd("restore", []string{"CONTAINER [CONTAINER...]"}, "Restore one or more checkpointed containers", true)
	cmd.Require(flag.Min, 1)

	var (
		flImgDir  = cmd.String([]string{"-image-dir"}, "", "directory to restore image files from")
		flWorkDir = cmd.String([]string{"-work-dir"}, "", "directory for restore log")
		flForce   = cmd.Bool([]string{"-force"}, false, "bypass checks for current container state")
	)

	if err := cmd.ParseFlags(args, true); err != nil {
		return err
	}

	if cmd.NArg() < 1 {
		cmd.Usage()
		return nil
	}

	restoreOpts := &runconfig.RestoreConfig{
		CriuOpts: runconfig.CriuConfig{
			ImagesDirectory:         *flImgDir,
			WorkDirectory:           *flWorkDir,
			TCPEstablished:          true,
			ExternalUnixConnections: true,
			FileLocks:               true,
		},
		ForceRestore: *flForce,
	}

	var encounteredError error
	for _, name := range cmd.Args() {
		_, _, err := readBody(cli.call("POST", "/containers/"+name+"/restore", restoreOpts, nil))
		if err != nil {
			fmt.Fprintf(cli.err, "%s\n", err)
			encounteredError = fmt.Errorf("Error: failed to restore one or more containers")
		} else {
			fmt.Fprintf(cli.out, "%s\n", name)
		}
	}
	return encounteredError
}
