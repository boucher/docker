package runconfig

type VethPairName struct {
	InName  string
	OutName string
}

type CriuConfig struct {
	ImagesDirectory         string
	WorkDirectory           string
	LeaveRunning            bool
	TcpEstablished          bool
	ExternalUnixConnections bool
	ShellJob                bool
	VethPairs               []VethPairName
}

type RestoreConfig struct {
	CriuOpts     CriuConfig
	ForceRestore bool
}
