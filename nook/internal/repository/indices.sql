-- Performance optimization indices for nook database
-- These indices should be created after the initial schema migrations

-- Index on machine_id for SSH keys lookups
CREATE INDEX IF NOT EXISTS idx_ssh_keys_machine_id ON ssh_keys(machine_id);

-- Index on ipv4 for machine lookups by IP
CREATE INDEX IF NOT EXISTS idx_machines_ipv4 ON machines(ipv4);

-- Index on name for machine lookups by name
CREATE INDEX IF NOT EXISTS idx_machines_name ON machines(name);

-- Index on network_id for machine lookups by network
CREATE INDEX IF NOT EXISTS idx_machines_network_id ON machines(network_id);

-- Index on name for network lookups by name
CREATE INDEX IF NOT EXISTS idx_networks_name ON networks(name);

-- Index on bridge for network lookups by bridge
CREATE INDEX IF NOT EXISTS idx_networks_bridge ON networks(bridge);

-- Index on network_id for DHCP range lookups
CREATE INDEX IF NOT EXISTS idx_dhcp_ranges_network_id ON dhcp_ranges(network_id);

-- Index on network_id for IP lease lookups
CREATE INDEX IF NOT EXISTS idx_ip_leases_network_id ON ip_address_leases(network_id);

-- Index on machine_id for IP lease lookups
CREATE INDEX IF NOT EXISTS idx_ip_leases_machine_id ON ip_address_leases(machine_id);

-- Index on ip_address for IP lease lookups
CREATE INDEX IF NOT EXISTS idx_ip_leases_ip_address ON ip_address_leases(ip_address);

-- Composite index for machine-network IP lease lookups
CREATE INDEX IF NOT EXISTS idx_ip_leases_machine_network ON ip_address_leases(machine_id, network_id);