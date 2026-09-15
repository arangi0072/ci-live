package chat

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//
// PostgreSQL Repository
//

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{
		db: db,
	}
}

//
// Create Message
//

func (r *repository) CreateMessage(
	ctx context.Context,
	message *ChatMessage,
) (*ChatMessage, error) {

	const query = `
		INSERT INTO chat_messages (
			stream_id,
			user_id,
			message
		)
		VALUES ($1, $2, $3)
		RETURNING
			id,
			stream_id,
			user_id,
			message,
			is_deleted,
			deleted_at,
			created_at
	`

	var result ChatMessage

	err := r.db.QueryRow(
		ctx,
		query,
		message.StreamID,
		message.UserID,
		message.Message,
	).Scan(
		&result.ID,
		&result.StreamID,
		&result.UserID,
		&result.Message,
		&result.IsDeleted,
		&result.DeletedAt,
		&result.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

//
// Get Message By ID
//

func (r *repository) GetMessageByID(
	ctx context.Context,
	messageID int64,
) (*ChatMessage, error) {

	const query = `
		SELECT
			id,
			stream_id,
			user_id,
			message,
			is_deleted,
			deleted_at,
			created_at
		FROM chat_messages
		WHERE id = $1
	`

	var message ChatMessage

	err := r.db.QueryRow(
		ctx,
		query,
		messageID,
	).Scan(
		&message.ID,
		&message.StreamID,
		&message.UserID,
		&message.Message,
		&message.IsDeleted,
		&message.DeletedAt,
		&message.CreatedAt,
	)

	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMessageNotFound
		}

		return nil, err
	}

	return &message, nil
}

//
// List Messages
//

func (r *repository) ListMessages(
	ctx context.Context,
	filter MessageFilter,
) ([]ChatMessage, int64, error) {

	conditions := []string{
		"stream_id = $1",
	}

	args := []any{
		filter.StreamID,
	}

	parameterIndex := 2

	//
	// User filter
	//

	if filter.UserID != nil {
		conditions = append(
			conditions,
			fmt.Sprintf(
				"user_id = $%d",
				parameterIndex,
			),
		)

		args = append(
			args,
			*filter.UserID,
		)

		parameterIndex++
	}

	//
	// Hide deleted messages by default.
	//

	if !filter.IncludeDeleted {
		conditions = append(
			conditions,
			"is_deleted = FALSE",
		)
	}

	whereClause := strings.Join(
		conditions,
		" AND ",
	)

	//
	// Count query
	//

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM chat_messages
		WHERE %s
	`, whereClause)

	var total int64

	err := r.db.QueryRow(
		ctx,
		countQuery,
		args...,
	).Scan(&total)

	if err != nil {
		return nil, 0, err
	}

	//
	// Pagination
	//

	page := filter.Page

	if page < 1 {
		page = 1
	}

	limit := filter.Limit

	if limit < 1 {
		limit = 50
	}

	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	//
	// Message query
	//

	messageQuery := fmt.Sprintf(`
		SELECT
			id,
			stream_id,
			user_id,
			message,
			is_deleted,
			deleted_at,
			created_at
		FROM chat_messages
		WHERE %s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d
		OFFSET $%d
	`,
		whereClause,
		parameterIndex,
		parameterIndex+1,
	)

	args = append(
		args,
		limit,
		offset,
	)

	rows, err := r.db.Query(
		ctx,
		messageQuery,
		args...,
	)

	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	messages := make(
		[]ChatMessage,
		0,
		limit,
	)

	for rows.Next() {

		var message ChatMessage

		err := rows.Scan(
			&message.ID,
			&message.StreamID,
			&message.UserID,
			&message.Message,
			&message.IsDeleted,
			&message.DeletedAt,
			&message.CreatedAt,
		)

		if err != nil {
			return nil, 0, err
		}

		messages = append(
			messages,
			message,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

//
// Delete Message
//
// Soft delete:
// is_deleted = TRUE
// deleted_at = NOW()
//

func (r *repository) DeleteMessage(
	ctx context.Context,
	messageID int64,
) error {

	const query = `
		UPDATE chat_messages
		SET
			is_deleted = TRUE,
			deleted_at = NOW()
		WHERE
			id = $1
			AND is_deleted = FALSE
	`

	commandTag, err := r.db.Exec(
		ctx,
		query,
		messageID,
	)

	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrMessageNotFound
	}

	return nil
}

//
// Get Message Owner
//

func (r *repository) GetMessageOwner(
	ctx context.Context,
	messageID int64,
) (*MessageOwner, error) {

	const query = `
		SELECT
			id,
			stream_id,
			user_id
		FROM chat_messages
		WHERE id = $1
	`

	var owner MessageOwner

	err := r.db.QueryRow(
		ctx,
		query,
		messageID,
	).Scan(
		&owner.MessageID,
		&owner.StreamID,
		&owner.UserID,
	)

	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMessageNotFound
		}

		return nil, err
	}

	return &owner, nil
}

//
// Get Stream Owner
//
// Resolves:
//
// chat_messages
//      ↓
// streams
//      ↓
// channels
//      ↓
// channels.user_id
//

type StreamOwner struct {
	StreamID  uuid.UUID
	ChannelID uuid.UUID
	UserID    uuid.UUID
}

func (r *repository) GetStreamOwner(
	ctx context.Context,
	streamID uuid.UUID,
) (*StreamOwner, error) {

	const query = `
		SELECT
			s.id,
			c.id,
			c.user_id
		FROM streams s
		INNER JOIN channels c
			ON c.id = s.channel_id
		WHERE s.id = $1
	`

	var owner StreamOwner

	err := r.db.QueryRow(
		ctx,
		query,
		streamID,
	).Scan(
		&owner.StreamID,
		&owner.ChannelID,
		&owner.UserID,
	)

	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrStreamNotFound
		}

		return nil, err
	}

	return &owner, nil
}

//
// Get Message Statistics
//

func (r *repository) GetMessageStats(
	ctx context.Context,
	streamID uuid.UUID,
) (*MessageStats, error) {

	const query = `
		SELECT
			COUNT(*) AS total_messages,
			COUNT(*) FILTER (
				WHERE is_deleted = TRUE
			) AS deleted_messages,
			COUNT(*) FILTER (
				WHERE is_deleted = FALSE
			) AS active_messages
		FROM chat_messages
		WHERE stream_id = $1
	`

	var stats MessageStats

	stats.StreamID = streamID

	err := r.db.QueryRow(
		ctx,
		query,
		streamID,
	).Scan(
		&stats.TotalMessages,
		&stats.DeletedMessages,
		&stats.ActiveMessages,
	)

	if err != nil {
		return nil, err
	}

	return &stats, nil
}
