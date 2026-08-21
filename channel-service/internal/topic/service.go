package topic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/google/uuid"
)

var (
	ErrTopicNotFound = errors.New("topic not found")
	ErrTopicExists   = errors.New("topic already exists")
	ErrInvalidTopic  = errors.New("invalid topic")
)

type Repository interface {
	Create(
		ctx context.Context,
		topic *Topic,
	) error

	GetByID(
		ctx context.Context,
		id string,
	) (*Topic, error)

	GetBySlug(
		ctx context.Context,
		slug string,
	) (*Topic, error)

	List(
		ctx context.Context,
		categoryID string,
	) ([]Topic, error)

	Update(
		ctx context.Context,
		id string,
		req UpdateTopicRequest,
	) (*Topic, error)

	Delete(
		ctx context.Context,
		id string,
	) error
}

type Service interface {
	Create(
		ctx context.Context,
		req CreateTopicRequest,
	) (*Topic, error)

	GetByID(
		ctx context.Context,
		id string,
	) (*Topic, error)

	GetBySlug(
		ctx context.Context,
		slug string,
	) (*Topic, error)

	List(
		ctx context.Context,
		categoryID string,
	) ([]Topic, error)

	Update(
		ctx context.Context,
		id string,
		req UpdateTopicRequest,
	) (*Topic, error)

	Delete(
		ctx context.Context,
		id string,
	) error
}

type topicService struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &topicService{
		repository: repository,
	}
}

var _ Service = (*topicService)(nil)

// ============================================================
// Create Topic
// ============================================================

func (s *topicService) Create(
	ctx context.Context,
	req CreateTopicRequest,
) (*Topic, error) {

	categoryID := strings.TrimSpace(req.CategoryID)

	if !validUUID(categoryID) {
		return nil, ErrInvalidTopic
	}

	name := strings.TrimSpace(req.Name)

	if !isValidTopicName(name) {
		return nil, ErrInvalidTopic
	}

	slug := generateSlug(name)

	if slug == "" {
		return nil, ErrInvalidTopic
	}

	// Check slug uniqueness.
	existing, err := s.repository.GetBySlug(
		ctx,
		slug,
	)

	if err == nil && existing != nil {
		return nil, ErrTopicExists
	}

	if !errors.Is(err, ErrTopicNotFound) && err != nil {
		return nil, fmt.Errorf(
			"check topic existence: %w",
			err,
		)
	}

	description := strings.TrimSpace(
		req.Description,
	)

	if len(description) > 500 {
		return nil, ErrInvalidTopic
	}

	topic := &Topic{
		ID:          uuid.NewString(),
		CategoryID:  categoryID,
		Name:        name,
		Slug:        slug,
		Description: description,
		StreamCount: 0,
	}

	if err := s.repository.Create(
		ctx,
		topic,
	); err != nil {
		return nil, fmt.Errorf(
			"create topic: %w",
			err,
		)
	}

	return topic, nil
}

// ============================================================
// Get Topic By ID
// ============================================================

func (s *topicService) GetByID(
	ctx context.Context,
	id string,
) (*Topic, error) {

	id = strings.TrimSpace(id)

	if !validUUID(id) {
		return nil, ErrInvalidTopic
	}

	topic, err := s.repository.GetByID(
		ctx,
		id,
	)

	if err != nil {
		if errors.Is(err, ErrTopicNotFound) {
			return nil, ErrTopicNotFound
		}

		return nil, fmt.Errorf(
			"get topic: %w",
			err,
		)
	}

	return topic, nil
}

// ============================================================
// Get Topic By Slug
// ============================================================

func (s *topicService) GetBySlug(
	ctx context.Context,
	slug string,
) (*Topic, error) {

	slug = normalizeSlug(slug)

	if slug == "" {
		return nil, ErrInvalidTopic
	}

	topic, err := s.repository.GetBySlug(
		ctx,
		slug,
	)

	if err != nil {
		if errors.Is(err, ErrTopicNotFound) {
			return nil, ErrTopicNotFound
		}

		return nil, fmt.Errorf(
			"get topic by slug: %w",
			err,
		)
	}

	return topic, nil
}

// ============================================================
// List Topics
// ============================================================

func (s *topicService) List(
	ctx context.Context,
	categoryID string,
) ([]Topic, error) {

	categoryID = strings.TrimSpace(categoryID)

	// category_id is optional.
	// Empty means return all topics.
	if categoryID != "" && !validUUID(categoryID) {
		return nil, ErrInvalidTopic
	}

	topics, err := s.repository.List(
		ctx,
		categoryID,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"list topics: %w",
			err,
		)
	}

	if topics == nil {
		topics = make([]Topic, 0)
	}

	return topics, nil
}

// ============================================================
// Update Topic
// ============================================================

func (s *topicService) Update(
	ctx context.Context,
	id string,
	req UpdateTopicRequest,
) (*Topic, error) {

	id = strings.TrimSpace(id)

	if !validUUID(id) {
		return nil, ErrInvalidTopic
	}

	// Make sure topic exists.
	existing, err := s.repository.GetByID(
		ctx,
		id,
	)

	if err != nil {
		if errors.Is(err, ErrTopicNotFound) {
			return nil, ErrTopicNotFound
		}

		return nil, fmt.Errorf(
			"get existing topic: %w",
			err,
		)
	}

	// --------------------------------------------------------
	// Name
	// --------------------------------------------------------

	if req.Name != nil {
		name := strings.TrimSpace(
			*req.Name,
		)

		if !isValidTopicName(name) {
			return nil, ErrInvalidTopic
		}

		// Only check duplicate slug if the name changed.
		if name != existing.Name {
			slug := generateSlug(name)

			other, err := s.repository.GetBySlug(
				ctx,
				slug,
			)

			if err == nil && other != nil {
				return nil, ErrTopicExists
			}

			if !errors.Is(err, ErrTopicNotFound) &&
				err != nil {

				return nil, fmt.Errorf(
					"check updated topic: %w",
					err,
				)
			}
		}

		req.Name = &name
	}

	// --------------------------------------------------------
	// Description
	// --------------------------------------------------------

	if req.Description != nil {
		description := strings.TrimSpace(
			*req.Description,
		)

		if len(description) > 500 {
			return nil, ErrInvalidTopic
		}

		req.Description = &description
	}

	// --------------------------------------------------------
	// Nothing to update
	// --------------------------------------------------------

	if req.Name == nil &&
		req.Description == nil {

		return nil, ErrInvalidTopic
	}

	// --------------------------------------------------------
	// Update
	// --------------------------------------------------------

	updated, err := s.repository.Update(
		ctx,
		id,
		req,
	)

	if err != nil {
		if errors.Is(err, ErrTopicNotFound) {
			return nil, ErrTopicNotFound
		}

		if errors.Is(err, ErrTopicExists) {
			return nil, ErrTopicExists
		}

		return nil, fmt.Errorf(
			"update topic: %w",
			err,
		)
	}

	return updated, nil
}

// ============================================================
// Delete Topic
// ============================================================

func (s *topicService) Delete(
	ctx context.Context,
	id string,
) error {

	id = strings.TrimSpace(id)

	if !validUUID(id) {
		return ErrInvalidTopic
	}

	if err := s.repository.Delete(
		ctx,
		id,
	); err != nil {

		if errors.Is(err, ErrTopicNotFound) {
			return ErrTopicNotFound
		}

		return fmt.Errorf(
			"delete topic: %w",
			err,
		)
	}

	return nil
}

// ============================================================
// Validation
// ============================================================

func isValidTopicName(
	name string,
) bool {

	name = strings.TrimSpace(name)

	if len(name) < 2 || len(name) > 100 {
		return false
	}

	for _, r := range name {
		if unicode.IsControl(r) {
			return false
		}
	}

	return true
}

// ============================================================
// UUID Validation
// ============================================================

func validUUID(
	value string,
) bool {

	_, err := uuid.Parse(value)

	return err == nil
}

// ============================================================
// Slug Generation
// ============================================================

func generateSlug(
	value string,
) string {

	value = strings.ToLower(
		strings.TrimSpace(value),
	)

	var builder strings.Builder

	lastWasDash := false

	for _, r := range value {

		switch {

		case unicode.IsLetter(r) ||
			unicode.IsDigit(r):

			builder.WriteRune(r)
			lastWasDash = false

		case r == '-' ||
			r == '_' ||
			unicode.IsSpace(r):

			if !lastWasDash &&
				builder.Len() > 0 {

				builder.WriteRune('-')
				lastWasDash = true
			}
		}
	}

	return strings.Trim(
		builder.String(),
		"-",
	)
}

// ============================================================
// Normalize Slug
// ============================================================

func normalizeSlug(
	slug string,
) string {

	return generateSlug(slug)
}
