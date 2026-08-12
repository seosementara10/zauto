package store

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schemaSQL string

type Store struct {
	pool *pgxpool.Pool
}

func Open(ctx context.Context, url string) (*Store, error) {
	if url == "" {
		return nil, fmt.Errorf("database url is empty")
	}
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	if cfg.MaxConns == 0 {
		cfg.MaxConns = 8
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	if s != nil && s.pool != nil {
		s.pool.Close()
	}
}

func (s *Store) Ping(ctx context.Context) error {
	if s == nil || s.pool == nil {
		return fmt.Errorf("database not connected")
	}
	return s.pool.Ping(ctx)
}

func (s *Store) Migrate(ctx context.Context) error {
	if _, err := s.pool.Exec(ctx, schemaSQL); err != nil {
		return err
	}
	return s.SeedDefaultPostTexts(ctx)
}
func (s *Store) PostTextCounts(ctx context.Context) (map[string]int, error) {
	out := map[string]int{}
	for _, cat := range []string{PostTextCategoryPersonal, PostTextCategoryFanpage, PostTextCategoryGroup} {
		n, err := s.CountPostTexts(ctx, cat)
		if err != nil {
			return nil, err
		}
		out[cat] = n
	}
	return out, nil
}

func (s *Store) CountAccounts(ctx context.Context) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM facebook_accounts WHERE status = 'active'`).Scan(&n)
	return n, err
}

func (s *Store) CountAssignments(ctx context.Context) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM device_account_slots WHERE active = true`).Scan(&n)
	return n, err
}

func (s *Store) UpsertDevice(ctx context.Context, serial, label string, adbIndex, maxAccounts int) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `
		INSERT INTO devices (serial, label, adb_index, max_accounts, updated_at)
		VALUES ($1, $2, $3, $4, now())
		ON CONFLICT (serial) DO UPDATE SET
			label = EXCLUDED.label,
			adb_index = EXCLUDED.adb_index,
			max_accounts = GREATEST(devices.max_accounts, EXCLUDED.max_accounts),
			updated_at = now()
		RETURNING id`, serial, label, adbIndex, maxAccounts).Scan(&id)
	return id, err
}

func (s *Store) GetAccountByID(ctx context.Context, id int64) (Account, error) {
	var a Account
	var paramsRaw []byte
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, email, password, profile_id,
			automation_flow, automation_params, automation_enabled
		FROM facebook_accounts WHERE id = $1 AND status = 'active'`, id).Scan(
		&a.ID, &a.Name, &a.Email, &a.Password, &a.ProfileID,
		&a.AutomationFlow, &paramsRaw, &a.AutomationEnabled)
	if err != nil {
		return Account{}, fmt.Errorf("account %d: %w", id, err)
	}
	a.AutomationParams = scanAutomationParams(paramsRaw)
	fps, err := s.listFanpages(ctx, id)
	if err != nil {
		return Account{}, err
	}
	a.Fanpages = fps
	return a, nil
}

func (s *Store) listFanpages(ctx context.Context, accountID int64) ([]Fanpage, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, fb_page_id, name FROM fanpages
		WHERE account_id = $1 AND status = 'active' ORDER BY id`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Fanpage
	for rows.Next() {
		var f Fanpage
		if err := rows.Scan(&f.ID, &f.FBPageID, &f.Name); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// GetNextAccountForDevice picks the next rotating slot (least recently used) for this HP.
func (s *Store) GetNextAccountForDevice(ctx context.Context, serial string) (Account, int64, error) {
	var slotID, accountID int64
	err := s.pool.QueryRow(ctx, `
		SELECT das.id, das.account_id
		FROM device_account_slots das
		JOIN devices d ON d.id = das.device_id
		WHERE d.serial = $1 AND das.active = true AND d.status = 'active'
		ORDER BY das.last_used_at NULLS FIRST, das.slot_no
		LIMIT 1`, serial).Scan(&slotID, &accountID)
	if err != nil {
		return Account{}, 0, fmt.Errorf("no account assigned to device %s: assign slots in database first", serial)
	}
	acc, err := s.GetAccountByID(ctx, accountID)
	if err != nil {
		return Account{}, 0, err
	}
	return acc, slotID, nil
}

func (s *Store) TouchAssignment(ctx context.Context, slotID int64) error {
	_, err := s.pool.Exec(ctx, `UPDATE device_account_slots SET last_used_at = now() WHERE id = $1`, slotID)
	return err
}

func (s *Store) StartRun(ctx context.Context, serial string, accountID int64, task string) (int64, error) {
	var deviceID int64
	if err := s.pool.QueryRow(ctx, `SELECT id FROM devices WHERE serial = $1`, serial).Scan(&deviceID); err != nil {
		return 0, err
	}
	var runID int64
	err := s.pool.QueryRow(ctx, `
		INSERT INTO runs (device_id, account_id, task, status)
		VALUES ($1, $2, $3, 'running') RETURNING id`, deviceID, accountID, task).Scan(&runID)
	return runID, err
}

func (s *Store) FinishRun(ctx context.Context, runID int64, status, errMsg string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE runs SET status = $2, error_message = $3, finished_at = now() WHERE id = $1`,
		runID, status, errMsg)
	return err
}

// AssignAccount creates or updates a device↔account slot mapping.
func (s *Store) AssignAccount(ctx context.Context, serial string, accountID int64, slotNo int) error {
	var deviceID int64
	if err := s.pool.QueryRow(ctx, `SELECT id FROM devices WHERE serial = $1`, serial).Scan(&deviceID); err != nil {
		return fmt.Errorf("device %s not found", serial)
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO device_account_slots (device_id, account_id, slot_no, active)
		VALUES ($1, $2, $3, true)
		ON CONFLICT (device_id, slot_no) DO UPDATE SET
			account_id = EXCLUDED.account_id,
			active = true`, deviceID, accountID, slotNo)
	return err
}

// AutoAssignDevices assigns account index i to device serials[i] at slot 1 (migration helper).
func (s *Store) AutoAssignDevices(ctx context.Context, serials []string, maxAccountsPerDevice int) (int, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id FROM facebook_accounts WHERE status = 'active' ORDER BY id`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var accountIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return 0, err
		}
		accountIDs = append(accountIDs, id)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	if len(accountIDs) == 0 {
		return 0, fmt.Errorf("no accounts to assign")
	}
	assigned := 0
	for i, serial := range serials {
		if _, err := s.UpsertDevice(ctx, serial, serial, i, maxAccountsPerDevice); err != nil {
			return assigned, err
		}
		if i >= len(accountIDs) {
			break
		}
		acc := accountIDs[i]
		if err := s.AssignAccount(ctx, serial, acc, 1); err != nil {
			return assigned, err
		}
		assigned++
	}
	return assigned, nil
}

// AutoAssignAllToDevice puts every active account on one device with slot 1..N (for 1 HP many accounts).
func (s *Store) AutoAssignAllToDevice(ctx context.Context, serial string, maxAccountsPerDevice int) (int, error) {
	if _, err := s.UpsertDevice(ctx, serial, serial, 0, maxAccountsPerDevice); err != nil {
		return 0, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id FROM facebook_accounts WHERE status = 'active' ORDER BY id LIMIT $1`, maxAccountsPerDevice)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	n := 0
	slot := 1
	for rows.Next() {
		var accID int64
		if err := rows.Scan(&accID); err != nil {
			return n, err
		}
		if err := s.AssignAccount(ctx, serial, accID, slot); err != nil {
			return n, err
		}
		slot++
		n++
	}
	return n, rows.Err()
}

func (s *Store) CountDevicesWithAssignments(ctx context.Context) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT d.id)
		FROM devices d
		JOIN device_account_slots das ON das.device_id = d.id
		WHERE das.active = true AND d.status = 'active'`).Scan(&n)
	return n, err
}

// DeviceAssignment maps a connected device serial to its primary assigned account (slot 1 or first active).
type DeviceAssignment struct {
	Serial    string
	AccountID int64
	Name      string
	LoginID   string
	SlotNo    int
}

func (s *Store) ListDeviceAssignments(ctx context.Context) (map[string]DeviceAssignment, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT ON (d.serial)
			d.serial, fa.id, fa.name, fa.email, fa.profile_id, das.slot_no
		FROM device_account_slots das
		JOIN devices d ON d.id = das.device_id
		JOIN facebook_accounts fa ON fa.id = das.account_id
		WHERE das.active = true AND d.status = 'active' AND fa.status = 'active'
		ORDER BY d.serial, das.slot_no, das.last_used_at NULLS FIRST`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]DeviceAssignment{}
	for rows.Next() {
		var a DeviceAssignment
		var email, profileID string
		if err := rows.Scan(&a.Serial, &a.AccountID, &a.Name, &email, &profileID, &a.SlotNo); err != nil {
			return nil, err
		}
		acc := Account{Email: email, ProfileID: profileID}
		a.LoginID = acc.LoginID()
		out[a.Serial] = a
	}
	return out, rows.Err()
}

func DefaultConnectTimeout() time.Duration {
	return 10 * time.Second
}
