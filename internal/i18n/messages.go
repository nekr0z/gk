package i18n

import "github.com/nicksnyder/go-i18n/v2/i18n"

var messages = []*i18n.Message{
	{
		ID:    "gk.rootcmd.short",
		Other: "GophKeeper password manager",
	},
	{
		ID:    "gk.rootcmd.long",
		Other: "A password manager written in Go.",
	},
	{
		ID:    "gk.rootcmd.flags.db",
		Other: "database file (default is gk.sqlite in current directory)",
	},
	{
		ID:    "gk.rootcmd.flags.passphrase",
		Other: "passphrase for encryption",
	},
	{
		ID:    "gk.rootcmd.flags.server",
		Other: "server address",
	},
	{
		ID:    "gk.rootcmd.flags.username",
		Other: "user name",
	},
	{
		ID:    "gk.rootcmd.flags.password",
		Other: "password",
	},
	{
		ID:    "gk.rootcmd.flags.insecure",
		Other: "disable TLS verification",
	},
	{
		ID:    "gk.rootcmd.flags.prefer",
		Other: "`remote` or `local`",
	},
	{
		ID:    "gk.rootcmd.flags.config",
		Other: "config file (if not set, will look for .gk.yaml in the home directory)",
	},
	{
		ID:    "gk.create.short",
		Other: "Create a new secret",
	},
	{
		ID:    "gk.create.flags.metadata",
		Other: "metadata for the secret (key=value), multiple can be provided",
	},
	{
		ID:    "gk.create.text.use",
		Other: "text <name> <value>",
	},
	{
		ID:    "gk.create.text.short",
		Other: "Create a new text secret",
	},
	{
		ID:    "gk.create.binary.use",
		Other: "binary <name> <filename>",
	},
	{
		ID:    "gk.create.binary.short",
		Other: "Create a new binary secret from file",
	},
	{
		ID:    "gk.create.password.use",
		Other: "password <name> <username> <password>",
	},
	{
		ID:    "gk.create.password.short",
		Other: "Create a new password secret",
	},
	{
		ID:    "gk.create.card.use",
		Other: "card <name> <number> <expiry> <cvv> [<username>]",
	},
	{
		ID:    "gk.create.card.short",
		Other: "Create a new card secret",
	},
	{
		ID:    "gk.delete.use",
		Other: "delete <name>",
	},
	{
		ID:    "gk.delete.short",
		Other: "Delete a secret",
	},
	{
		ID:    "gk.show.use",
		Other: "show <name>",
	},
	{
		ID:    "gk.show.short",
		Other: "Show the secret",
	},
	{
		ID:    "gk.show.flags.target-file",
		Other: "file to save the secret content to (otherwise will only print to stdout)",
	},
	{
		ID:    "gk.signup.short",
		Other: "Sign up for a new account",
	},
	{
		ID:    "gk.signup.long",
		Other: "Sign up for a new account on the configured server using the configured credentials.",
	},
	{
		ID:    "gk.signup.signing",
		Other: "Signing up...",
	},
	{
		ID:    "gk.signup.success",
		Other: "Signup with username {{.Username}} successful!",
	},
	{
		ID:    "gk.sync.short",
		Other: "Sync secrets with the server",
	},
}
