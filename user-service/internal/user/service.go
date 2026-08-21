package user

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrProfileNotFound      = errors.New("profile not found")
	ErrProfileAlreadyExists = errors.New("profile already exists")
	ErrUsernameTaken        = errors.New("username taken")
	ErrInvalidUsername      = errors.New("invalid username")
	ErrInvalidProfile       = errors.New("invalid profile")
)

type Repository interface {
	Create(
		ctx context.Context,
		profile *UserProfile,
	) error

	GetByUserID(
		ctx context.Context,
		userID string,
	) (*UserProfile, error)

	Update(
		ctx context.Context,
		userID string,
		req UpdateProfileRequest,
	) (*UserProfile, error)

	UsernameExists(
		ctx context.Context,
		username string,
	) (bool, error)
}

type Service interface {
	CreateProfile(
		ctx context.Context,
		userID string,
		req CreateProfileRequest,
	) (*ProfileResponse, error)

	GetProfile(
		ctx context.Context,
		userID string,
	) (*ProfileResponse, error)

	UpdateProfile(
		ctx context.Context,
		userID string,
		req UpdateProfileRequest,
	) (*ProfileResponse, error)

	CheckUsernameAvailability(
		ctx context.Context,
		username string,
	) (*UsernameAvailabilityResponse, error)
}

type userService struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &userService{
		repository: repository,
	}
}

var _ Service = (*userService)(nil)

// ============================================================
// Create Profile
// ============================================================

func (s *userService) CreateProfile(
	ctx context.Context,
	userID string,
	req CreateProfileRequest,
) (*ProfileResponse, error) {

	userID = strings.TrimSpace(userID)

	if userID == "" {
		return nil, ErrInvalidProfile
	}

	username, err := normalizeUsername(req.Username)
	if err != nil {
		return nil, err
	}

	displayName := strings.TrimSpace(req.DisplayName)
	bio := strings.TrimSpace(req.Bio)
	avatarURL := strings.TrimSpace(req.AvatarURL)

	if displayName == "" {
		return nil, ErrInvalidProfile
	}

	if len(displayName) > 100 {
		return nil, ErrInvalidProfile
	}

	if len(bio) > 500 {
		return nil, ErrInvalidProfile
	}

	if len(avatarURL) > 2048 {
		return nil, ErrInvalidProfile
	}

	// Check whether profile already exists.
	existing, err := s.repository.GetByUserID(
		ctx,
		userID,
	)

	if err == nil && existing != nil {
		return nil, ErrProfileAlreadyExists
	}

	if !errors.Is(err, ErrProfileNotFound) {
		if err != nil {
			return nil, fmt.Errorf(
				"check existing profile: %w",
				err,
			)
		}
	}

	// Check username uniqueness.
	exists, err := s.repository.UsernameExists(
		ctx,
		username,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"check username: %w",
			err,
		)
	}

	if exists {
		return nil, ErrUsernameTaken
	}

	now := time.Now()

	profile := &UserProfile{
		UserID:      userID,
		Username:    username,
		DisplayName: displayName,
		Bio:         bio,
		AvatarURL:   avatarURL,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repository.Create(
		ctx,
		profile,
	); err != nil {
		return nil, fmt.Errorf(
			"create profile: %w",
			err,
		)
	}

	return toProfileResponse(profile), nil
}

// ============================================================
// Get Profile
// ============================================================

func (s *userService) GetProfile(
	ctx context.Context,
	userID string,
) (*ProfileResponse, error) {

	userID = strings.TrimSpace(userID)

	if userID == "" {
		return nil, ErrInvalidProfile
	}

	profile, err := s.repository.GetByUserID(
		ctx,
		userID,
	)

	if err != nil {
		if errors.Is(err, ErrProfileNotFound) {
			return nil, ErrProfileNotFound
		}

		return nil, fmt.Errorf(
			"get profile: %w",
			err,
		)
	}

	return toProfileResponse(profile), nil
}

// ============================================================
// Update Profile
// ============================================================

func (s *userService) UpdateProfile(
	ctx context.Context,
	userID string,
	req UpdateProfileRequest,
) (*ProfileResponse, error) {

	userID = strings.TrimSpace(userID)

	if userID == "" {
		return nil, ErrInvalidProfile
	}

	// Make sure the profile exists.
	_, err := s.repository.GetByUserID(
		ctx,
		userID,
	)

	if err != nil {
		if errors.Is(err, ErrProfileNotFound) {
			return nil, ErrProfileNotFound
		}

		return nil, fmt.Errorf(
			"get existing profile: %w",
			err,
		)
	}

	// --------------------------------------------------------
	// Validate username
	// --------------------------------------------------------

	if req.Username != nil {
		username, err := normalizeUsername(
			*req.Username,
		)

		if err != nil {
			return nil, err
		}

		// Check if username belongs to another profile.
		exists, err := s.repository.UsernameExists(
			ctx,
			username,
		)

		if err != nil {
			return nil, fmt.Errorf(
				"check username: %w",
				err,
			)
		}

		if exists {
			// Repository should enforce the unique constraint,
			// but this gives the service a clean error before
			// attempting the update.
			current, err := s.repository.GetByUserID(
				ctx,
				userID,
			)

			if err != nil {
				return nil, fmt.Errorf(
					"get current profile: %w",
					err,
				)
			}

			if current.Username != username {
				return nil, ErrUsernameTaken
			}
		}

		req.Username = &username
	}

	// --------------------------------------------------------
	// Validate display name
	// --------------------------------------------------------

	if req.DisplayName != nil {
		displayName := strings.TrimSpace(
			*req.DisplayName,
		)

		if displayName == "" ||
			len(displayName) > 100 {
			return nil, ErrInvalidProfile
		}

		req.DisplayName = &displayName
	}

	// --------------------------------------------------------
	// Validate bio
	// --------------------------------------------------------

	if req.Bio != nil {
		bio := strings.TrimSpace(*req.Bio)

		if len(bio) > 500 {
			return nil, ErrInvalidProfile
		}

		req.Bio = &bio
	}

	// --------------------------------------------------------
	// Validate avatar URL
	// --------------------------------------------------------

	if req.AvatarURL != nil {
		avatarURL := strings.TrimSpace(
			*req.AvatarURL,
		)

		if len(avatarURL) > 2048 {
			return nil, ErrInvalidProfile
		}

		req.AvatarURL = &avatarURL
	}

	// Make sure there is actually something to update.
	if req.Username == nil &&
		req.DisplayName == nil &&
		req.Bio == nil &&
		req.AvatarURL == nil {
		return nil, ErrInvalidProfile
	}

	profile, err := s.repository.Update(
		ctx,
		userID,
		req,
	)

	if err != nil {
		if errors.Is(err, ErrProfileNotFound) {
			return nil, ErrProfileNotFound
		}

		if errors.Is(err, ErrUsernameTaken) {
			return nil, ErrUsernameTaken
		}

		return nil, fmt.Errorf(
			"update profile: %w",
			err,
		)
	}

	return toProfileResponse(profile), nil
}

// ============================================================
// Username Availability
// ============================================================

func (s *userService) CheckUsernameAvailability(
	ctx context.Context,
	username string,
) (*UsernameAvailabilityResponse, error) {

	username, err := normalizeUsername(username)
	if err != nil {
		return nil, err
	}

	exists, err := s.repository.UsernameExists(
		ctx,
		username,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"check username availability: %w",
			err,
		)
	}

	return &UsernameAvailabilityResponse{
		Username:  username,
		Available: !exists,
	}, nil
}

// ============================================================
// Helpers
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
		return "", ErrInvalidUsername
	}

	return username, nil
}

func toProfileResponse(
	profile *UserProfile,
) *ProfileResponse {

	if profile == nil {
		return nil
	}

	return &ProfileResponse{
		UserID:      profile.UserID,
		Username:    profile.Username,
		DisplayName: profile.DisplayName,
		Bio:         profile.Bio,
		AvatarURL:   profile.AvatarURL,
		CreatedAt:   profile.CreatedAt,
	}
}

// Kept here if we need to generate profile-related IDs later.
func generateID() string {
	return uuid.NewString()
}
