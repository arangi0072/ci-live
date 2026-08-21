package channel

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrChannelNotFound     = errors.New("channel not found")
	ErrChannelExists       = errors.New("channel already exists")
	ErrUsernameTaken       = errors.New("username already taken")
	ErrInvalidChannel      = errors.New("invalid channel")
	ErrUnauthorizedChannel = errors.New("unauthorized channel")
)

type Repository interface {
	Create(
		ctx context.Context,
		channel *Channel,
	) error

	GetByID(
		ctx context.Context,
		id string,
	) (*Channel, error)

	GetByUserID(
		ctx context.Context,
		userID string,
	) (*Channel, error)

	GetByUsername(
		ctx context.Context,
		username string,
	) (*Channel, error)

	Update(
		ctx context.Context,
		id string,
		req UpdateChannelRequest,
	) (*Channel, error)

	Delete(
		ctx context.Context,
		id string,
	) error
}

type Service interface {
	Create(
		ctx context.Context,
		userID string,
		req CreateChannelRequest,
	) (*Channel, error)

	GetByID(
		ctx context.Context,
		id string,
	) (*Channel, error)

	GetMyChannel(
		ctx context.Context,
		userID string,
	) (*Channel, error)

	GetByUsername(
		ctx context.Context,
		username string,
	) (*Channel, error)

	Update(
		ctx context.Context,
		userID string,
		req UpdateChannelRequest,
	) (*Channel, error)
}

type channelService struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &channelService{
		repository: repository,
	}
}

var _ Service = (*channelService)(nil)

// ============================================================
// Create Channel
// ============================================================

func (s *channelService) Create(
	ctx context.Context,
	userID string,
	req CreateChannelRequest,
) (*Channel, error) {

	userID = strings.TrimSpace(userID)

	if userID == "" {
		return nil, ErrInvalidChannel
	}

	// --------------------------------------------------------
	// One channel per user
	// --------------------------------------------------------

	existing, err := s.repository.GetByUserID(
		ctx,
		userID,
	)

	if err == nil && existing != nil {
		return nil, ErrChannelExists
	}

	if !errors.Is(err, ErrChannelNotFound) && err != nil {
		return nil, fmt.Errorf(
			"check existing channel: %w",
			err,
		)
	}

	// --------------------------------------------------------
	// Validate username
	// --------------------------------------------------------

	username, err := normalizeUsername(
		req.Username,
	)

	if err != nil {
		return nil, err
	}

	// --------------------------------------------------------
	// Check username uniqueness
	// --------------------------------------------------------

	existing, err = s.repository.GetByUsername(
		ctx,
		username,
	)

	if err == nil && existing != nil {
		return nil, ErrUsernameTaken
	}

	if !errors.Is(err, ErrChannelNotFound) && err != nil {
		return nil, fmt.Errorf(
			"check channel username: %w",
			err,
		)
	}

	// --------------------------------------------------------
	// Validate name
	// --------------------------------------------------------

	name := strings.TrimSpace(req.Name)

	if name == "" || len(name) > 100 {
		return nil, ErrInvalidChannel
	}

	// --------------------------------------------------------
	// Optional fields
	// --------------------------------------------------------

	description := strings.TrimSpace(
		req.Description,
	)

	if len(description) > 1000 {
		return nil, ErrInvalidChannel
	}

	avatarURL := strings.TrimSpace(
		req.AvatarURL,
	)

	if len(avatarURL) > 2048 {
		return nil, ErrInvalidChannel
	}

	bannerURL := strings.TrimSpace(
		req.BannerURL,
	)

	if len(bannerURL) > 2048 {
		return nil, ErrInvalidChannel
	}

	// --------------------------------------------------------
	// Create channel
	// --------------------------------------------------------

	channel := &Channel{
		ID:            uuid.NewString(),
		UserID:        userID,
		Username:      username,
		Name:          name,
		Description:   description,
		AvatarURL:     avatarURL,
		BannerURL:     bannerURL,
		IsVerified:    false,
		FollowerCount: 0,
		TotalViews:    0,
	}

	if err := s.repository.Create(
		ctx,
		channel,
	); err != nil {
		return nil, fmt.Errorf(
			"create channel: %w",
			err,
		)
	}

	return channel, nil
}

// ============================================================
// Get Channel by ID
// ============================================================

func (s *channelService) GetByID(
	ctx context.Context,
	id string,
) (*Channel, error) {

	id = strings.TrimSpace(id)

	if id == "" {
		return nil, ErrInvalidChannel
	}

	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidChannel
	}

	channel, err := s.repository.GetByID(
		ctx,
		id,
	)

	if err != nil {
		if errors.Is(err, ErrChannelNotFound) {
			return nil, ErrChannelNotFound
		}

		return nil, fmt.Errorf(
			"get channel: %w",
			err,
		)
	}

	return channel, nil
}

// ============================================================
// Get My Channel
// ============================================================

func (s *channelService) GetMyChannel(
	ctx context.Context,
	userID string,
) (*Channel, error) {

	userID = strings.TrimSpace(userID)

	if userID == "" {
		return nil, ErrInvalidChannel
	}

	if _, err := uuid.Parse(userID); err != nil {
		return nil, ErrInvalidChannel
	}

	channel, err := s.repository.GetByUserID(
		ctx,
		userID,
	)

	if err != nil {
		if errors.Is(err, ErrChannelNotFound) {
			return nil, ErrChannelNotFound
		}

		return nil, fmt.Errorf(
			"get my channel: %w",
			err,
		)
	}

	return channel, nil
}

// ============================================================
// Get Channel by Username
// ============================================================

func (s *channelService) GetByUsername(
	ctx context.Context,
	username string,
) (*Channel, error) {

	username, err := normalizeUsername(
		username,
	)

	if err != nil {
		return nil, err
	}

	channel, err := s.repository.GetByUsername(
		ctx,
		username,
	)

	if err != nil {
		if errors.Is(err, ErrChannelNotFound) {
			return nil, ErrChannelNotFound
		}

		return nil, fmt.Errorf(
			"get channel by username: %w",
			err,
		)
	}

	return channel, nil
}

// ============================================================
// Update Channel
// ============================================================

func (s *channelService) Update(
	ctx context.Context,
	userID string,
	req UpdateChannelRequest,
) (*Channel, error) {

	userID = strings.TrimSpace(userID)

	if userID == "" {
		return nil, ErrInvalidChannel
	}

	if _, err := uuid.Parse(userID); err != nil {
		return nil, ErrInvalidChannel
	}

	// --------------------------------------------------------
	// Get channel owned by authenticated user
	// --------------------------------------------------------

	channel, err := s.repository.GetByUserID(
		ctx,
		userID,
	)

	if err != nil {
		if errors.Is(err, ErrChannelNotFound) {
			return nil, ErrChannelNotFound
		}

		return nil, fmt.Errorf(
			"get channel for update: %w",
			err,
		)
	}

	// --------------------------------------------------------
	// Ownership check
	// --------------------------------------------------------

	if channel.UserID != userID {
		return nil, ErrUnauthorizedChannel
	}

	// --------------------------------------------------------
	// Validate name
	// --------------------------------------------------------

	if req.Name != nil {
		name := strings.TrimSpace(
			*req.Name,
		)

		if name == "" || len(name) > 100 {
			return nil, ErrInvalidChannel
		}

		req.Name = &name
	}

	// --------------------------------------------------------
	// Validate description
	// --------------------------------------------------------

	if req.Description != nil {
		description := strings.TrimSpace(
			*req.Description,
		)

		if len(description) > 1000 {
			return nil, ErrInvalidChannel
		}

		req.Description = &description
	}

	// --------------------------------------------------------
	// Validate avatar URL
	// --------------------------------------------------------

	if req.AvatarURL != nil {
		avatarURL := strings.TrimSpace(
			*req.AvatarURL,
		)

		if len(avatarURL) > 2048 {
			return nil, ErrInvalidChannel
		}

		req.AvatarURL = &avatarURL
	}

	// --------------------------------------------------------
	// Validate banner URL
	// --------------------------------------------------------

	if req.BannerURL != nil {
		bannerURL := strings.TrimSpace(
			*req.BannerURL,
		)

		if len(bannerURL) > 2048 {
			return nil, ErrInvalidChannel
		}

		req.BannerURL = &bannerURL
	}

	// --------------------------------------------------------
	// Nothing to update
	// --------------------------------------------------------

	if req.Name == nil &&
		req.Description == nil &&
		req.AvatarURL == nil &&
		req.BannerURL == nil {

		return nil, ErrInvalidChannel
	}

	// --------------------------------------------------------
	// Update
	// --------------------------------------------------------

	updated, err := s.repository.Update(
		ctx,
		channel.ID,
		req,
	)

	if err != nil {
		if errors.Is(err, ErrChannelNotFound) {
			return nil, ErrChannelNotFound
		}

		if errors.Is(err, ErrUsernameTaken) {
			return nil, ErrUsernameTaken
		}

		return nil, fmt.Errorf(
			"update channel: %w",
			err,
		)
	}

	return updated, nil
}

// ============================================================
// Username Validation
// ============================================================

var usernameRegex = regexp.MustCompile(
	`^[a-z0-9_]{3,50}$`,
)

func normalizeUsername(
	username string,
) (string, error) {

	username = strings.ToLower(
		strings.TrimSpace(username),
	)

	if !usernameRegex.MatchString(username) {
		return "", ErrInvalidChannel
	}

	return username, nil
}
