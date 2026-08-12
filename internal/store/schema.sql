CREATE TABLE IF NOT EXISTS devices (
    id BIGSERIAL PRIMARY KEY,
    serial TEXT NOT NULL UNIQUE,
    label TEXT NOT NULL DEFAULT '',
    adb_index INT NOT NULL DEFAULT 0,
    max_accounts INT NOT NULL DEFAULT 50,
    mirror_enabled BOOLEAN NOT NULL DEFAULT false,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE devices ADD COLUMN IF NOT EXISTS mirror_enabled BOOLEAN NOT NULL DEFAULT false;

CREATE TABLE IF NOT EXISTS facebook_accounts (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL DEFAULT '',
    email TEXT NOT NULL DEFAULT '',
    password TEXT NOT NULL,
    profile_id TEXT NOT NULL DEFAULT '',
    automation_flow TEXT NOT NULL DEFAULT 'facebook_login_logout',
    automation_params JSONB NOT NULL DEFAULT '{}',
    automation_enabled BOOLEAN NOT NULL DEFAULT true,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE facebook_accounts ADD COLUMN IF NOT EXISTS automation_flow TEXT NOT NULL DEFAULT 'facebook_login_logout';
ALTER TABLE facebook_accounts ADD COLUMN IF NOT EXISTS automation_params JSONB NOT NULL DEFAULT '{}';
ALTER TABLE facebook_accounts ADD COLUMN IF NOT EXISTS automation_enabled BOOLEAN NOT NULL DEFAULT true;

CREATE UNIQUE INDEX IF NOT EXISTS facebook_accounts_profile_id_uidx
    ON facebook_accounts (profile_id) WHERE profile_id <> '';

CREATE UNIQUE INDEX IF NOT EXISTS facebook_accounts_email_uidx
    ON facebook_accounts (email) WHERE email <> '';

CREATE TABLE IF NOT EXISTS fanpages (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGINT NOT NULL REFERENCES facebook_accounts(id) ON DELETE CASCADE,
    fb_page_id TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (account_id, fb_page_id)
);

CREATE TABLE IF NOT EXISTS device_account_slots (
    id BIGSERIAL PRIMARY KEY,
    device_id BIGINT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    account_id BIGINT NOT NULL REFERENCES facebook_accounts(id) ON DELETE CASCADE,
    slot_no INT NOT NULL CHECK (slot_no >= 1),
    active BOOLEAN NOT NULL DEFAULT true,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (device_id, slot_no),
    UNIQUE (device_id, account_id)
);

CREATE INDEX IF NOT EXISTS device_account_slots_device_active_idx
    ON device_account_slots (device_id, active, last_used_at NULLS FIRST);

CREATE TABLE IF NOT EXISTS runs (
    id BIGSERIAL PRIMARY KEY,
    device_id BIGINT NOT NULL REFERENCES devices(id),
    account_id BIGINT NOT NULL REFERENCES facebook_accounts(id),
    fanpage_id BIGINT REFERENCES fanpages(id),
    task TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'running',
    error_message TEXT NOT NULL DEFAULT '',
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS runs_device_started_idx ON runs (device_id, started_at DESC);

CREATE TABLE IF NOT EXISTS post_texts (
    id BIGSERIAL PRIMARY KEY,
    category TEXT NOT NULL CHECK (category IN ('personal', 'fanpage', 'group')),
    body TEXT NOT NULL,
    image_file TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS post_texts_category_active_idx
    ON post_texts (category, id DESC) WHERE status = 'active';
