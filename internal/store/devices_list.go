package store

import (
	"context"
	"fmt"
	"strings"
)

// RegisteredDevice is a device row stored in PostgreSQL (may or may not be USB-connected).
type RegisteredDevice struct {
	Serial string
	Label  string
}

func (s *Store) ListRegisteredDevices(ctx context.Context) ([]RegisteredDevice, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT serial, label FROM devices WHERE status = 'active' ORDER BY adb_index, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RegisteredDevice
	for rows.Next() {
		var d RegisteredDevice
		if err := rows.Scan(&d.Serial, &d.Label); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Store) CreateAccount(ctx context.Context, name, email, profileID, password string) (int64, error) {
	email = strings.TrimSpace(email)
	profileID = strings.TrimSpace(profileID)
	password = strings.TrimSpace(password)
	name = strings.TrimSpace(name)
	if password == "" {
		return 0, fmt.Errorf("password wajib diisi")
	}
	if email == "" && profileID == "" {
		return 0, fmt.Errorf("email atau profile ID wajib diisi")
	}
	var id int64
	err := s.pool.QueryRow(ctx, `
		INSERT INTO facebook_accounts (name, email, password, profile_id)
		VALUES ($1, $2, $3, $4) RETURNING id`,
		name, email, password, profileID).Scan(&id)
	return id, err
}
