package runconfig


type VethPairName struct {
	ContainerInterfaceName string
	HostInterfaceName string
}

// CriuConfig holds configuration options passed down to libcontainer and CRIU
type CriuConfig struct {
	ImagesDirectory         string
	WorkDirectory           string
	LeaveRunning            bool
	TCPEstablished          bool
	ExternalUnixConnections bool
	ShellJob                bool
	FileLocks               bool
	VethPairs               []VethPairName
}

// RestoreConfig holds the restore command options, which is a superset of the CRIU options
type RestoreConfig struct {
	CriuOpts     CriuConfig
	ForceRestore bool
}
