package models

import "time"

type Pincode struct {
	ID        uint16    `json:"id" gorm:"primaryKey"`
	Pincode   uint32    `json:"pincode" gorm:"column:pincode;uniqueIndex;not null"`
	City      string    `json:"city" gorm:"column:city;not null"`
	District  string    `json:"district" gorm:"column:district;not null"`
	State     string    `json:"state" gorm:"column:state;not null"`
	Areas     string    `json:"-" gorm:"column:areas;type:text"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at;not null"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"column:updated_at;not null"`
}
