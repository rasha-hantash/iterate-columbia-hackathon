-- Market data table for USDA terminal market pricing
CREATE TABLE IF NOT EXISTS market_data (
    id                  SERIAL PRIMARY KEY,
    report_date         DATE NOT NULL,
    location            VARCHAR(255) NOT NULL,
    commodity           VARCHAR(255) NOT NULL,
    variety             VARCHAR(255),
    package             VARCHAR(255),
    origin              VARCHAR(255),
    item_size           VARCHAR(255),
    low_price           DECIMAL(10,2),
    high_price          DECIMAL(10,2),
    mostly_low_price    DECIMAL(10,2),
    mostly_high_price   DECIMAL(10,2),
    properties          VARCHAR(255),
    comment             TEXT
);

CREATE INDEX IF NOT EXISTS idx_market_data_location ON market_data (location);
CREATE INDEX IF NOT EXISTS idx_market_data_date ON market_data (report_date);
CREATE INDEX IF NOT EXISTS idx_market_data_commodity ON market_data (commodity);
