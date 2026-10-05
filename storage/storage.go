package storage

import (
	"github.com/DNSGeek/filebrowser/v3/auth"
	"github.com/DNSGeek/filebrowser/v3/settings"
	"github.com/DNSGeek/filebrowser/v3/share"
	"github.com/DNSGeek/filebrowser/v3/token"
	"github.com/DNSGeek/filebrowser/v3/users"
)

// Storage is a storage powered by a Backend which makes the necessary
// verifications when fetching and saving data to ensure consistency.
type Storage struct {
	Users    users.Store
	Share    *share.Storage
	Auth     *auth.Storage
	Settings *settings.Storage
	Tokens   *token.Storage
}
