package device

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrDeviceNotFound     = errors.New("device not found")
	ErrInvalidDevice      = errors.New("invalid device")
	ErrUnauthorizedDevice = errors.New("unauthorized device")
)

type Repository interface {
	Create(
		ctx context.Context,
		device *Device,
	) error

	GetByUserAndDeviceID(
		ctx context.Context,
		userID string,
		deviceID string,
	) (*Device, error)

	ListByUser(
		ctx context.Context,
		userID string,
	) ([]Device, error)

	Update(
		ctx context.Context,
		userID string,
		deviceID string,
		req UpdateDeviceRequest,
	) (*Device, error)

	Delete(
		ctx context.Context,
		userID string,
		deviceID string,
	) error
}

type Service interface {
	Register(
		ctx context.Context,
		userID string,
		req RegisterDeviceRequest,
	) (*DeviceResponse, error)

	List(
		ctx context.Context,
		userID string,
	) (*DeviceListResponse, error)

	Update(
		ctx context.Context,
		userID string,
		deviceID string,
		req UpdateDeviceRequest,
	) (*DeviceResponse, error)

	Delete(
		ctx context.Context,
		userID string,
		deviceID string,
	) error
}

type deviceService struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &deviceService{
		repository: repository,
	}
}

// Make sure implementation satisfies Service.
var _ Service = (*deviceService)(nil)

// ============================================================
// Register Device
// ============================================================

func (s *deviceService) Register(
	ctx context.Context,
	userID string,
	req RegisterDeviceRequest,
) (*DeviceResponse, error) {

	userID = strings.TrimSpace(userID)

	if userID == "" {
		return nil, ErrInvalidDevice
	}

	deviceID := strings.TrimSpace(req.DeviceID)
	platform := strings.ToLower(strings.TrimSpace(req.Platform))
	fcmToken := strings.TrimSpace(req.FCMToken)
	appVersion := strings.TrimSpace(req.AppVersion)

	if deviceID == "" ||
		platform == "" ||
		fcmToken == "" {
		return nil, ErrInvalidDevice
	}

	if !isValidPlatform(platform) {
		return nil, ErrInvalidDevice
	}

	// Check if this device is already registered.
	existing, err := s.repository.GetByUserAndDeviceID(
		ctx,
		userID,
		deviceID,
	)

	if err == nil && existing != nil {
		// Device already exists.
		//
		// Usually this happens when the user opens the app
		// again and Firebase generates/updates its token.
		updated, updateErr := s.repository.Update(
			ctx,
			userID,
			deviceID,
			UpdateDeviceRequest{
				FCMToken:   stringPtr(fcmToken),
				AppVersion: stringPtr(appVersion),
				IsActive:   boolPtr(true),
			},
		)

		if updateErr != nil {
			return nil, fmt.Errorf(
				"update existing device: %w",
				updateErr,
			)
		}

		return toDeviceResponse(updated), nil
	}

	if !errors.Is(err, ErrDeviceNotFound) && err != nil {
		return nil, fmt.Errorf(
			"check existing device: %w",
			err,
		)
	}

	device := &Device{
		ID:         uuid.NewString(),
		UserID:     userID,
		DeviceID:   deviceID,
		Platform:   platform,
		FCMToken:   fcmToken,
		AppVersion: appVersion,
		IsActive:   true,
	}

	if err := s.repository.Create(
		ctx,
		device,
	); err != nil {
		return nil, fmt.Errorf(
			"create device: %w",
			err,
		)
	}

	return toDeviceResponse(device), nil
}

// ============================================================
// List Devices
// ============================================================

func (s *deviceService) List(
	ctx context.Context,
	userID string,
) (*DeviceListResponse, error) {

	userID = strings.TrimSpace(userID)

	if userID == "" {
		return nil, ErrInvalidDevice
	}

	devices, err := s.repository.ListByUser(
		ctx,
		userID,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"list devices: %w",
			err,
		)
	}

	responses := make(
		[]DeviceResponse,
		0,
		len(devices),
	)

	for i := range devices {
		responses = append(
			responses,
			*toDeviceResponse(&devices[i]),
		)
	}

	return &DeviceListResponse{
		Devices: responses,
		Total:   len(responses),
	}, nil
}

// ============================================================
// Update Device
// ============================================================

func (s *deviceService) Update(
	ctx context.Context,
	userID string,
	deviceID string,
	req UpdateDeviceRequest,
) (*DeviceResponse, error) {

	userID = strings.TrimSpace(userID)
	deviceID = strings.TrimSpace(deviceID)

	if userID == "" || deviceID == "" {
		return nil, ErrInvalidDevice
	}

	// Make sure the device belongs to the authenticated user.
	existing, err := s.repository.GetByUserAndDeviceID(
		ctx,
		userID,
		deviceID,
	)

	if err != nil {
		if errors.Is(err, ErrDeviceNotFound) {
			return nil, ErrDeviceNotFound
		}

		return nil, fmt.Errorf(
			"get device: %w",
			err,
		)
	}

	if existing.UserID != userID {
		return nil, ErrUnauthorizedDevice
	}

	// Normalize optional fields.
	if req.FCMToken != nil {
		token := strings.TrimSpace(*req.FCMToken)

		if token == "" {
			return nil, ErrInvalidDevice
		}

		req.FCMToken = &token
	}

	if req.AppVersion != nil {
		version := strings.TrimSpace(*req.AppVersion)

		req.AppVersion = &version
	}

	updated, err := s.repository.Update(
		ctx,
		userID,
		deviceID,
		req,
	)

	if err != nil {
		if errors.Is(err, ErrDeviceNotFound) {
			return nil, ErrDeviceNotFound
		}

		return nil, fmt.Errorf(
			"update device: %w",
			err,
		)
	}

	return toDeviceResponse(updated), nil
}

// ============================================================
// Delete Device
// ============================================================

func (s *deviceService) Delete(
	ctx context.Context,
	userID string,
	deviceID string,
) error {

	userID = strings.TrimSpace(userID)
	deviceID = strings.TrimSpace(deviceID)

	if userID == "" || deviceID == "" {
		return ErrInvalidDevice
	}

	// Verify ownership before deletion.
	existing, err := s.repository.GetByUserAndDeviceID(
		ctx,
		userID,
		deviceID,
	)

	if err != nil {
		if errors.Is(err, ErrDeviceNotFound) {
			return ErrDeviceNotFound
		}

		return fmt.Errorf(
			"get device: %w",
			err,
		)
	}

	if existing.UserID != userID {
		return ErrUnauthorizedDevice
	}

	if err := s.repository.Delete(
		ctx,
		userID,
		deviceID,
	); err != nil {
		if errors.Is(err, ErrDeviceNotFound) {
			return ErrDeviceNotFound
		}

		return fmt.Errorf(
			"delete device: %w",
			err,
		)
	}

	return nil
}

// ============================================================
// Helpers
// ============================================================

func toDeviceResponse(
	device *Device,
) *DeviceResponse {

	if device == nil {
		return nil
	}

	return &DeviceResponse{
		ID:         device.ID,
		DeviceID:   device.DeviceID,
		Platform:   device.Platform,
		AppVersion: device.AppVersion,
		IsActive:   device.IsActive,
		LastSeenAt: device.LastSeenAt,
		CreatedAt:  device.CreatedAt,
		UpdatedAt:  device.UpdatedAt,
	}
}

func isValidPlatform(platform string) bool {
	switch platform {
	case "android", "ios", "web":
		return true

	default:
		return false
	}
}

func stringPtr(value string) *string {
	return &value
}

func boolPtr(value bool) *bool {
	return &value
}
