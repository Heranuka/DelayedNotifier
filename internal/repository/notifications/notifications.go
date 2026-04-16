// Package comments provides access to notification persistence operations.
package comments

import (
	"context"
	"delay/internal/domain"
	"fmt"
	"time"

	"github.com/google/uuid"
	pgxdriver "github.com/wb-go/wbf/dbpg/pgx-driver"
)

// Repository provides notification storage operations backed by PostgreSQL.
type Repository struct {
	pool *pgxdriver.Postgres
}

// New creates a new Repository instance.
func New(pool *pgxdriver.Postgres) *Repository {
	return &Repository{pool: pool}
}

// Create inserts a new notification into the database and returns its ID.
func (r *Repository) Create(ctx context.Context, notification *domain.Notification) (uuid.UUID, error) {
	query := `
		INSERT INTO notifications(id, message, destination, channel, data_sent_at, status)
		VALUES($1, $2, $3, $4, $5, $6)
	`

	id := uuid.New()
	status := "created"

	_, err := r.pool.Exec(
		ctx,
		query,
		id,
		notification.Message,
		notification.Destination,
		notification.Channel,
		notification.DataToSent,
		status,
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create: %w", err)
	}

	return id, nil
}

// Status returns the current notification status by ID.
func (r *Repository) Status(ctx context.Context, id uuid.UUID) (string, error) {
	query := `SELECT status FROM notifications WHERE id = $1`

	var status string
	err := r.pool.QueryRow(ctx, query, id).Scan(&status)
	if err != nil {
		if err.Error() == "no rows in result set" || err.Error() == "pgx: no rows in result set" {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("failed to get current status: %w", err)
	}

	return status, nil
}

// Cancel deletes a notification by ID.
func (r *Repository) Cancel(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM notifications WHERE id = $1`

	res, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to cancel: %w", err)
	}

	if res.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
func (r *Repository) GetAll(ctx context.Context) (*[]domain.Notification, error) {
	query := `SELECT id, message, destination, channel, status, data_sent_at, created_at FROM notifications`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}
	defer rows.Close()

	var notifications []domain.Notification

	for rows.Next() {
		var (
			n          domain.Notification
			dataSentAt *time.Time
		)

		if err := rows.Scan(
			&n.ID,
			&n.Message,
			&n.Destination,
			&n.Channel,
			&n.Status,
			&dataSentAt,
			&n.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		if dataSentAt != nil {
			n.DataToSent = *dataSentAt
		}

		notifications = append(notifications, n)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to rows: %w", err)
	}

	return &notifications, nil
}

// UpdateStatus updates the notification status by ID.
func (r *Repository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE notifications SET status = $1 WHERE id = $2`

	_, err := r.pool.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

// Get returns a notification by ID.
func (r *Repository) Get(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	query := `SELECT id, message, destination, channel, status, data_sent_at, created_at FROM notifications WHERE id = $1`

	var note domain.Notification
	if err := r.pool.QueryRow(ctx, query, id).Scan(
		&note.ID,
		&note.Message,
		&note.Destination,
		&note.Channel,
		&note.Status,
		&note.DataToSent,
		&note.CreatedAt,
	); err != nil {
		return nil, fmt.Errorf("failed to get notification by ID: %w", err)
	}

	return &note, nil
}
