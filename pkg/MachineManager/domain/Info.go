package domain

type StartNeworkingOpts struct {
	// Create an random virtual ethernet interface
	Veth bool

	// Provide Veth-name as hostname:clienname or only hostname
	VethName string

	// Provide an existing bridge
	BridgeName string
}

type StartOpts struct {
	MachineName string
	Hostname    string

	ImageFileName string
	IsSystemd     bool // start systemd in container
	Ephemeral     bool // persistance disabled

	CopyResolvConf bool

	Environment map[string]string

	Bind         map[string]string
	BindReadOnly map[string]string

	Networking StartNeworkingOpts

	StdOutPath string
	StdErrPath string
}

type MachineInfo struct {
	OS        string
	Class     string
	Service   string
	Version   string
	Addresses struct {
		IPv4 string
		IPv6 string
	}
}

type MachineList map[string]MachineInfo
