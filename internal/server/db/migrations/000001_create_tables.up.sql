CREATE TABLE IF NOT EXISTS users (
    username TEXT NOT NULL UNIQUE PRIMARY KEY,
    password BYTEA NOT NULL
);
CREATE TABLE IF NOT EXISTS secrets (
    username TEXT,
    key TEXT,
    data BYTEA,
    hash BYTEA,
    PRIMARY KEY (username, key)
);