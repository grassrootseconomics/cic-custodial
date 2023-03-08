-- Keystore table
CREATE TABLE IF NOT EXISTS keystore (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    public_key TEXT NOT NULL,
    private_key TEXT NOT NULL,
    active BOOLEAN DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS public_key_idx ON keystore(public_key);