package migrations

import (
	"database/sql"
)

// GetInitialMigrations returns all initial migrations
func GetInitialMigrations() []Migration {
	migrations := []Migration{
		{
			Version: 1,
			Name:    "create_initial_tables",
			Up: func(db *sql.DB) error {
				// Check if tables already exist (for backward compatibility)
				var count int
				err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='machines'").Scan(&count)
				if err != nil {
					return err
				}

				if count > 0 {
					// Tables already exist, check if we need to add new columns
					return upgradeExistingTables(db)
				}

				// Create machines table
				_, err = db.Exec(`
					CREATE TABLE machines (
						id INTEGER PRIMARY KEY AUTOINCREMENT,
						name TEXT NOT NULL UNIQUE,
						hostname TEXT NOT NULL,
						ipv4 TEXT NOT NULL UNIQUE,
						network_id INTEGER,
						created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
						updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
					)
				`)
				if err != nil {
					return err
				}

				// Create ssh_keys table
				_, err = db.Exec(`
					CREATE TABLE ssh_keys (
						id INTEGER PRIMARY KEY AUTOINCREMENT,
						machine_id INTEGER NOT NULL,
						key_text TEXT NOT NULL,
						created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
						updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
						FOREIGN KEY (machine_id) REFERENCES machines(id) ON DELETE CASCADE
					)
				`)
				if err != nil {
					return err
				}

				// Create indexes for better performance
				_, err = db.Exec(`CREATE INDEX idx_ssh_keys_machine_id ON ssh_keys(machine_id)`)
				if err != nil {
					return err
				}

				_, err = db.Exec(`CREATE INDEX idx_machines_ipv4 ON machines(ipv4)`)
				return err
			},
			Down: func(db *sql.DB) error {
				// Drop tables in reverse order due to foreign key constraints
				_, err := db.Exec(`DROP TABLE IF EXISTS ssh_keys`)
				if err != nil {
					return err
				}

				_, err = db.Exec(`DROP TABLE IF EXISTS machines`)
				return err
			},
		},
		{
			Version: 2,
			Name:    "create_networks_table",
			Up: func(db *sql.DB) error {
				// Create networks table
				_, err := db.Exec(`
					CREATE TABLE networks (
						id INTEGER PRIMARY KEY AUTOINCREMENT,
						name TEXT NOT NULL UNIQUE,
						bridge TEXT NOT NULL,
						subnet TEXT NOT NULL,
						gateway TEXT,
						dns_servers TEXT,
						description TEXT,
						created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
						updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
					)
				`)
				if err != nil {
					return err
				}

				// Create dhcp_ranges table
				_, err = db.Exec(`
					CREATE TABLE dhcp_ranges (
						id INTEGER PRIMARY KEY AUTOINCREMENT,
						network_id INTEGER NOT NULL,
						start_ip TEXT NOT NULL,
						end_ip TEXT NOT NULL,
						lease_time TEXT DEFAULT '24h',
						created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
						updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
						FOREIGN KEY (network_id) REFERENCES networks(id) ON DELETE CASCADE
					)
				`)
				if err != nil {
					return err
				}

				// Create indexes
				_, err = db.Exec(`CREATE INDEX idx_dhcp_ranges_network_id ON dhcp_ranges(network_id)`)
				if err != nil {
					return err
				}

				_, err = db.Exec(`CREATE INDEX idx_networks_bridge ON networks(bridge)`)
				return err
			},
			Down: func(db *sql.DB) error {
				// Drop tables in reverse order due to foreign key constraints
				_, err := db.Exec(`DROP TABLE IF EXISTS dhcp_ranges`)
				if err != nil {
					return err
				}

				_, err = db.Exec(`DROP TABLE IF EXISTS networks`)
				return err
			},
		},
		{
			Version: 3,
			Name:    "create_ip_address_leases_table",
			Up: func(db *sql.DB) error {
				// Create ip_address_leases table
				_, err := db.Exec(`
					CREATE TABLE ip_address_leases (
						id INTEGER PRIMARY KEY AUTOINCREMENT,
						machine_id INTEGER NOT NULL,
						network_id INTEGER NOT NULL,
						ip_address TEXT NOT NULL UNIQUE,
						lease_time TEXT DEFAULT '24h',
						created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
						updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
						FOREIGN KEY (machine_id) REFERENCES machines(id) ON DELETE CASCADE,
						FOREIGN KEY (network_id) REFERENCES networks(id) ON DELETE CASCADE,
						UNIQUE(machine_id, network_id) -- One IP per machine per network
					)
				`)
				if err != nil {
					return err
				}

				// Create indexes for performance
				_, err = db.Exec(`CREATE INDEX idx_ip_leases_machine_id ON ip_address_leases(machine_id)`)
				if err != nil {
					return err
				}

				_, err = db.Exec(`CREATE INDEX idx_ip_leases_network_id ON ip_address_leases(network_id)`)
				if err != nil {
					return err
				}

				_, err = db.Exec(`CREATE INDEX idx_ip_leases_ip_address ON ip_address_leases(ip_address)`)
				return err
			},
			Down: func(db *sql.DB) error {
				_, err := db.Exec(`DROP TABLE IF EXISTS ip_address_leases`)
				return err
			},
		},
	}

	// Append performance migrations
	migrations = append(migrations, GetPerformanceMigrations()...)
	return migrations
}

// upgradeExistingTables upgrades existing tables to add new columns if needed
func upgradeExistingTables(db *sql.DB) error {
	// Check if created_at column exists in machines table
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('machines') WHERE name='created_at'").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		// For SQLite, we need to recreate the table to add columns with defaults
		// First, backup existing data
		rows, err := db.Query("SELECT id, name, hostname, ipv4 FROM machines")
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := rows.Close(); closeErr != nil {
				// Log error but don't fail migration
			}
		}()

		var machines []struct {
			id       int64
			name     string
			hostname string
			ipv4     string
		}

		for rows.Next() {
			var m struct {
				id       int64
				name     string
				hostname string
				ipv4     string
			}
			if err := rows.Scan(&m.id, &m.name, &m.hostname, &m.ipv4); err != nil {
				return err
			}
			machines = append(machines, m)
		}

		// Drop existing table
		_, err = db.Exec(`DROP TABLE machines`)
		if err != nil {
			return err
		}

		// Recreate with new schema
		_, err = db.Exec(`
			CREATE TABLE machines (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL UNIQUE,
				hostname TEXT NOT NULL,
				ipv4 TEXT NOT NULL UNIQUE,
				network_id INTEGER,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)
		`)
		if err != nil {
			return err
		}

		// Restore data
		for _, m := range machines {
			_, err = db.Exec("INSERT INTO machines (id, name, hostname, ipv4) VALUES (?, ?, ?, ?)",
				m.id, m.name, m.hostname, m.ipv4)
			if err != nil {
				return err
			}
		}
	}

	// Check if created_at column exists in ssh_keys table
	err = db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('ssh_keys') WHERE name='created_at'").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		// For SQLite, we need to recreate the table to add columns with defaults
		// First, backup existing data
		rows, err := db.Query("SELECT id, machine_id, key_text FROM ssh_keys")
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := rows.Close(); closeErr != nil {
				// Log error but don't fail migration
			}
		}()

		var sshKeys []struct {
			id        int64
			machineID int64
			keyText   string
		}

		for rows.Next() {
			var k struct {
				id        int64
				machineID int64
				keyText   string
			}
			if err := rows.Scan(&k.id, &k.machineID, &k.keyText); err != nil {
				return err
			}
			sshKeys = append(sshKeys, k)
		}

		// Drop existing table
		_, err = db.Exec(`DROP TABLE ssh_keys`)
		if err != nil {
			return err
		}

		// Recreate with new schema
		_, err = db.Exec(`
			CREATE TABLE ssh_keys (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				machine_id INTEGER NOT NULL,
				key_text TEXT NOT NULL,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (machine_id) REFERENCES machines(id) ON DELETE CASCADE
			)
		`)
		if err != nil {
			return err
		}

		// Restore data
		for _, k := range sshKeys {
			_, err = db.Exec("INSERT INTO ssh_keys (id, machine_id, key_text) VALUES (?, ?, ?)",
				k.id, k.machineID, k.keyText)
			if err != nil {
				return err
			}
		}
	}

	// Create indexes if they don't exist
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_ssh_keys_machine_id ON ssh_keys(machine_id)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_machines_ipv4 ON machines(ipv4)`)
	if err != nil {
		return err
	}

	// Check if ip_address_leases table exists (for migration 3)
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='ip_address_leases'").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		// Create ip_address_leases table for existing databases
		_, err = db.Exec(`
			CREATE TABLE ip_address_leases (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				machine_id INTEGER NOT NULL,
				network_id INTEGER NOT NULL,
				ip_address TEXT NOT NULL UNIQUE,
				lease_time TEXT DEFAULT '24h',
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (machine_id) REFERENCES machines(id) ON DELETE CASCADE,
				FOREIGN KEY (network_id) REFERENCES networks(id) ON DELETE CASCADE,
				UNIQUE(machine_id, network_id)
			)
		`)
		if err != nil {
			return err
		}

		// Create indexes
		_, err = db.Exec(`CREATE INDEX idx_ip_leases_machine_id ON ip_address_leases(machine_id)`)
		if err != nil {
			return err
		}

		_, err = db.Exec(`CREATE INDEX idx_ip_leases_network_id ON ip_address_leases(network_id)`)
		if err != nil {
			return err
		}

		_, err = db.Exec(`CREATE INDEX idx_ip_leases_ip_address ON ip_address_leases(ip_address)`)
		if err != nil {
			return err
		}
	}

	// Check if network_id column exists in machines table
	err = db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('machines') WHERE name='network_id'").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		// Add network_id column to machines table
		_, err = db.Exec(`ALTER TABLE machines ADD COLUMN network_id INTEGER`)
		if err != nil {
			return err
		}

		// Add foreign key constraint
		_, err = db.Exec(`CREATE INDEX idx_machines_network_id ON machines(network_id)`)
		if err != nil {
			return err
		}
	}

	return err
}
