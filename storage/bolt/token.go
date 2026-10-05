package bolt

import (
	"errors"

	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
)

type revokedToken struct {
	ID      string `storm:"id"`
	Expires int64  `storm:"index"`
}

type tokenBackend struct {
	db *storm.DB
}

func (t tokenBackend) Add(id string, expires int64) error {
	return t.db.Save(&revokedToken{ID: id, Expires: expires})
}

func (t tokenBackend) Has(id string) (bool, error) {
	var v revokedToken
	err := t.db.One("ID", id, &v)
	if errors.Is(err, storm.ErrNotFound) {
		return false, nil
	}
	return err == nil, err
}

func (t tokenBackend) Prune(before int64) error {
	err := t.db.Select(q.Lt("Expires", before)).Delete(new(revokedToken))
	if errors.Is(err, storm.ErrNotFound) {
		return nil
	}
	return err
}
