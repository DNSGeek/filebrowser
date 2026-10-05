package bolt

import (
	"github.com/asdine/storm/v3"

	"github.com/DNSGeek/filebrowser/v3/auth"
	"github.com/DNSGeek/filebrowser/v3/settings"
	"github.com/DNSGeek/filebrowser/v3/share"
	"github.com/DNSGeek/filebrowser/v3/storage"
	"github.com/DNSGeek/filebrowser/v3/token"
	"github.com/DNSGeek/filebrowser/v3/users"
)

// NewStorage creates a storage.Storage based on Bolt DB.
func NewStorage(db *storm.DB) (*storage.Storage, error) {
	userStore := users.NewStorage(usersBackend{db: db})
	shareStore := share.NewStorage(shareBackend{db: db})
	settingsStore := settings.NewStorage(settingsBackend{db: db})
	tokenStore := token.NewStorage(tokenBackend{db: db})
	authStore := auth.NewStorage(authBackend{db: db}, userStore)

	err := save(db, "version", 2)
	if err != nil {
		return nil, err
	}

	return &storage.Storage{
		Auth:     authStore,
		Users:    userStore,
		Share:    shareStore,
		Settings: settingsStore,
		Tokens:   tokenStore,
	}, nil
}
