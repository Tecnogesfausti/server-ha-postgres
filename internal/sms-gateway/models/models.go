package models

import (
	"time"
)

type TimedModel struct {
	CreatedAt time.Time `gorm:"not null;autoCreateTime"`
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime"`
}

type SoftDeletableModel struct {
	TimedModel

	DeletedAt *time.Time `gorm:"<-:update"`
}

type Device struct {
	SoftDeletableModel

	ID        string  `gorm:"primaryKey;type:char(21)"`
	Name      *string `gorm:"type:varchar(128)"`
	AuthToken string  `gorm:"not null;uniqueIndex;type:char(21)"`
	PushToken *string `gorm:"type:varchar(256)"`

	LastSeen time.Time `gorm:"not null;autoCreateTime;index:idx_devices_last_seen"`

	UserID string `gorm:"not null;type:varchar(32)"`
}

func NewDevice(name, pushToken *string) *Device {
	//nolint:exhaustruct // partial constructor
	return &Device{
		Name:      name,
		PushToken: pushToken,
	}
}

func (d *Device) IsEmpty() bool {
	if d == nil {
		return true
	}

	return d.ID == ""
}
