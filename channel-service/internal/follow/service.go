package follow

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrFollowNotFound   = errors.New("follow relationship not found")
	ErrAlreadyFollowing = errors.New("already following channel")
	ErrInvalidFollow    = errors.New("invalid follow request")
)

type Repository interface {
	Create(
		ctx context.Context,
		follow *ChannelFollow,
	) error

	Delete(
		ctx context.Context,
		userID string,
		channelID string,
	) error

	Exists(
		ctx context.Context,
		userID string,
		channelID string,
	) (bool, error)

	CountFollowers(
		ctx context.Context,
		channelID string,
	) (int64, error)

	GetFollowing(
		ctx context.Context,
		userID string,
	) ([]ChannelFollow, error)
}

type Service interface {
	Follow(
		ctx context.Context,
		userID string,
		channelID string,
	) error

	Unfollow(
		ctx context.Context,
		userID string,
		channelID string,
	) error

	IsFollowing(
		ctx context.Context,
		userID string,
		channelID string,
	) (bool, error)

	GetFollowerCount(
		ctx context.Context,
		channelID string,
	) (int64, error)

	GetFollowing(
		ctx context.Context,
		userID string,
	) ([]ChannelFollow, error)
}

type followService struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &followService{
		repository: repository,
	}
}

var _ Service = (*followService)(nil)

// ============================================================
// Follow Channel
// ============================================================

func (s *followService) Follow(
	ctx context.Context,
	userID string,
	channelID string,
) error {

	userID = strings.TrimSpace(userID)
	channelID = strings.TrimSpace(channelID)

	if !validUUID(userID) ||
		!validUUID(channelID) {

		return ErrInvalidFollow
	}

	// --------------------------------------------------------
	// Check existing relationship
	// --------------------------------------------------------

	exists, err := s.repository.Exists(
		ctx,
		userID,
		channelID,
	)

	if err != nil {
		return fmt.Errorf(
			"check existing follow: %w",
			err,
		)
	}

	if exists {
		return ErrAlreadyFollowing
	}

	// --------------------------------------------------------
	// Create follow
	// --------------------------------------------------------

	follow := &ChannelFollow{
		ID:        uuid.NewString(),
		UserID:    userID,
		ChannelID: channelID,
	}

	if err := s.repository.Create(
		ctx,
		follow,
	); err != nil {
		return fmt.Errorf(
			"create follow: %w",
			err,
		)
	}

	return nil
}

// ============================================================
// Unfollow Channel
// ============================================================

func (s *followService) Unfollow(
	ctx context.Context,
	userID string,
	channelID string,
) error {

	userID = strings.TrimSpace(userID)
	channelID = strings.TrimSpace(channelID)

	if !validUUID(userID) ||
		!validUUID(channelID) {

		return ErrInvalidFollow
	}

	// --------------------------------------------------------
	// Check whether relationship exists
	// --------------------------------------------------------

	exists, err := s.repository.Exists(
		ctx,
		userID,
		channelID,
	)

	if err != nil {
		return fmt.Errorf(
			"check follow before unfollow: %w",
			err,
		)
	}

	if !exists {
		return ErrFollowNotFound
	}

	// --------------------------------------------------------
	// Delete relationship
	// --------------------------------------------------------

	if err := s.repository.Delete(
		ctx,
		userID,
		channelID,
	); err != nil {

		if errors.Is(err, ErrFollowNotFound) {
			return ErrFollowNotFound
		}

		return fmt.Errorf(
			"delete follow: %w",
			err,
		)
	}

	return nil
}

// ============================================================
// Check Follow Status
// ============================================================

func (s *followService) IsFollowing(
	ctx context.Context,
	userID string,
	channelID string,
) (bool, error) {

	userID = strings.TrimSpace(userID)
	channelID = strings.TrimSpace(channelID)

	if !validUUID(userID) ||
		!validUUID(channelID) {

		return false, ErrInvalidFollow
	}

	following, err := s.repository.Exists(
		ctx,
		userID,
		channelID,
	)

	if err != nil {
		return false, fmt.Errorf(
			"check follow status: %w",
			err,
		)
	}

	return following, nil
}

// ============================================================
// Get Follower Count
// ============================================================

func (s *followService) GetFollowerCount(
	ctx context.Context,
	channelID string,
) (int64, error) {

	channelID = strings.TrimSpace(channelID)

	if !validUUID(channelID) {
		return 0, ErrInvalidFollow
	}

	count, err := s.repository.CountFollowers(
		ctx,
		channelID,
	)

	if err != nil {
		return 0, fmt.Errorf(
			"get follower count: %w",
			err,
		)
	}

	return count, nil
}

// ============================================================
// Get Channels Followed By User
// ============================================================

func (s *followService) GetFollowing(
	ctx context.Context,
	userID string,
) ([]ChannelFollow, error) {

	userID = strings.TrimSpace(userID)

	if !validUUID(userID) {
		return nil, ErrInvalidFollow
	}

	following, err := s.repository.GetFollowing(
		ctx,
		userID,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"get following: %w",
			err,
		)
	}

	if following == nil {
		following = make(
			[]ChannelFollow,
			0,
		)
	}

	return following, nil
}

// ============================================================
// UUID Validation
// ============================================================

func validUUID(value string) bool {
	_, err := uuid.Parse(value)

	return err == nil
}
