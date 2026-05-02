CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';


CREATE TRIGGER update_layers_updated_at
    BEFORE UPDATE ON layers
    FOR EACH ROW
    EXECUTE PROCEDURE update_updated_at_column();