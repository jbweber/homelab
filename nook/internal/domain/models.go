package domain

// Machine represents a virtual machine in the system
type Machine struct {
	ID       int64  // Unique identifier
	Name     string // Machine name
	Hostname string // Hostname for NoCloud metadata
	IPv4     string // Unique IPv4 address
}

// SSHKey represents an SSH public key associated with a machine
type SSHKey struct {
	ID        int64  // Unique identifier
	MachineID int64  // Foreign key to Machine
	KeyText   string // Public SSH key text
}
