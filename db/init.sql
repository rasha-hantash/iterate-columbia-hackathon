-- =============================================================================
-- Edge Technical Interview: Database Schema & Seed Data
-- =============================================================================
-- This script creates all tables and seeds realistic test data.
-- The database is ready to use — no migrations needed.
-- =============================================================================

-- ---------------------------------------------------------------------------
-- SCHEMA
-- ---------------------------------------------------------------------------

CREATE TABLE clients (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE users (
    id          SERIAL PRIMARY KEY,
    client_id   INTEGER      NOT NULL REFERENCES clients(id),
    name        VARCHAR(255) NOT NULL,
    email       VARCHAR(255) NOT NULL UNIQUE,
    role        VARCHAR(50)  NOT NULL DEFAULT 'member',
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE commodities (
    id          SERIAL PRIMARY KEY,
    code        VARCHAR(50)  NOT NULL UNIQUE,
    name        VARCHAR(255) NOT NULL,
    unit        VARCHAR(50)  NOT NULL
);

CREATE TABLE positions (
    id          SERIAL PRIMARY KEY,
    client_id   INTEGER        NOT NULL REFERENCES clients(id),
    user_id     INTEGER        NOT NULL REFERENCES users(id),
    commodity_id INTEGER       NOT NULL REFERENCES commodities(id),
    volume      DECIMAL(15,4)  NOT NULL,
    direction   VARCHAR(10)    NOT NULL CHECK (direction IN ('long', 'short')),
    entry_price DECIMAL(15,4)  NOT NULL,
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE TABLE price_data (
    id           SERIAL PRIMARY KEY,
    commodity_id INTEGER       NOT NULL REFERENCES commodities(id),
    price        DECIMAL(15,4) NOT NULL,
    recorded_at  DATE          NOT NULL,
    UNIQUE (commodity_id, recorded_at)
);

CREATE TABLE price_alerts (
    id                SERIAL PRIMARY KEY,
    client_id         INTEGER        NOT NULL REFERENCES clients(id),
    user_id           INTEGER        NOT NULL REFERENCES users(id),
    commodity_id      INTEGER        NOT NULL REFERENCES commodities(id),
    condition         VARCHAR(10)    NOT NULL CHECK (condition IN ('above', 'below')),
    threshold_price   DECIMAL(15,4)  NOT NULL CHECK (threshold_price > 0),
    status            VARCHAR(20)    NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'triggered', 'paused')),
    notes             TEXT           NOT NULL DEFAULT '',
    triggered_count   INTEGER        NOT NULL DEFAULT 0,
    last_triggered_at TIMESTAMPTZ,
    created_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ
);

CREATE INDEX idx_alerts_client_status     ON price_alerts (client_id, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_alerts_user_status       ON price_alerts (user_id, status)   WHERE deleted_at IS NULL;
CREATE INDEX idx_alerts_client_commodity  ON price_alerts (client_id, commodity_id) WHERE deleted_at IS NULL;

CREATE TABLE alert_history (
    id                  SERIAL PRIMARY KEY,
    alert_id            INTEGER        NOT NULL REFERENCES price_alerts(id),
    changed_by_user_id  INTEGER        REFERENCES users(id),
    change_type         VARCHAR(20)    NOT NULL,  -- 'created', 'updated', 'triggered', 'deleted'
    previous_status     VARCHAR(20),
    new_status          VARCHAR(20),
    previous_threshold  DECIMAL(15,4),
    new_threshold       DECIMAL(15,4),
    metadata            JSONB,
    changed_at          TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_history_alert ON alert_history (alert_id, changed_at DESC);


-- ---------------------------------------------------------------------------
-- SEED DATA
-- ---------------------------------------------------------------------------

-- Clients
INSERT INTO clients (id, name) VALUES
    (1, 'Acme Foods'),
    (2, 'Global Grain Co');

-- Users
INSERT INTO users (id, client_id, name, email, role) VALUES
    (1, 1, 'Alice Smith',   'alice@acme.com',   'risk_manager'),
    (2, 1, 'Bob Jones',     'bob@acme.com',     'procurement'),
    (3, 2, 'Carol Chen',    'carol@globalgrain.com', 'risk_manager');

-- Commodities
INSERT INTO commodities (id, code, name, unit) VALUES
    (1, 'CORN', 'Corn', '4 dozen');

-- Positions (Alice @ Acme has long and short corn exposure)
INSERT INTO positions (client_id, user_id, commodity_id, volume, direction, entry_price) VALUES
    (1, 1, 1, 50000,  'long',  33.00),
    (1, 1, 1, 20000,  'short', 38.00),
    (1, 2, 1, 30000,  'long',  34.00);

-- Positions (Carol @ Global Grain)
INSERT INTO positions (client_id, user_id, commodity_id, volume, direction, entry_price) VALUES
    (2, 3, 1, 100000, 'long',  32.00);

-- Price data: 30 days for corn
-- (Using generate_series for realistic historical data)
INSERT INTO price_data (commodity_id, price, recorded_at)
SELECT
    1,
    ROUND((26.00 + random() * 16.00)::numeric, 4),
    d::date
FROM generate_series(CURRENT_DATE - INTERVAL '30 days', CURRENT_DATE, '1 day') AS d;

-- Reset sequences to avoid ID conflicts
SELECT setval('clients_id_seq',    (SELECT MAX(id) FROM clients));
SELECT setval('users_id_seq',      (SELECT MAX(id) FROM users));
SELECT setval('commodities_id_seq',(SELECT MAX(id) FROM commodities));
