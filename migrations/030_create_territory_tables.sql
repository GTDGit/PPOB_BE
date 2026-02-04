-- Migration: 030_create_territory_tables
-- Description: Tables for Indonesia territory data (provinces, cities, districts, sub-districts, postal codes)

-- Provinces (38 records)
CREATE TABLE IF NOT EXISTS provinces (
    code VARCHAR(2) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_provinces_name ON provinces(name);

COMMENT ON TABLE provinces IS 'Indonesia provinces data';
COMMENT ON COLUMN provinces.code IS 'Province code (2 digits)';
COMMENT ON COLUMN provinces.name IS 'Province name';

-- Cities/Regencies (514 records)
CREATE TABLE IF NOT EXISTS cities (
    code VARCHAR(4) PRIMARY KEY,
    province_code VARCHAR(2) NOT NULL,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_cities_province
        FOREIGN KEY (province_code)
        REFERENCES provinces(code)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_cities_province ON cities(province_code);
CREATE INDEX IF NOT EXISTS idx_cities_name ON cities(name);

COMMENT ON TABLE cities IS 'Indonesia cities/regencies data';
COMMENT ON COLUMN cities.code IS 'City code (4 digits)';
COMMENT ON COLUMN cities.province_code IS 'Reference to province';

-- Districts (7,266 records)
CREATE TABLE IF NOT EXISTS districts (
    code VARCHAR(6) PRIMARY KEY,
    city_code VARCHAR(4) NOT NULL,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_districts_city
        FOREIGN KEY (city_code)
        REFERENCES cities(code)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_districts_city ON districts(city_code);
CREATE INDEX IF NOT EXISTS idx_districts_name ON districts(name);

COMMENT ON TABLE districts IS 'Indonesia districts data';
COMMENT ON COLUMN districts.code IS 'District code (6 digits)';
COMMENT ON COLUMN districts.city_code IS 'Reference to city';

-- Sub-Districts/Villages (83,436 records)
CREATE TABLE IF NOT EXISTS sub_districts (
    code VARCHAR(10) PRIMARY KEY,
    district_code VARCHAR(6) NOT NULL,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_sub_districts_district
        FOREIGN KEY (district_code)
        REFERENCES districts(code)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_sub_districts_district ON sub_districts(district_code);
CREATE INDEX IF NOT EXISTS idx_sub_districts_name ON sub_districts(name);

COMMENT ON TABLE sub_districts IS 'Indonesia sub-districts/villages data';
COMMENT ON COLUMN sub_districts.code IS 'Sub-district code (10 digits)';
COMMENT ON COLUMN sub_districts.district_code IS 'Reference to district';

-- Postal Codes
CREATE TABLE IF NOT EXISTS postal_codes (
    id SERIAL PRIMARY KEY,
    sub_district_code VARCHAR(10) NOT NULL,
    postal_code VARCHAR(5) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_postal_codes_sub_district
        FOREIGN KEY (sub_district_code)
        REFERENCES sub_districts(code)
        ON DELETE CASCADE,
    CONSTRAINT uq_postal_code_sub_district
        UNIQUE (sub_district_code, postal_code)
);

CREATE INDEX IF NOT EXISTS idx_postal_codes_sub_district ON postal_codes(sub_district_code);
CREATE INDEX IF NOT EXISTS idx_postal_codes_postal_code ON postal_codes(postal_code);

COMMENT ON TABLE postal_codes IS 'Indonesia postal codes data';
COMMENT ON COLUMN postal_codes.sub_district_code IS 'Reference to sub-district';
COMMENT ON COLUMN postal_codes.postal_code IS 'Postal code (5 digits)';

-- Sync metadata
CREATE TABLE IF NOT EXISTS territory_sync_log (
    id SERIAL PRIMARY KEY,
    sync_type VARCHAR(20) NOT NULL,
    total_records INT NOT NULL,
    status VARCHAR(20) NOT NULL,
    error_message TEXT,
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_territory_sync_log_sync_type ON territory_sync_log(sync_type);
CREATE INDEX IF NOT EXISTS idx_territory_sync_log_status ON territory_sync_log(status);

COMMENT ON TABLE territory_sync_log IS 'Log for territory data synchronization';
COMMENT ON COLUMN territory_sync_log.sync_type IS 'Type of sync: provinces, cities, districts, sub_districts, postal_codes';
COMMENT ON COLUMN territory_sync_log.status IS 'Sync status: running, success, failed';
