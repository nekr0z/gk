CREATE TABLE IF NOT EXISTS secrets (
    id TEXT PRIMARY KEY,
    encrypted_payload BLOB,
    payload_hash BLOB,
    server_hash BLOB
);