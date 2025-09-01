package domain

// Machine represents a virtual machine in the system
type Machine struct {
	ID        int64  // Unique identifier
	Name      string // Machine name
	Hostname  string // Hostname for NoCloud metadata
	IPv4      string // Static IPv4 address (optional, for static assignments)
	NetworkID *int64 // Network ID for dynamic IP assignment (optional)
}

// SSHKey represents an SSH public key associated with a machine
type SSHKey struct {
	ID        int64  // Unique identifier
	MachineID int64  // Foreign key to Machine
	KeyText   string // Public SSH key text
}

// Network represents a network configuration on a hypervisor
type Network struct {
	ID          int64  // Unique identifier
	Name        string // Network name (e.g., "br0", "internal")
	Bridge      string // Bridge interface name (e.g., "br0")
	Subnet      string // Subnet in CIDR notation (e.g., "192.168.1.0/24")
	Gateway     string // Gateway IP address
	DNSServers  string // Comma-separated DNS server IPs
	Description string // Optional description
}

// DHCPRange represents a DHCP range within a network
type DHCPRange struct {
	ID        int64  // Unique identifier
	NetworkID int64  // Foreign key to Network
	StartIP   string // Start of DHCP range
	EndIP     string // End of DHCP range
	LeaseTime string // DHCP lease time (e.g., "12h", "24h")
}

// IPAddressLease represents an IP address leased to a machine from a network
type IPAddressLease struct {
	ID        int64  // Unique identifier
	MachineID int64  // Foreign key to Machine
	NetworkID int64  // Foreign key to Network
	IPAddress string // The leased IP address
	LeaseTime string // Lease duration (e.g., "24h", "infinite")
	CreatedAt string // When the lease was created
	UpdatedAt string // When the lease was last updated
}
