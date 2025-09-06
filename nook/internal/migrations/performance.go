package migrations

import (
	"database/sql"
)

// GetPerformanceMigrations returns performance optimization migrations
func GetPerformanceMigrations() []Migration {
	return []Migration{
		{
			Version: 10,
			Name:    "add_performance_indices",
			Up: func(db *sql.DB) error {
				// Add indices for better query performance
				indices := []string{
					"CREATE INDEX IF NOT EXISTS idx_ssh_keys_machine_id ON ssh_keys(machine_id)",
					"CREATE INDEX IF NOT EXISTS idx_machines_ipv4 ON machines(ipv4)",
					"CREATE INDEX IF NOT EXISTS idx_machines_name ON machines(name)",
					"CREATE INDEX IF NOT EXISTS idx_machines_network_id ON machines(network_id)",
					"CREATE INDEX IF NOT EXISTS idx_networks_name ON networks(name)",
					"CREATE INDEX IF NOT EXISTS idx_networks_bridge ON networks(bridge)",
					"CREATE INDEX IF NOT EXISTS idx_dhcp_ranges_network_id ON dhcp_ranges(network_id)",
				}

				for _, indexSQL := range indices {
					if _, err := db.Exec(indexSQL); err != nil {
						return err
					}
				}

				return nil
			},
			Down: func(db *sql.DB) error {
				// Drop performance indices
				indices := []string{
					"DROP INDEX IF EXISTS idx_ssh_keys_machine_id",
					"DROP INDEX IF EXISTS idx_machines_ipv4",
					"DROP INDEX IF EXISTS idx_machines_name",
					"DROP INDEX IF EXISTS idx_machines_network_id",
					"DROP INDEX IF EXISTS idx_networks_name",
					"DROP INDEX IF EXISTS idx_networks_bridge",
					"DROP INDEX IF EXISTS idx_dhcp_ranges_network_id",
				}

				for _, dropSQL := range indices {
					if _, err := db.Exec(dropSQL); err != nil {
						return err
					}
				}

				return nil
			},
		},
	}
}
