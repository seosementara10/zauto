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

// UpdateAccount updates name/login. Empty password keeps the existing value.
func (s *Store) UpdateAccount(ctx context.Context, id int64, name, email, profileID, password string) error {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)
	profileID = strings.TrimSpace(profileID)
	password = strings.TrimSpace(password)
	if email == "" && profileID == "" {
		return fmt.Errorf("email atau profile ID wajib diisi")
	}
	var n int64
	if password != "" {
		tag, err := s.pool.Exec(ctx, `
			UPDATE facebook_accounts
			SET name = $1, email = $2, profile_id = $3, password = $4
			WHERE id = $5 AND status = 'active'`,
			name, email, profileID, password, id)
		if err != nil {
			return err
		}
		n = tag.RowsAffected()
	} else {
		tag, err := s.pool.Exec(ctx, `
			UPDATE facebook_accounts
			SET name = $1, email = $2, profile_id = $3
			WHERE id = $4 AND status = 'active'`,
			name, email, profileID, id)
		if err != nil {
			return err
		}
		n = tag.RowsAffected()
	}
	if n == 0 {
		return fmt.Errorf("akun %d tidak ditemukan", id)
	}
	return nil
}
