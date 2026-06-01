package pgvector

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/azcov/gokit/vector"
)

var _ vector.Store = (*PgVector)(nil)

// PgVector stores and queries embedding vectors using the pgvector Postgres extension.
// Prerequisite: CREATE EXTENSION IF NOT EXISTS vector;
//
// The table schema expected:
//
//	CREATE TABLE IF NOT EXISTS {namespace} (
//	    id TEXT PRIMARY KEY,
//	    embedding vector({dimension}),
//	    metadata JSONB
//	);
type PgVector struct {
	pool *pgxpool.Pool
	cfg  vector.Config
}

func New(pool *pgxpool.Pool, cfg vector.Config) *PgVector {
	if cfg.Namespace == "" {
		cfg.Namespace = "embeddings"
	}
	return &PgVector{pool: pool, cfg: cfg}
}

// CreateTable creates the vector table if it does not exist.
func (p *PgVector) CreateTable(ctx context.Context) error {
	_, err := p.pool.Exec(ctx, fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id       TEXT    PRIMARY KEY,
			embedding vector(%d),
			metadata JSONB   NOT NULL DEFAULT '{}'
		)`, p.cfg.Namespace, p.cfg.Dimension))
	if err != nil {
		return fmt.Errorf("pgvector: create table: %w", err)
	}
	_, err = p.pool.Exec(ctx, fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS %s_embedding_idx
		ON %s USING ivfflat (embedding vector_cosine_ops)`,
		p.cfg.Namespace, p.cfg.Namespace))
	if err != nil {
		// Index creation failing is non-fatal (requires enough rows).
		_ = err
	}
	return nil
}

func (p *PgVector) Upsert(ctx context.Context, records []vector.Record) error {
	for _, r := range records {
		ns := r.Namespace
		if ns == "" {
			ns = p.cfg.Namespace
		}
		_, err := p.pool.Exec(ctx,
			fmt.Sprintf(`INSERT INTO %s (id, embedding, metadata)
				VALUES ($1, $2::vector, $3::jsonb)
				ON CONFLICT (id) DO UPDATE
				SET embedding = EXCLUDED.embedding,
				    metadata  = EXCLUDED.metadata`, ns),
			r.ID, vectorToString(r.Vector), metadataToJSON(r.Metadata),
		)
		if err != nil {
			return fmt.Errorf("pgvector: upsert %s: %w", r.ID, err)
		}
	}
	return nil
}

func (p *PgVector) Query(ctx context.Context, vec vector.Vector, opts vector.QueryOptions) ([]vector.QueryResult, error) {
	ns := opts.Namespace
	if ns == "" {
		ns = p.cfg.Namespace
	}
	topK := opts.TopK
	if topK <= 0 {
		topK = 10
	}

	rows, err := p.pool.Query(ctx,
		fmt.Sprintf(`SELECT id, metadata::text, 1 - (embedding <=> $1::vector) AS score
			FROM %s
			ORDER BY embedding <=> $1::vector
			LIMIT $2`, ns),
		vectorToString(vec), topK,
	)
	if err != nil {
		return nil, fmt.Errorf("pgvector: query: %w", err)
	}
	defer rows.Close()

	var results []vector.QueryResult
	for rows.Next() {
		var id, metaJSON string
		var score float64
		if err := rows.Scan(&id, &metaJSON, &score); err != nil {
			return nil, fmt.Errorf("pgvector: scan: %w", err)
		}
		results = append(results, vector.QueryResult{
			ID:       id,
			Score:    score,
			Metadata: jsonToMetadata(metaJSON),
		})
	}
	return results, rows.Err()
}

func (p *PgVector) Delete(ctx context.Context, ids []string, namespace string) error {
	if len(ids) == 0 {
		return nil
	}
	ns := namespace
	if ns == "" {
		ns = p.cfg.Namespace
	}
	_, err := p.pool.Exec(ctx,
		fmt.Sprintf(`DELETE FROM %s WHERE id = ANY($1)`, ns),
		ids,
	)
	if err != nil {
		return fmt.Errorf("pgvector: delete: %w", err)
	}
	return nil
}

func (p *PgVector) Close() error {
	p.pool.Close()
	return nil
}

// vectorToString converts []float64 to Postgres vector literal: [0.1,0.2,0.3]
func vectorToString(v vector.Vector) string {
	parts := make([]string, len(v))
	for i, f := range v {
		parts[i] = fmt.Sprintf("%g", f)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func metadataToJSON(m map[string]any) string {
	if len(m) == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteByte('{')
	first := true
	for k, v := range m {
		if !first {
			sb.WriteByte(',')
		}
		first = false
		fmt.Fprintf(&sb, `"%s":"%v"`, k, v)
	}
	sb.WriteByte('}')
	return sb.String()
}

func jsonToMetadata(s string) map[string]any {
	// minimal parse — for production use encoding/json
	return map[string]any{"_raw": s}
}
