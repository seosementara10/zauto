package store

import "context"

// LoadDeviceMirrorEnabled returns mirror/switch ON flags persisted per serial.
func (s *Store) LoadDeviceMirrorEnabled(ctx context.Context) (map[string]bool, error) {
	out := map[string]bool{}
	rows, err := s.pool.Query(ctx, `
		SELECT serial, mirror_enabled FROM devices WHERE status = 'active'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var serial string
		var on bool
		if err := rows.Scan(&serial, &on); err != nil {
			return nil, err
		}
		out[serial] = on
	}
	return out, rows.Err()
}

// SetDeviceMirrorEnabled persists one device's switch ON/OFF state.
func (s *Store) SetDeviceMirrorEnabled(ctx context.Context, serial string, enabled bool) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO devices (serial, label, mirror_enabled, updated_at)
		VALUES ($1, $1, $2, now())
		ON CONFLICT (serial) DO UPDATE SET
			mirror_enabled = EXCLUDED.mirror_enabled,
			updated_at = now()`, serial, enabled)
	return err
}

// SetDevicesMirrorEnabled bulk-updates mirror flags (enable-all / disable-all).
func (s *Store) SetDevicesMirrorEnabled(ctx context.Context, serials []string, enabled bool) error {
	if len(serials) == 0 {
		return nil
	}
	_, err := s.pool.Exec(ctx, `
		UPDATE devices SET mirror_enabled = $2, updated_at = now()
		WHERE status = 'active' AND serial = ANY($1)`, serials, enabled)
	return err
}
