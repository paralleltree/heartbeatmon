package persistence

import "context"

type PersistentStore interface {
	Load(context.Context) ([]byte, error)
	Save(context.Context, []byte) error
}
