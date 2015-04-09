package daemon

import (
    "fmt"

    "github.com/docker/docker/engine"
    "github.com/docker/libcontainer"
)

// Checkpoint a running container.
func (daemon *Daemon) ContainerCheckpoint(job *engine.Job) error {
   if len(job.Args) != 1 {
       return fmt.Errorf("Usage: %s CONTAINER\n", job.Name)
   }

   name := job.Args[0]
   container, err := daemon.Get(name)
   if err != nil {
       return err
   }
   if !container.IsRunning() {
       return fmt.Errorf("Container %s not running", name)
   }

   opts := &libcontainer.CriuOpts{}

   if job.EnvExists("ImagesDirectory") {
      opts.ImagesDirectory = job.Getenv("ImagesDirectory")
   }
   if job.EnvExists("PreviousImagesDirectory") {
      opts.PreviousImagesDirectory = job.Getenv("PreviousImagesDirectory")
   }
   if job.EnvExists("LeaveRunning") {
      opts.LeaveRunning = job.GetenvBool("LeaveRunning")
   }
   if job.EnvExists("TcpEstablished") {
      opts.TcpEstablished = job.GetenvBool("TcpEstablished")
   }
   if job.EnvExists("ExternalUnixConnections") {
      opts.ExternalUnixConnections = job.GetenvBool("ExternalUnixConnections")
   }
   if job.EnvExists("ShellJob") {
      opts.ShellJob = job.GetenvBool("ShellJob")
   }

   if err := container.Checkpoint(opts); err != nil {
       return fmt.Errorf("Cannot checkpoint container %s: %s", name, err)
   }

   container.LogEvent("checkpoint")
   return nil
}

// Restore a checkpointed container.
func (daemon *Daemon) ContainerRestore(job *engine.Job) error {
   if len(job.Args) != 1 {
       return fmt.Errorf("Usage: %s CONTAINER\n", job.Name)
   }

   name := job.Args[0]
   container, err := daemon.Get(name)
   if err != nil {
       return err
   }

   if container.IsRunning() {
       return fmt.Errorf("Container %s already running", name)
   }

   // TODO: how should we handle the notion of checkpoint and keep running?
   // right now, having ever been checkpointed is sufficient for our desires
   // still requires manually calling stop before you are able to then restore
   if !container.HasBeenCheckpointed() {
       return fmt.Errorf("Container %s is not checkpointed", name)
   }

   opts := &libcontainer.CriuOpts{}

   if job.EnvExists("ImagesDirectory") {
      opts.ImagesDirectory = job.Getenv("ImagesDirectory")
   }
   if job.EnvExists("TcpEstablished") {
      opts.TcpEstablished = job.GetenvBool("TcpEstablished")
   }
   if job.EnvExists("ExternalUnixConnections") {
      opts.ExternalUnixConnections = job.GetenvBool("ExternalUnixConnections")
   }
   if job.EnvExists("ShellJob") {
      opts.ShellJob = job.GetenvBool("ShellJob")
   }

   if err = container.Restore(opts); err != nil {
       container.LogEvent("die")
       return fmt.Errorf("Cannot restore container %s: %s", name, err)
   }

   container.LogEvent("restore")
   return nil
}
