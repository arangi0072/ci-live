package endpoint

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(
		ctx context.Context,
		endpoint *StreamEndpoint,
	) error

	GetByID(
		ctx context.Context,
		endpointID uuid.UUID,
	) (*StreamEndpoint, error)

	List(
		ctx context.Context,
		filter EndpointFilter,
	) ([]StreamEndpoint, int64, error)

	Update(
		ctx context.Context,
		endpoint *StreamEndpoint,
	) error

	SetActive(
		ctx context.Context,
		endpointID uuid.UUID,
		active bool,
	) error

	Delete(
		ctx context.Context,
		endpointID uuid.UUID,
	) error

	GetEndpointOwner(
		ctx context.Context,
		endpointID uuid.UUID,
	) (*EndpointOwner, error)

	GetStreamOwner(
		ctx context.Context,
		streamID uuid.UUID,
	) (*EndpointOwner, error)

	GetByStreamTypeProtocol(
		ctx context.Context,
		streamID uuid.UUID,
		endpointType EndpointType,
		protocol EndpointProtocol,
	) (*StreamEndpoint, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{
		db: db,
	}
}

//
// Create Endpoint
//

func (r *repository) Create(
	ctx context.Context,
	endpoint *StreamEndpoint,
) error {
	const query = `
		INSERT INTO stream_endpoints (
			id,
			stream_id,
			endpoint_type,
			protocol,
			url,
			is_active,
			created_at
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7
		)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		endpoint.ID,
		endpoint.StreamID,
		endpoint.EndpointType,
		endpoint.Protocol,
		endpoint.URL,
		endpoint.IsActive,
		endpoint.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			// Unique constraint:
			// (stream_id, endpoint_type, protocol)
			if pgErr.Code == "23505" {
				return ErrEndpointAlreadyExists
			}
		}

		return fmt.Errorf(
			"create endpoint: %w",
			err,
		)
	}

	return nil
}

//
// Get Endpoint By ID
//

func (r *repository) GetByID(
	ctx context.Context,
	endpointID uuid.UUID,
) (*StreamEndpoint, error) {
	const query = `
		SELECT
			id,
			stream_id,
			endpoint_type,
			protocol,
			url,
			is_active,
			created_at
		FROM stream_endpoints
		WHERE id = $1
	`

	endpoint := &StreamEndpoint{}

	err := r.db.QueryRow(
		ctx,
		query,
		endpointID,
	).Scan(
		&endpoint.ID,
		&endpoint.StreamID,
		&endpoint.EndpointType,
		&endpoint.Protocol,
		&endpoint.URL,
		&endpoint.IsActive,
		&endpoint.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrEndpointNotFound
		}

		return nil, fmt.Errorf(
			"get endpoint: %w",
			err,
		)
	}

	return endpoint, nil
}

//
// List Endpoints
//

func (r *repository) List(
	ctx context.Context,
	filter EndpointFilter,
) ([]StreamEndpoint, int64, error) {

	conditions := make([]string, 0, 4)
	args := make([]interface{}, 0, 4)

	argIndex := 1

	if filter.StreamID != nil {
		conditions = append(
			conditions,
			fmt.Sprintf("stream_id = $%d", argIndex),
		)

		args = append(args, *filter.StreamID)
		argIndex++
	}

	if filter.EndpointType != nil {
		conditions = append(
			conditions,
			fmt.Sprintf("endpoint_type = $%d", argIndex),
		)

		args = append(args, *filter.EndpointType)
		argIndex++
	}

	if filter.Protocol != nil {
		conditions = append(
			conditions,
			fmt.Sprintf("protocol = $%d", argIndex),
		)

		args = append(args, *filter.Protocol)
		argIndex++
	}

	if filter.IsActive != nil {
		conditions = append(
			conditions,
			fmt.Sprintf("is_active = $%d", argIndex),
		)

		args = append(args, *filter.IsActive)
		argIndex++
	}

	whereClause := ""

	if len(conditions) > 0 {
		whereClause = "WHERE " + joinConditions(conditions)
	}

	//
	// Count
	//

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM stream_endpoints
		%s
	`, whereClause)

	var total int64

	if err := r.db.QueryRow(
		ctx,
		countQuery,
		args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf(
			"count endpoints: %w",
			err,
		)
	}

	//
	// Data
	//

	query := fmt.Sprintf(`
		SELECT
			id,
			stream_id,
			endpoint_type,
			protocol,
			url,
			is_active,
			created_at
		FROM stream_endpoints
		%s
		ORDER BY created_at DESC
	`, whereClause)

	rows, err := r.db.Query(
		ctx,
		query,
		args...,
	)

	if err != nil {
		return nil, 0, fmt.Errorf(
			"list endpoints: %w",
			err,
		)
	}

	defer rows.Close()

	endpoints := make(
		[]StreamEndpoint,
		0,
	)

	for rows.Next() {
		var endpoint StreamEndpoint

		if err := rows.Scan(
			&endpoint.ID,
			&endpoint.StreamID,
			&endpoint.EndpointType,
			&endpoint.Protocol,
			&endpoint.URL,
			&endpoint.IsActive,
			&endpoint.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf(
				"scan endpoint: %w",
				err,
			)
		}

		endpoints = append(
			endpoints,
			endpoint,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf(
			"iterate endpoints: %w",
			err,
		)
	}

	return endpoints, total, nil
}

//
// Update Endpoint
//

func (r *repository) Update(
	ctx context.Context,
	endpoint *StreamEndpoint,
) error {
	const query = `
		UPDATE stream_endpoints
		SET
			url = $2,
			is_active = $3
		WHERE id = $1
	`

	result, err := r.db.Exec(
		ctx,
		query,
		endpoint.ID,
		endpoint.URL,
		endpoint.IsActive,
	)

	if err != nil {
		return fmt.Errorf(
			"update endpoint: %w",
			err,
		)
	}

	if result.RowsAffected() == 0 {
		return ErrEndpointNotFound
	}

	return nil
}

//
// Set Active
//

func (r *repository) SetActive(
	ctx context.Context,
	endpointID uuid.UUID,
	active bool,
) error {
	const query = `
		UPDATE stream_endpoints
		SET is_active = $2
		WHERE id = $1
	`

	result, err := r.db.Exec(
		ctx,
		query,
		endpointID,
		active,
	)

	if err != nil {
		return fmt.Errorf(
			"set endpoint active: %w",
			err,
		)
	}

	if result.RowsAffected() == 0 {
		return ErrEndpointNotFound
	}

	return nil
}

//
// Delete Endpoint
//

func (r *repository) Delete(
	ctx context.Context,
	endpointID uuid.UUID,
) error {
	const query = `
		DELETE FROM stream_endpoints
		WHERE id = $1
	`

	result, err := r.db.Exec(
		ctx,
		query,
		endpointID,
	)

	if err != nil {
		return fmt.Errorf(
			"delete endpoint: %w",
			err,
		)
	}

	if result.RowsAffected() == 0 {
		return ErrEndpointNotFound
	}

	return nil
}

//
// Get Endpoint Owner
//
// stream_endpoints
//        ↓ stream_id
// streams
//        ↓ channel_id
// channels
//        ↓ user_id
//

func (r *repository) GetEndpointOwner(
	ctx context.Context,
	endpointID uuid.UUID,
) (*EndpointOwner, error) {
	const query = `
		SELECT
			se.id AS endpoint_id,
			se.stream_id,
			s.channel_id,
			c.user_id
		FROM stream_endpoints se
		INNER JOIN streams s
			ON s.id = se.stream_id
		INNER JOIN channels c
			ON c.id = s.channel_id
		WHERE se.id = $1
	`

	owner := &EndpointOwner{}

	err := r.db.QueryRow(
		ctx,
		query,
		endpointID,
	).Scan(
		&owner.EndpointID,
		&owner.StreamID,
		&owner.ChannelID,
		&owner.UserID,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrEndpointNotFound
		}

		return nil, fmt.Errorf(
			"get endpoint owner: %w",
			err,
		)
	}

	return owner, nil
}

//
// Get Stream Owner
//

func (r *repository) GetStreamOwner(
	ctx context.Context,
	streamID uuid.UUID,
) (*EndpointOwner, error) {
	const query = `
		SELECT
			s.id AS stream_id,
			s.channel_id,
			c.user_id
		FROM streams s
		INNER JOIN channels c
			ON c.id = s.channel_id
		WHERE s.id = $1
	`

	owner := &EndpointOwner{}

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

		return nil, fmt.Errorf(
			"get stream owner: %w",
			err,
		)
	}

	return owner, nil
}

//
// Get Endpoint By Stream + Type + Protocol
//

func (r *repository) GetByStreamTypeProtocol(
	ctx context.Context,
	streamID uuid.UUID,
	endpointType EndpointType,
	protocol EndpointProtocol,
) (*StreamEndpoint, error) {
	const query = `
		SELECT
			id,
			stream_id,
			endpoint_type,
			protocol,
			url,
			is_active,
			created_at
		FROM stream_endpoints
		WHERE stream_id = $1
		  AND endpoint_type = $2
		  AND protocol = $3
		LIMIT 1
	`

	endpoint := &StreamEndpoint{}

	err := r.db.QueryRow(
		ctx,
		query,
		streamID,
		endpointType,
		protocol,
	).Scan(
		&endpoint.ID,
		&endpoint.StreamID,
		&endpoint.EndpointType,
		&endpoint.Protocol,
		&endpoint.URL,
		&endpoint.IsActive,
		&endpoint.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrEndpointNotFound
		}

		return nil, fmt.Errorf(
			"get endpoint by stream type protocol: %w",
			err,
		)
	}

	return endpoint, nil
}

//
// Join SQL Conditions
//

func joinConditions(
	conditions []string,
) string {
	if len(conditions) == 0 {
		return ""
	}

	result := conditions[0]

	for i := 1; i < len(conditions); i++ {
		result += " AND " + conditions[i]
	}

	return result
}
