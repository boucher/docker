package client

import (
  "fmt"

  "github.com/docker/libcontainer"
)

func (cli *DockerCli) CmdRestore(args ...string) error {
    cmd := cli.Subcmd("restore", "CONTAINER [CONTAINER...]", "Restore one or more checkpointed containers", true)

    var (
        flImgDir       = cmd.String([]string{"-image-dir"}, "", "(optional) directory to restore image files from")
        flWorkDir      = cmd.String([]string{"-work-dir"}, "", "directory to store temp files and restore.log")
        flCheckTcp     = cmd.Bool([]string{"-allow-tcp"}, false, "allow restoring tcp connections")
        flExtUnix      = cmd.Bool([]string{"-allow-ext-unix"}, false, "allow restoring external unix connections")
        flShell        = cmd.Bool([]string{"-allow-shell"}, false, "allow restoring shell jobs")
    )

    if err := cmd.ParseFlags(args, true); err != nil {
        return err
    }

    if cmd.NArg() < 1 {
        cmd.Usage()
        return nil
    }

    criuOpts := &libcontainer.CriuOpts{
        ImagesDirectory:         *flImgDir,
        WorkDirectory:           *flWorkDir,
        TcpEstablished:          *flCheckTcp,
        ExternalUnixConnections: *flExtUnix,
        ShellJob:                *flShell,
    }

    var encounteredError error
    for _, name := range cmd.Args() {
        _, _, err := readBody(cli.call("POST", "/containers/"+name+"/restore", criuOpts, nil))
        if err != nil {
            fmt.Fprintf(cli.err, "%s\n", err)
            encounteredError = fmt.Errorf("Error: failed to restore one or more containers")
        } else {
            fmt.Fprintf(cli.out, "%s\n", name)
        }
    }
    return encounteredError
}
