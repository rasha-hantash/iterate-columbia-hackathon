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
    (1, 'CORN',        'Corn',        'crates'),
    (2, 'WHEAT',       'Wheat',       'bushels'),
    (3, 'SOYBEAN_OIL', 'Soybean Oil', 'lbs');

-- Positions (Alice @ Acme has exposure to all three)
INSERT INTO positions (client_id, user_id, commodity_id, volume, direction, entry_price) VALUES
    (1, 1, 1, 50000,  'long',  33.00),
    (1, 1, 2, 25000,  'long',  5.80),
    (1, 1, 3, 10000,  'short', 0.45),
    (1, 2, 1, 30000,  'long',  34.00);

-- Positions (Carol @ Global Grain)
INSERT INTO positions (client_id, user_id, commodity_id, volume, direction, entry_price) VALUES
    (2, 3, 2, 100000, 'long',  5.70);

-- Price data: 30 days for each commodity
-- (Using generate_series for realistic historical data)
INSERT INTO price_data (commodity_id, price, recorded_at)
SELECT
    1,
    ROUND((26.00 + random() * 16.00)::numeric, 4),
    d::date
FROM generate_series(CURRENT_DATE - INTERVAL '30 days', CURRENT_DATE, '1 day') AS d;

INSERT INTO price_data (commodity_id, price, recorded_at)
SELECT
    2,
    ROUND((5.50 + random() * 0.70)::numeric, 4),
    d::date
FROM generate_series(CURRENT_DATE - INTERVAL '30 days', CURRENT_DATE, '1 day') AS d;

INSERT INTO price_data (commodity_id, price, recorded_at)
SELECT
    3,
    ROUND((0.42 + random() * 0.06)::numeric, 4),
    d::date
FROM generate_series(CURRENT_DATE - INTERVAL '30 days', CURRENT_DATE, '1 day') AS d;

-- Existing alerts (so candidates can see real data)
INSERT INTO price_alerts (id, client_id, user_id, commodity_id, condition, threshold_price, status, notes) VALUES
    (1, 1, 1, 1, 'below', 28.00, 'active',    'Stop-loss on corn position'),
    (2, 1, 1, 2, 'above', 6.10, 'active',    'Take-profit on wheat'),
    (3, 1, 2, 1, 'below', 27.00, 'active',    'Bob watching corn dip'),
    (4, 2, 3, 2, 'below', 5.60, 'triggered', 'Wheat floor alert');

-- History for existing alerts
INSERT INTO alert_history (alert_id, changed_by_user_id, change_type, new_status, new_threshold, changed_at) VALUES
    (1, 1, 'created',   'active',    28.00, NOW() - INTERVAL '7 days'),
    (2, 1, 'created',   'active',    6.10, NOW() - INTERVAL '5 days'),
    (3, 2, 'created',   'active',    27.00, NOW() - INTERVAL '3 days'),
    (4, 3, 'created',   'active',    5.60, NOW() - INTERVAL '10 days'),
    (4, 3, 'triggered', 'triggered', NULL,  NOW() - INTERVAL '2 days');

-- Reset sequences to avoid ID conflicts
SELECT setval('clients_id_seq',    (SELECT MAX(id) FROM clients));
SELECT setval('users_id_seq',      (SELECT MAX(id) FROM users));
SELECT setval('commodities_id_seq',(SELECT MAX(id) FROM commodities));
SELECT setval('price_alerts_id_seq', (SELECT MAX(id) FROM price_alerts));
SELECT setval('alert_history_id_seq', (SELECT MAX(id) FROM alert_history));
