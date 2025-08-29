package store

import (
	"context"
	"database/sql"
	"fmt"
	"proxytrack/types"
	"strings"

	_ "github.com/lib/pq"
)

type SessionStore interface {
	InsertSession(context.Context, *types.SessionRecord) error
	UpdateSession(context.Context, string, map[string]any) error
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(connStr string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{
		db: db,
	}, nil
}

func (p *PostgresStore) UpdateSession(ctx context.Context, id string, querySet map[string]any) error {
	setClauses := []string{}
	args := []any{}
	argPos := 1
	fmt.Println(id)
	for k, v := range querySet {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", k, argPos))
		args = append(args, v)
		argPos++
	}

	args = append(args, id)

	query := fmt.Sprintf(`
	Update session
	SET %s
	WHERE id=$%d
	`, strings.Join(setClauses, ", "), argPos)

	fmt.Println("-------", query)

	updSession := types.SessionRecord{}
	err := p.db.QueryRowContext(ctx, query, args...).Scan(
		&updSession.DurationMs)

	if err != nil {
		fmt.Println("no rows found")
		return sql.ErrNoRows
	}

	return nil
}

func (p *PostgresStore) InsertSession(ctx context.Context, session *types.SessionRecord) error {
	query := `INSERT INTO sessions (id, path, msg, request_time, response, response_time, status, error, duration_ms, created_at)
               VALUES ($1, $2, $3::jsonb, $4, $5::jsonb, $6, $7, $8, $9, $10)
		RETURNING id`

	insSession := &types.SessionRecord{}
	err := p.db.QueryRowContext(
		ctx,
		query,
		session.ID,
		session.Path,
		session.Msg,
		session.RequestTime,
		session.Response,
		session.ResponseTime,
		session.Status,
		session.Error,
		session.DurationMs,
		session.Created_at,
	).Scan(
		&insSession.ID,
	)

	return err
}

func (p *PostgresStore) createSessionTable() error {

	query := `CREATE EXTENSION IF NOT EXISTS pgcrypto;
	
		create table if not exists sessions (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  		msg           JSONB NOT NULL,
  		request_time  TIMESTAMPTZ NOT NULL,
  		response      JSONB,
  		response_time TIMESTAMPTZ,
  		status        INT,
  		error         TEXT,
  		duration_ms   BIGINT,
  		created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`

	_, err := p.db.Exec(query)
	if err != nil {
		return err
	}

	return err
}

func (p *PostgresStore) Init() error {
	return p.createSessionTable()
}
