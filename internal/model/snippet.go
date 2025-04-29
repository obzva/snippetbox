package model

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

type SnippetModel struct {
	DBPool *pgxpool.Pool
}

func (sm *SnippetModel) Insert(ctx context.Context, title string, content string, expires int) (int, error) {
	stmt := `INSERT INTO snippet (title, content, created, expires)
	VALUES($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP + MAKE_INTERVAL(days => $3))
	RETURNING id`

	var id int
	if err := sm.DBPool.QueryRow(ctx, stmt, title, content, expires).Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (sm *SnippetModel) Get(ctx context.Context, id int) (Snippet, error) {
	stmt := `SELECT *
	FROM snippet
	WHERE
		expires > CURRENT_TIMESTAMP
		AND id = $1`

	rows, err := sm.DBPool.Query(ctx, stmt, id)
	if err != nil {
		return Snippet{}, err
	}

	s, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Snippet])
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Snippet{}, ErrNoRecord
		} else {
			return Snippet{}, err
		}
	}

	return s, nil
}

// return the 10 most recently created snippets
func (sm *SnippetModel) Latest(ctx context.Context) ([]Snippet, error) {
	stmt := `SELECT *
	FROM snippet
	WHERE expires > CURRENT_TIMESTAMP
	ORDER BY created DESC
	LIMIT 10`

	rows, err := sm.DBPool.Query(ctx, stmt)
	if err != nil {
		return nil, err
	}

	s, err := pgx.CollectRows(rows, pgx.RowToStructByName[Snippet])
	if err != nil {
		return nil, err
	}

	return s, nil
}
