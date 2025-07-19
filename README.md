# Gophkeeper

A secret/password manager written in Go.

Supported secrets are:

- password: `username` and `password` pair;
- text: UTF-8 string;
- binary: any sequence of bytes;
- card: `card number`, `expiration date`, `security code` and, optionally, `cardholder name`.

Each secret can hold arbitrary metadata (key-value UTF-8 string pairs).

## Manager

Stores secrets to a local SQLite3 database file. Secrets are stored encrypted.

Secrets can be synchronized with a remote server. The secrets themselves, and everything in them, are encrypted locally and never decrypted on the server; however, the names of the secrets are stored in plain text.

Use the same (preferably strong) passphrase on all the clients you intend to synchronize. The passphrase to read a secret should be the same that was used to create it.

### Configuration

Configuration is done via a config file; any setting can be overridden by environment variables and command line flags. The default config file location in `.gk.yaml` in the home directory, it can also be overriden using `-c` or `--config` flags or `GK_CONFIG` environment variable.

```yaml
db: "gk.sqlite" # defaults to gk.sqlite in the current directory, override with `-d`, `--database` or `GK_DB` environment variable
passhprase: "my secret passphrase" # not recommended to be stored in the config file, override with `-p`, `--passphrase` or `GK_PASSPHRASE` environment variable

server: # optional server configuration
  address: "localhost:8080" # server address, override with `-s`, `--server` or `GK_SERVER` environment variable
  insecure: false # enable insecure mode (not recommended, use only for testing), override with `-i`, `--insecure` or `GK_INSECURE` environment variable
  username: "user" # username on server, override with `-u`, `--username` or `GK_USERNAME` environment variable
  password: "password" # password on server, not recommended to be stored in the config file, override with `-p`, `--password` or `GK_PASSWORD` environment variable

prefer: # "local" or "remote" in case of conflict, override with `-g`, `--prefer` or `GK_PREFER` environment variable
```

### Usage

Create a password secret named "mysecret" with username "user@example.com", password "monkey123" and some helpful metadata:
```
gk create password mysecret "user@example.com" "monkey123" -m "url=https://example.com" -m "description=My password for example.com"
```

Show a secret named "mysecret":
```
gk show mysecret
```

Save a secret to a file:
```
gk show mysecret -t mysecret.txt
```

Delete a secret:
```
gk delete mysecret
```

Sign up on a server (this will create a new user on the server):
```
gk signup -u user -w password -s server:8080
```

Synchronize all secrets with a server:
```
gk sync
```

## Server

Allows users to sign up and to synchronize secrets between clients.

### Configuration

Configuration is done via a config file; any setting can be overridden by environment variables and command line flags. The default config file location in `.gk.yaml` in the home directory, it can also be overriden using `-c` or `--config` flags or `GK_CONFIG` environment variable.

```yaml
dsn: "" # DSN of the PostgreSQL database, override with `-d`, `--dsn` or `GK_SERVER_DSN` environment variable
key: "" # JWT signing key (random one will be created if not provided), not recommended to be stored in the config file, override with `-k`, `--key` or `GK_SERVER_KEY` environment variable
address: "" # server listening address, override with `-a`, `--address` or `GK_SERVER_ADDRESS` environment variable
```

### Usage

Start the server:
```
gk server -d postgres://user:password@localhost/gk -k secret -a :8080
```

Users can sign up on the server using the `gk signup` command and synchronize secrets from a client using the `gk sync` command.
