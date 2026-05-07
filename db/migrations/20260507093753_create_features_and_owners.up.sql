CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE owners (
    id SERIAL PRIMARY KEY,
    external_id INTEGER NOT NULL UNIQUE,
    login VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE features (
    id SERIAL PRIMARY KEY,
    layer_id INTEGER NOT NULL,
    owner_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50),
    

    geometry GEOMETRY(Geometry, 4326) NOT NULL,
    
    properties JSONB,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_features_layer FOREIGN KEY (layer_id) REFERENCES layers(id) ON DELETE CASCADE,
    CONSTRAINT fk_features_owner FOREIGN KEY (owner_id) REFERENCES owners(id) ON DELETE CASCADE
);

CREATE INDEX idx_features_geometry ON features USING GIST (geometry);

CREATE TRIGGER update_features_updated_at
    BEFORE UPDATE ON features
    FOR EACH ROW
    EXECUTE PROCEDURE update_updated_at_column();