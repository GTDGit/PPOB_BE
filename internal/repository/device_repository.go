package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/GTDGit/PPOB_BE/internal/domain"
)

type DeviceRepository interface {
	FindByID(ctx context.Context, id string) (*domain.Device, error)
	FindByUserIDAndDeviceID(ctx context.Context, userID, deviceID string) (*domain.Device, error)
	FindByUserID(ctx context.Context, userID string) ([]*domain.Device, error)
	Create(ctx context.Context, device *domain.Device) error
	Update(ctx context.Context, device *domain.Device) error
	UpdateLastActive(ctx context.Context, id string, location, ipAddress string) error
	Delete(ctx context.Context, id string) error
	DeleteByUserIDAndDeviceID(ctx context.Context, userID, deviceID string) error
	CountByUserID(ctx context.Context, userID string) (int, error)
	DeleteInactiveDevices(ctx context.Context, userID string, inactiveDays int) error
}

type deviceRepository struct {
	db *sqlx.DB
}

func NewDeviceRepository(db *sqlx.DB) DeviceRepository {
	return &deviceRepository{db: db}
}

// Column constants for explicit SELECT
const deviceColumns = `id, user_id, device_id, device_name, platform, 
                       last_active_at, location, ip_address, is_active, created_at`

func (r *deviceRepository) FindByID(ctx context.Context, id string) (*domain.Device, error) {
	var device domain.Device
	query := fmt.Sprintf(`SELECT %s FROM devices WHERE id = $1 AND is_active = true`, deviceColumns)
	err := r.db.GetContext(ctx, &device, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *deviceRepository) FindByUserIDAndDeviceID(ctx context.Context, userID, deviceID string) (*domain.Device, error) {
	var device domain.Device
	query := fmt.Sprintf(`SELECT %s FROM devices WHERE user_id = $1 AND device_id = $2 AND is_active = true`, deviceColumns)
	err := r.db.GetContext(ctx, &device, query, userID, deviceID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *deviceRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.Device, error) {
	var devices []*domain.Device
	query := fmt.Sprintf(`SELECT %s FROM devices WHERE user_id = $1 AND is_active = true ORDER BY last_active_at DESC NULLS LAST`, deviceColumns)
	err := r.db.SelectContext(ctx, &devices, query, userID)
	if err != nil {
		return nil, err
	}
	return devices, nil
}

func (r *deviceRepository) Create(ctx context.Context, device *domain.Device) error {
	if device.ID == "" {
		device.ID = "dev_" + uuid.New().String()[:8]
	}
	device.CreatedAt = time.Now()
	device.IsActive = true

	query := `
		INSERT INTO devices (id, user_id, device_id, device_name, platform, last_active_at, location, ip_address, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (user_id, device_id) DO UPDATE SET
			device_name = EXCLUDED.device_name,
			platform = EXCLUDED.platform,
			last_active_at = EXCLUDED.last_active_at,
			location = EXCLUDED.location,
			ip_address = EXCLUDED.ip_address,
			is_active = true
	`
	now := time.Now()
	device.LastActiveAt = &now

	_, err := r.db.ExecContext(ctx, query,
		device.ID, device.UserID, device.DeviceID, device.DeviceName,
		device.Platform, device.LastActiveAt, device.Location, device.IPAddress,
		device.IsActive, device.CreatedAt,
	)
	return err
}

func (r *deviceRepository) Update(ctx context.Context, device *domain.Device) error {
	query := `
		UPDATE devices SET
			device_name = $2, platform = $3, last_active_at = $4, location = $5, ip_address = $6, is_active = $7
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query,
		device.ID, device.DeviceName, device.Platform, device.LastActiveAt,
		device.Location, device.IPAddress, device.IsActive,
	)
	return err
}

func (r *deviceRepository) UpdateLastActive(ctx context.Context, id string, location, ipAddress string) error {
	query := `UPDATE devices SET last_active_at = $2, location = $3, ip_address = $4 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, time.Now(), location, ipAddress)
	return err
}

func (r *deviceRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE devices SET is_active = false WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *deviceRepository) DeleteByUserIDAndDeviceID(ctx context.Context, userID, deviceID string) error {
	query := `UPDATE devices SET is_active = false WHERE user_id = $1 AND device_id = $2`
	_, err := r.db.ExecContext(ctx, query, userID, deviceID)
	return err
}

func (r *deviceRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM devices WHERE user_id = $1 AND is_active = true`
	err := r.db.GetContext(ctx, &count, query, userID)
	return count, err
}

func (r *deviceRepository) DeleteInactiveDevices(ctx context.Context, userID string, inactiveDays int) error {
	query := `UPDATE devices SET is_active = false WHERE user_id = $1 AND last_active_at < $2`
	cutoff := time.Now().AddDate(0, 0, -inactiveDays)
	_, err := r.db.ExecContext(ctx, query, userID, cutoff)
	return err
}
