package like

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

//
// Service Errors
//

var (
	ErrUnauthorized = errors.New(
		"unauthorized",
	)

	ErrForbidden = errors.New(
		"forbidden",
	)

	ErrLikeNotFound = errors.New(
		"like not found",
	)

	ErrStreamNotFound = errors.New(
		"stream not found",
	)

	ErrAlreadyLiked = errors.New(
		"stream already liked",
	)

	ErrNotLiked = errors.New(
		"stream is not liked",
	)

	ErrInvalidUUID = errors.New(
		"invalid UUID",
	)
)

//
// Repository Interface
//
// Implemented by repository.go.
//

type Repository interface {
	CreateLike(
		ctx context.Context,
		like *StreamLike,
	) (*StreamLike, error)

	GetLike(
		ctx context.Context,
		streamID uuid.UUID,
		userID uuid.UUID,
	) (*StreamLike, error)

	DeleteLike(
		ctx context.Context,
		streamID uuid.UUID,
		userID uuid.UUID,
	) error

	IsLiked(
		ctx context.Context,
		streamID uuid.UUID,
		userID uuid.UUID,
	) (bool, error)

	ListLikes(
		ctx context.Context,
		filter LikeFilter,
	) ([]StreamLike, int64, error)

	GetStreamOwner(
		ctx context.Context,
		streamID uuid.UUID,
	) (*StreamOwner, error)

	GetLikeStats(
		ctx context.Context,
		streamID uuid.UUID,
	) (*LikeStats, error)
}

//
// Service Interface
//

type Service interface {
	CreateLike(
		ctx context.Context,
		userID uuid.UUID,
		req CreateLikeRequest,
	) (*LikeResponse, error)

	RemoveLike(
		ctx context.Context,
		userID uuid.UUID,
		streamID uuid.UUID,
	) error

	GetLike(
		ctx context.Context,
		userID uuid.UUID,
		streamID uuid.UUID,
	) (*LikeResponse, error)

	IsLiked(
		ctx context.Context,
		userID uuid.UUID,
		streamID uuid.UUID,
	) (bool, error)

	ListLikes(
		ctx context.Context,
		filter LikeFilter,
	) (*LikeListResponse, error)

	GetLikeStats(
		ctx context.Context,
		streamID uuid.UUID,
	) (*LikeStats, error)
}

//
// Service Implementation
//

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{
		repository: repository,
	}
}

//
// Create Like
//

func (s *service) CreateLike(
	ctx context.Context,
	userID uuid.UUID,
	req CreateLikeRequest,
) (*LikeResponse, error) {

	if userID == uuid.Nil {
		return nil, ErrUnauthorized
	}

	if req.StreamID == uuid.Nil {
		return nil, ErrInvalidUUID
	}

	//
	// Verify stream exists.
	//

	_, err := s.repository.GetStreamOwner(
		ctx,
		req.StreamID,
	)

	if err != nil {
		return nil, err
	}

	//
	// Prevent duplicate likes.
	//

	liked, err := s.repository.IsLiked(
		ctx,
		req.StreamID,
		userID,
	)

	if err != nil {
		return nil, err
	}

	if liked {
		return nil, ErrAlreadyLiked
	}

	//
	// Create like.
	//

	like := &StreamLike{
		StreamID: req.StreamID,
		UserID:   userID,
	}

	createdLike, err := s.repository.CreateLike(
		ctx,
		like,
	)

	if err != nil {
		return nil, err
	}

	return ToLikeResponse(createdLike), nil
}

//
// Remove Like
//

func (s *service) RemoveLike(
	ctx context.Context,
	userID uuid.UUID,
	streamID uuid.UUID,
) error {

	if userID == uuid.Nil {
		return ErrUnauthorized
	}

	if streamID == uuid.Nil {
		return ErrInvalidUUID
	}

	//
	// Verify stream exists.
	//

	_, err := s.repository.GetStreamOwner(
		ctx,
		streamID,
	)

	if err != nil {
		return err
	}

	//
	// Make sure the user has actually liked it.
	//

	liked, err := s.repository.IsLiked(
		ctx,
		streamID,
		userID,
	)

	if err != nil {
		return err
	}

	if !liked {
		return ErrNotLiked
	}

	return s.repository.DeleteLike(
		ctx,
		streamID,
		userID,
	)
}

//
// Get Like
//

func (s *service) GetLike(
	ctx context.Context,
	userID uuid.UUID,
	streamID uuid.UUID,
) (*LikeResponse, error) {

	if userID == uuid.Nil {
		return nil, ErrUnauthorized
	}

	if streamID == uuid.Nil {
		return nil, ErrInvalidUUID
	}

	like, err := s.repository.GetLike(
		ctx,
		streamID,
		userID,
	)

	if err != nil {
		return nil, err
	}

	if like == nil {
		return nil, ErrLikeNotFound
	}

	return ToLikeResponse(like), nil
}

//
// Is Liked
//

func (s *service) IsLiked(
	ctx context.Context,
	userID uuid.UUID,
	streamID uuid.UUID,
) (bool, error) {

	if userID == uuid.Nil {
		return false, ErrUnauthorized
	}

	if streamID == uuid.Nil {
		return false, ErrInvalidUUID
	}

	//
	// Verify stream exists.
	//

	_, err := s.repository.GetStreamOwner(
		ctx,
		streamID,
	)

	if err != nil {
		return false, err
	}

	return s.repository.IsLiked(
		ctx,
		streamID,
		userID,
	)
}

//
// List Likes
//

func (s *service) ListLikes(
	ctx context.Context,
	filter LikeFilter,
) (*LikeListResponse, error) {

	if filter.StreamID == nil ||
		*filter.StreamID == uuid.Nil {

		return nil, ErrInvalidUUID
	}

	//
	// Verify stream exists.
	//

	_, err := s.repository.GetStreamOwner(
		ctx,
		*filter.StreamID,
	)

	if err != nil {
		return nil, err
	}

	//
	// Pagination defaults.
	//

	if filter.Page < 1 {
		filter.Page = 1
	}

	if filter.Limit < 1 {
		filter.Limit = 50
	}

	if filter.Limit > 100 {
		filter.Limit = 100
	}

	likes, total, err := s.repository.ListLikes(
		ctx,
		filter,
	)

	if err != nil {
		return nil, err
	}

	return &LikeListResponse{
		Likes: ToLikeResponses(likes),
		Total: total,
		Page:  filter.Page,
		Limit: filter.Limit,
	}, nil
}

//
// Get Like Statistics
//

func (s *service) GetLikeStats(
	ctx context.Context,
	streamID uuid.UUID,
) (*LikeStats, error) {

	if streamID == uuid.Nil {
		return nil, ErrInvalidUUID
	}

	//
	// Verify stream exists.
	//

	_, err := s.repository.GetStreamOwner(
		ctx,
		streamID,
	)

	if err != nil {
		return nil, err
	}

	return s.repository.GetLikeStats(
		ctx,
		streamID,
	)
}
