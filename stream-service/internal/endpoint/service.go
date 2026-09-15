package endpoint

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUnauthorized          = errors.New("unauthorized")
	ErrForbidden             = errors.New("forbidden")
	ErrEndpointNotFound      = errors.New("endpoint not found")
	ErrStreamNotFound        = errors.New("stream not found")
	ErrInvalidEndpointType   = errors.New("invalid endpoint type")
	ErrInvalidProtocol       = errors.New("invalid protocol")
	ErrEndpointAlreadyExists = errors.New("endpoint already exists")
)

type Service interface {
	CreateEndpoint(
		ctx context.Context,
		userID uuid.UUID,
		req CreateEndpointRequest,
	) (*EndpointResponse, error)

	GetEndpoint(
		ctx context.Context,
		endpointID uuid.UUID,
	) (*EndpointResponse, error)

	ListEndpoints(
		ctx context.Context,
		filter EndpointFilter,
	) (*EndpointListResponse, error)

	UpdateEndpoint(
		ctx context.Context,
		userID uuid.UUID,
		endpointID uuid.UUID,
		req UpdateEndpointRequest,
	) (*EndpointResponse, error)

	SetEndpointActive(
		ctx context.Context,
		userID uuid.UUID,
		endpointID uuid.UUID,
		active bool,
	) (*EndpointResponse, error)

	DeleteEndpoint(
		ctx context.Context,
		userID uuid.UUID,
		endpointID uuid.UUID,
	) error
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{
		repository: repository,
	}
}

//
// Create Endpoint
//

func (s *service) CreateEndpoint(
	ctx context.Context,
	userID uuid.UUID,
	req CreateEndpointRequest,
) (*EndpointResponse, error) {

	if userID == uuid.Nil {
		return nil, ErrUnauthorized
	}

	if req.StreamID == uuid.Nil {
		return nil, ErrStreamNotFound
	}

	if !req.EndpointType.IsValid() {
		return nil, ErrInvalidEndpointType
	}

	if !req.Protocol.IsValid() {
		return nil, ErrInvalidProtocol
	}

	if req.URL == "" {
		return nil, errors.New("endpoint URL is required")
	}

	//
	// Verify stream ownership.
	//

	owner, err := s.repository.GetStreamOwner(
		ctx,
		req.StreamID,
	)

	if err != nil {
		if errors.Is(err, ErrStreamNotFound) {
			return nil, ErrStreamNotFound
		}

		return nil, fmt.Errorf(
			"get stream owner: %w",
			err,
		)
	}

	if owner.UserID != userID {
		return nil, ErrForbidden
	}

	//
	// Check duplicate endpoint.
	//

	existing, err := s.repository.GetByStreamTypeProtocol(
		ctx,
		req.StreamID,
		req.EndpointType,
		req.Protocol,
	)

	if err == nil && existing != nil {
		return nil, ErrEndpointAlreadyExists
	}

	if err != nil &&
		!errors.Is(err, ErrEndpointNotFound) {

		return nil, fmt.Errorf(
			"check existing endpoint: %w",
			err,
		)
	}

	//
	// Default active state.
	//

	isActive := true

	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	now := time.Now().UTC()

	endpoint := &StreamEndpoint{
		ID:           uuid.New(),
		StreamID:     req.StreamID,
		EndpointType: req.EndpointType,
		Protocol:     req.Protocol,
		URL:          req.URL,
		IsActive:     isActive,
		CreatedAt:    now,
	}

	if err := s.repository.Create(
		ctx,
		endpoint,
	); err != nil {

		if errors.Is(err, ErrEndpointAlreadyExists) {
			return nil, ErrEndpointAlreadyExists
		}

		return nil, fmt.Errorf(
			"create endpoint: %w",
			err,
		)
	}

	return ToEndpointResponse(endpoint), nil
}

//
// Get Endpoint
//

func (s *service) GetEndpoint(
	ctx context.Context,
	endpointID uuid.UUID,
) (*EndpointResponse, error) {

	if endpointID == uuid.Nil {
		return nil, ErrEndpointNotFound
	}

	endpoint, err := s.repository.GetByID(
		ctx,
		endpointID,
	)

	if err != nil {
		if errors.Is(err, ErrEndpointNotFound) {
			return nil, ErrEndpointNotFound
		}

		return nil, fmt.Errorf(
			"get endpoint: %w",
			err,
		)
	}

	return ToEndpointResponse(endpoint), nil
}

//
// List Endpoints
//

func (s *service) ListEndpoints(
	ctx context.Context,
	filter EndpointFilter,
) (*EndpointListResponse, error) {

	endpoints, total, err := s.repository.List(
		ctx,
		filter,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"list endpoints: %w",
			err,
		)
	}

	return &EndpointListResponse{
		Endpoints: ToEndpointResponses(endpoints),
		Total:     total,
	}, nil
}

//
// Update Endpoint
//

func (s *service) UpdateEndpoint(
	ctx context.Context,
	userID uuid.UUID,
	endpointID uuid.UUID,
	req UpdateEndpointRequest,
) (*EndpointResponse, error) {

	if userID == uuid.Nil {
		return nil, ErrUnauthorized
	}

	endpoint, err := s.getOwnedEndpoint(
		ctx,
		userID,
		endpointID,
	)

	if err != nil {
		return nil, err
	}

	//
	// Only URL and active state can be changed.
	//

	if req.URL != nil {

		if *req.URL == "" {
			return nil, errors.New(
				"endpoint URL cannot be empty",
			)
		}

		endpoint.URL = *req.URL
	}

	if req.IsActive != nil {
		endpoint.IsActive = *req.IsActive
	}

	if err := s.repository.Update(
		ctx,
		endpoint,
	); err != nil {

		if errors.Is(err, ErrEndpointNotFound) {
			return nil, ErrEndpointNotFound
		}

		return nil, fmt.Errorf(
			"update endpoint: %w",
			err,
		)
	}

	return ToEndpointResponse(endpoint), nil
}

//
// Activate / Deactivate Endpoint
//

func (s *service) SetEndpointActive(
	ctx context.Context,
	userID uuid.UUID,
	endpointID uuid.UUID,
	active bool,
) (*EndpointResponse, error) {

	if userID == uuid.Nil {
		return nil, ErrUnauthorized
	}

	endpoint, err := s.getOwnedEndpoint(
		ctx,
		userID,
		endpointID,
	)

	if err != nil {
		return nil, err
	}

	endpoint.IsActive = active

	if err := s.repository.SetActive(
		ctx,
		endpointID,
		active,
	); err != nil {

		if errors.Is(err, ErrEndpointNotFound) {
			return nil, ErrEndpointNotFound
		}

		return nil, fmt.Errorf(
			"set endpoint active: %w",
			err,
		)
	}

	return ToEndpointResponse(endpoint), nil
}

//
// Delete Endpoint
//

func (s *service) DeleteEndpoint(
	ctx context.Context,
	userID uuid.UUID,
	endpointID uuid.UUID,
) error {

	if userID == uuid.Nil {
		return ErrUnauthorized
	}

	_, err := s.getOwnedEndpoint(
		ctx,
		userID,
		endpointID,
	)

	if err != nil {
		return err
	}

	if err := s.repository.Delete(
		ctx,
		endpointID,
	); err != nil {

		if errors.Is(err, ErrEndpointNotFound) {
			return ErrEndpointNotFound
		}

		return fmt.Errorf(
			"delete endpoint: %w",
			err,
		)
	}

	return nil
}

//
// Get Owned Endpoint
//

func (s *service) getOwnedEndpoint(
	ctx context.Context,
	userID uuid.UUID,
	endpointID uuid.UUID,
) (*StreamEndpoint, error) {

	if endpointID == uuid.Nil {
		return nil, ErrEndpointNotFound
	}

	endpoint, err := s.repository.GetByID(
		ctx,
		endpointID,
	)

	if err != nil {
		if errors.Is(err, ErrEndpointNotFound) {
			return nil, ErrEndpointNotFound
		}

		return nil, fmt.Errorf(
			"get endpoint: %w",
			err,
		)
	}

	owner, err := s.repository.GetEndpointOwner(
		ctx,
		endpointID,
	)

	if err != nil {
		if errors.Is(err, ErrEndpointNotFound) {
			return nil, ErrEndpointNotFound
		}

		return nil, fmt.Errorf(
			"get endpoint owner: %w",
			err,
		)
	}

	if owner.UserID != userID {
		return nil, ErrForbidden
	}

	return endpoint, nil
}
