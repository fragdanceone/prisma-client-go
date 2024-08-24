package engine

import (
	"context"
)

type Engine interface {
	Connect() error
	Disconnect() error
	Do(ctx context.Context, payload interface{}, into interface{}) error
	DoJson(ctx context.Context, payload interface{}) (err error, j []byte)
	GraphQL(ctx context.Context, payload string) (j []byte, err error)
	Batch(ctx context.Context, payload interface{}, into interface{}) error
	Name() string
}
