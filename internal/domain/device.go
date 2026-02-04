package domain

import (
	"time"
)

// Device represents a user's device
type Device struct {
	ID           string     `db:"id" json:"id"`
	UserID       string     `db:"user_id" json:"userId"`
	DeviceID     string     `db:"device_id" json:"deviceId"`
	DeviceName   string     `db:"device_name" json:"deviceName"`
	Platform     *string    `db:"platform" json:"platform"`
	LastActiveAt *time.Time `db:"last_active_at" json:"lastActiveAt"`
	Location     *string    `db:"location" json:"location"`
	IPAddress    *string    `db:"ip_address" json:"ipAddress"`
	IsActive     bool       `db:"is_active" json:"isActive"`
	CreatedAt    time.Time  `db:"created_at" json:"createdAt"`
}

// DeviceResponse for API response
type DeviceResponse struct {
	DeviceID        string  `json:"deviceId"`
	DeviceName      string  `json:"deviceName"`
	LastActiveAt    *string `json:"lastActiveAt"`
	IsCurrentDevice bool    `json:"isCurrentDevice"`
	Location        *string `json:"location"`
	IPAddress       *string `json:"ipAddress"`
}

// ToResponse converts Device to DeviceResponse
func (d *Device) ToResponse(currentDeviceID string) *DeviceResponse {
	var lastActiveAt *string
	if d.LastActiveAt != nil {
		formatted := d.LastActiveAt.Format(time.RFC3339)
		lastActiveAt = &formatted
	}

	// Mask IP address for security
	var maskedIP *string
	if d.IPAddress != nil {
		ip := *d.IPAddress
		if len(ip) > 6 {
			masked := ip[:3] + ".xxx.xxx." + ip[len(ip)-3:]
			maskedIP = &masked
		} else {
			maskedIP = d.IPAddress
		}
	}

	return &DeviceResponse{
		DeviceID:        d.DeviceID,
		DeviceName:      d.DeviceName,
		LastActiveAt:    lastActiveAt,
		IsCurrentDevice: d.DeviceID == currentDeviceID,
		Location:        d.Location,
		IPAddress:       maskedIP,
	}
}

// Platform enum
const (
	PlatformAndroid = "android"
	PlatformIOS     = "ios"
	PlatformWeb     = "web"
)
