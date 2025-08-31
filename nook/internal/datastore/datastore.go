package datastore

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type Machine struct {
	ID   int64  // Unique identifier
	Name string // Machine name
	IPv4 string // Unique IPv4 address
}

type SSHKey struct {
	ID        int64  // Unique identifier
	MachineID int64  // Foreign key to Machine
	KeyText   string // Public SSH key text
}

type Datastore struct {
	DB *sql.DB
}

// New creates a new Datastore and runs migrations.
func New(path string) (*Datastore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		return nil, err
	}
	return &Datastore{DB: db}, nil
}

// migrate creates tables for machines and ssh_keys if they do not exist.
func migrate(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS machines (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		ipv4 TEXT NOT NULL UNIQUE
	);`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS ssh_keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		machine_id INTEGER NOT NULL,
		key_text TEXT NOT NULL,
		FOREIGN KEY(machine_id) REFERENCES machines(id)
	);`)
	return err
}

// CreateMachine inserts a new machine into the database, validating required fields.
func (ds *Datastore) CreateMachine(m Machine) (Machine, error) {
	if m.Name == "" {
		return Machine{}, fmt.Errorf("machine name is required")
	}
	if m.IPv4 == "" {
		return Machine{}, fmt.Errorf("machine IPv4 is required")
	}
	res, err := ds.DB.Exec("INSERT INTO machines (name, ipv4) VALUES (?, ?)", m.Name, m.IPv4)
	if err != nil {
		return Machine{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Machine{}, err
	}
	m.ID = id
	return m, nil
}

// GetMachine retrieves a machine by ID.
func (ds *Datastore) GetMachine(id int64) (*Machine, error) {
	var m Machine
	err := ds.DB.QueryRow("SELECT id, name, ipv4 FROM machines WHERE id = ?", id).Scan(&m.ID, &m.Name, &m.IPv4)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

// ListMachines returns all machines in the database.
func (ds *Datastore) ListMachines() ([]Machine, error) {
	rows, err := ds.DB.Query("SELECT id, name, ipv4 FROM machines")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var machines []Machine
	for rows.Next() {
		var m Machine
		if err := rows.Scan(&m.ID, &m.Name, &m.IPv4); err != nil {
			return nil, err
		}
		machines = append(machines, m)
	}
	return machines, nil
}

// DeleteMachine removes a machine by ID.
func (ds *Datastore) DeleteMachine(id int64) error {
	_, err := ds.DB.Exec("DELETE FROM machines WHERE id = ?", id)
	return err
}

// GetMachineByName retrieves a machine by its unique name.
func (ds *Datastore) GetMachineByName(name string) (*Machine, error) {
	var m Machine
	err := ds.DB.QueryRow("SELECT id, name, ipv4 FROM machines WHERE name = ?", name).Scan(&m.ID, &m.Name, &m.IPv4)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

// GetMachineByIPv4 retrieves a machine by its unique IPv4 address.
func (ds *Datastore) GetMachineByIPv4(ipv4 string) (*Machine, error) {
	var m Machine
	err := ds.DB.QueryRow("SELECT id, name, ipv4 FROM machines WHERE ipv4 = ?", ipv4).Scan(&m.ID, &m.Name, &m.IPv4)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

// ListSSHKeys returns all SSH keys for a given machine, ordered by id.
func (ds *Datastore) ListSSHKeys(machineID int64) ([]SSHKey, error) {
	rows, err := ds.DB.Query("SELECT id, machine_id, key_text FROM ssh_keys WHERE machine_id = ? ORDER BY id ASC", machineID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var keys []SSHKey
	for rows.Next() {
		var k SSHKey
		if err := rows.Scan(&k.ID, &k.MachineID, &k.KeyText); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, nil
}

// CreateSSHKey inserts a new SSH key for a machine.
func (ds *Datastore) CreateSSHKey(machineID int64, keyText string) (int64, error) {
	res, err := ds.DB.Exec("INSERT INTO ssh_keys (machine_id, key_text) VALUES (?, ?)", machineID, keyText)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetSSHKey retrieves an SSH key by ID.
func (ds *Datastore) GetSSHKey(id int64) (*SSHKey, error) {
	var k SSHKey
	err := ds.DB.QueryRow("SELECT id, machine_id, key_text FROM ssh_keys WHERE id = ?", id).Scan(&k.ID, &k.MachineID, &k.KeyText)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &k, nil
}

// DeleteSSHKey removes an SSH key by ID.
func (ds *Datastore) DeleteSSHKey(id int64) error {
	_, err := ds.DB.Exec("DELETE FROM ssh_keys WHERE id = ?", id)
	return err
}
