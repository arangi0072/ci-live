package category

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/google/uuid"
)

var (
	ErrCategoryNotFound = errors.New("category not found")
	ErrCategoryExists   = errors.New("category already exists")
	ErrInvalidCategory  = errors.New("invalid category")
)

type Repository interface {
	Create(
		ctx context.Context,
		category *Category,
	) error

	GetByID(
		ctx context.Context,
		id string,
	) (*Category, error)

	GetBySlug(
		ctx context.Context,
		slug string,
	) (*Category, error)

	List(
		ctx context.Context,
	) ([]Category, error)

	Update(
		ctx context.Context,
		id string,
		req UpdateCategoryRequest,
	) (*Category, error)

	Delete(
		ctx context.Context,
		id string,
	) error
}

type Service interface {
	Create(
		ctx context.Context,
		req CreateCategoryRequest,
	) (*Category, error)

	GetByID(
		ctx context.Context,
		id string,
	) (*Category, error)

	GetBySlug(
		ctx context.Context,
		slug string,
	) (*Category, error)

	List(
		ctx context.Context,
	) ([]Category, error)

	Update(
		ctx context.Context,
		id string,
		req UpdateCategoryRequest,
	) (*Category, error)

	Delete(
		ctx context.Context,
		id string,
	) error
}

type categoryService struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &categoryService{
		repository: repository,
	}
}

var _ Service = (*categoryService)(nil)

// ============================================================
// Create Category
// ============================================================

func (s *categoryService) Create(
	ctx context.Context,
	req CreateCategoryRequest,
) (*Category, error) {

	name := strings.TrimSpace(req.Name)

	if !isValidCategoryName(name) {
		return nil, ErrInvalidCategory
	}

	slug := generateSlug(name)

	if slug == "" {
		return nil, ErrInvalidCategory
	}

	// Check whether slug already exists.
	existing, err := s.repository.GetBySlug(
		ctx,
		slug,
	)

	if err == nil && existing != nil {
		return nil, ErrCategoryExists
	}

	if !errors.Is(err, ErrCategoryNotFound) && err != nil {
		return nil, fmt.Errorf(
			"check category existence: %w",
			err,
		)
	}

	description := strings.TrimSpace(
		req.Description,
	)

	iconURL := strings.TrimSpace(
		req.IconURL,
	)

	category := &Category{
		ID:          uuid.NewString(),
		Name:        name,
		Slug:        slug,
		Description: description,
		IconURL:     iconURL,
		StreamCount: 0,
	}

	if err := s.repository.Create(
		ctx,
		category,
	); err != nil {
		return nil, fmt.Errorf(
			"create category: %w",
			err,
		)
	}

	return category, nil
}

// ============================================================
// Get Category by ID
// ============================================================

func (s *categoryService) GetByID(
	ctx context.Context,
	id string,
) (*Category, error) {

	id = strings.TrimSpace(id)

	if id == "" {
		return nil, ErrInvalidCategory
	}

	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidCategory
	}

	category, err := s.repository.GetByID(
		ctx,
		id,
	)

	if err != nil {
		if errors.Is(err, ErrCategoryNotFound) {
			return nil, ErrCategoryNotFound
		}

		return nil, fmt.Errorf(
			"get category: %w",
			err,
		)
	}

	return category, nil
}

// ============================================================
// Get Category by Slug
// ============================================================

func (s *categoryService) GetBySlug(
	ctx context.Context,
	slug string,
) (*Category, error) {

	slug = normalizeSlug(slug)

	if slug == "" {
		return nil, ErrInvalidCategory
	}

	category, err := s.repository.GetBySlug(
		ctx,
		slug,
	)

	if err != nil {
		if errors.Is(err, ErrCategoryNotFound) {
			return nil, ErrCategoryNotFound
		}

		return nil, fmt.Errorf(
			"get category by slug: %w",
			err,
		)
	}

	return category, nil
}

// ============================================================
// List Categories
// ============================================================

func (s *categoryService) List(
	ctx context.Context,
) ([]Category, error) {

	categories, err := s.repository.List(
		ctx,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"list categories: %w",
			err,
		)
	}

	if categories == nil {
		categories = make([]Category, 0)
	}

	return categories, nil
}

// ============================================================
// Update Category
// ============================================================

func (s *categoryService) Update(
	ctx context.Context,
	id string,
	req UpdateCategoryRequest,
) (*Category, error) {

	id = strings.TrimSpace(id)

	if id == "" {
		return nil, ErrInvalidCategory
	}

	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidCategory
	}

	// Make sure category exists.
	existing, err := s.repository.GetByID(
		ctx,
		id,
	)

	if err != nil {
		if errors.Is(err, ErrCategoryNotFound) {
			return nil, ErrCategoryNotFound
		}

		return nil, fmt.Errorf(
			"get existing category: %w",
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

		if !isValidCategoryName(name) {
			return nil, ErrInvalidCategory
		}

		// If name didn't change, don't need to check uniqueness.
		if name != existing.Name {
			slug := generateSlug(name)

			other, err := s.repository.GetBySlug(
				ctx,
				slug,
			)

			if err == nil && other != nil {
				return nil, ErrCategoryExists
			}

			if !errors.Is(err, ErrCategoryNotFound) &&
				err != nil {

				return nil, fmt.Errorf(
					"check updated category slug: %w",
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
			return nil, ErrInvalidCategory
		}

		req.Description = &description
	}

	// --------------------------------------------------------
	// Icon URL
	// --------------------------------------------------------

	if req.IconURL != nil {
		iconURL := strings.TrimSpace(
			*req.IconURL,
		)

		if len(iconURL) > 2048 {
			return nil, ErrInvalidCategory
		}

		req.IconURL = &iconURL
	}

	// Nothing to update.
	if req.Name == nil &&
		req.Description == nil &&
		req.IconURL == nil {

		return nil, ErrInvalidCategory
	}

	updated, err := s.repository.Update(
		ctx,
		id,
		req,
	)

	if err != nil {
		if errors.Is(err, ErrCategoryNotFound) {
			return nil, ErrCategoryNotFound
		}

		if errors.Is(err, ErrCategoryExists) {
			return nil, ErrCategoryExists
		}

		return nil, fmt.Errorf(
			"update category: %w",
			err,
		)
	}

	return updated, nil
}

// ============================================================
// Delete Category
// ============================================================

func (s *categoryService) Delete(
	ctx context.Context,
	id string,
) error {

	id = strings.TrimSpace(id)

	if id == "" {
		return ErrInvalidCategory
	}

	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidCategory
	}

	if err := s.repository.Delete(
		ctx,
		id,
	); err != nil {

		if errors.Is(err, ErrCategoryNotFound) {
			return ErrCategoryNotFound
		}

		return fmt.Errorf(
			"delete category: %w",
			err,
		)
	}

	return nil
}

// ============================================================
// Validation
// ============================================================

func isValidCategoryName(
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
// Slug generation
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

func normalizeSlug(
	slug string,
) string {

	return generateSlug(slug)
}
