// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: query.sql

package nhost

import (
	"context"

	"github.com/google/uuid"
)

const getAppDesiredState = `-- name: GetAppDesiredState :one
SELECT desired_state FROM apps WHERE id = $1
`

func (q *Queries) GetAppDesiredState(ctx context.Context, id uuid.UUID) (int32, error) {
	row := q.db.QueryRow(ctx, getAppDesiredState, id)
	var desired_state int32
	err := row.Scan(&desired_state)
	return desired_state, err
}
