package daemon

import (
	"fmt"

	"github.com/docker/docker/pkg/promise"
	"github.com/docker/docker/runconfig"
	"github.com/docker/libnetwork/netutils"
)

// Checkpoint checkpoints the running container, saving its state afterwards
func (container *Container) Checkpoint(opts *runconfig.CriuConfig) error {
	if err := container.daemon.Checkpoint(container, opts); err != nil {
		return err
	}

	if opts.LeaveRunning == false {
		container.cleanup()
	}

	if err := container.toDisk(); err != nil {
		return fmt.Errorf("Cannot update config for container: %s", err)
	}

	return nil
}

// Restore restores the container's process from images on disk
func (container *Container) Restore(opts *runconfig.CriuConfig, forceRestore bool) error {
	var err error
	container.Lock()
	defer container.Unlock()

	defer func() {
		if err != nil {
			container.setError(err)
			// if no one else has set it, make sure we don't leave it at zero
			if container.ExitCode == 0 {
				container.ExitCode = 128
			}
			container.toDisk()
			container.cleanup()
		}
	}()

	if err := container.Mount(); err != nil {
		return err
	}

	// Make sure NetworkMode has an acceptable value. We do this to ensure
	// backwards API compatibility.
	container.hostConfig = runconfig.SetDefaultNetModeIfBlank(container.hostConfig)
	if err = container.initializeNetworking(true); err != nil {
		return err
	}

	// This is an old implementation, to remove later
	//
	// nctl := container.daemon.netController
	// network, err := nctl.NetworkByID(container.NetworkSettings.NetworkID)
	// ep_t, err := network.EndpointByID(container.NetworkSettings.EndpointID)
	// sb, err := nctl.SandboxByID(container.NetworkSettings.SandboxID)
	// if err != nil {
	//	return err
	//}

	// for _, i := range ep_t.SandboxInterfaces() {
	// for _, i := range sb.Interfaces() {
	//	outname, err := netutils.GenerateIfaceName("veth", 7)
	//	if err != nil {
	//		return err
	//	}
	//	vethpair := runconfig.VethPairName{
	//		InName:  i.DstName(),
	//		OutName: outname,
	//	}
	//	opts.VethPairs = append(opts.VethPairs, vethpair)
	// }

	// TODO: the following implemtantion is temporary.
	//       re-implmentation after libnetwork patch is merged to upstream
	outname, err := netutils.GenerateIfaceName("veth", 7)
	if err != nil {
		return err
	}
	vethpair := runconfig.VethPairName{
		InName:  "eth0",
		OutName: outname,
	}
	opts.VethPairs = append(opts.VethPairs, vethpair)

	linkedEnv, err := container.setupLinkedContainers()
	if err != nil {
		return err
	}

	if err = container.setupWorkingDirectory(); err != nil {
		return err
	}

	env := container.createDaemonEnvironment(linkedEnv)
	if err = populateCommand(container, env); err != nil {
		return err
	}

	if !container.hostConfig.IpcMode.IsContainer() && !container.hostConfig.IpcMode.IsHost() {
		if err := container.setupIpcDirs(); err != nil {
			return err
		}
	}

	mounts, err := container.setupMounts()
	if err != nil {
		return err
	}
	mounts = append(mounts, container.ipcMounts()...)

	container.command.Mounts = mounts
	return container.waitForRestore(opts, forceRestore)
}

func (container *Container) waitForRestore(opts *runconfig.CriuConfig, forceRestore bool) error {
	container.monitor = newContainerMonitor(container, container.hostConfig.RestartPolicy)

	// After calling promise.Go() we'll have two goroutines:
	// - The current goroutine that will block in the select
	//   below until restore is done.
	// - A new goroutine that will restore the container and
	//   wait for it to exit.
	select {
	case <-container.monitor.restoreSignal:
		if container.ExitCode != 0 {
			return fmt.Errorf("restore process failed")
		}
	case err := <-promise.Go(func() error { return container.monitor.Restore(opts, forceRestore) }):
		return err
	}

	return nil
}
