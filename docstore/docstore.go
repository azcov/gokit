package docstore

import "context"

const (
	OpEq  = "=="
	OpNe  = "!="
	OpLt  = "<"
	OpLte = "<="
	OpGt  = ">"
	OpGte = ">="
)

type Filter struct {
	Field string
	Op    string
	Value any
}

type Query struct {
	Filters []Filter
	OrderBy string
	Desc    bool
	Limit   int
}

// Store is a generic document store interface.
// Each collection maps to a table/collection in the backing store.
// id is the document identifier; doc must be a struct or map.
type Store interface {
	Get(ctx context.Context, collection, id string, dest any) error
	Set(ctx context.Context, collection, id string, doc any) error
	Update(ctx context.Context, collection, id string, fields map[string]any) error
	Delete(ctx context.Context, collection, id string) error
	List(ctx context.Context, collection string, q Query) ([]map[string]any, error)
	Close() error
}
