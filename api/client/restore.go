// +build experimental

package client

import (
	"fmt"

	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/runconfig"
)

func (cli *DockerCli) CmdRestore(args ...string) error {
	cmd := cli.Subcmd("restore", []string{"CONTAINER [CONTAINER...]"}, "Restore one or more checkpointed containers", true)
	cmd.Require(flag.Min, 1)

	var (
		flImgDir   = cmd.String([]string{"-image-dir"}, "", "directory to restore image files from")
		flWorkDir  = cmd.String([]string{"-work-dir"}, "", "directory for restore log")
		flCheckTcp = cmd.Bool([]string{"-allow-tcp"}, false, "allow restoring tcp connections")
		flExtUnix  = cmd.Bool([]string{"-allow-ext-unix"}, false, "allow restoring external unix connections")
		flShell    = cmd.Bool([]string{"-allow-shell"}, false, "allow restoring shell jobs")
		flForce    = cmd.Bool([]string{"-force"}, false, "bypass checks for current container state")
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
			TcpEstablished:          *flCheckTcp,
			ExternalUnixConnections: *flExtUnix,
			ShellJob:                *flShell,
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
