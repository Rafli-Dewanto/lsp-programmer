package entity

import (
	"cakestore/internal/constants"
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type Payment struct {
	ID           int64                   `gorm:"column:id;primaryKey"`
	OrderID      int64                   `gorm:"column:order_id"`
	Order        Order                   `gorm:"foreignKey:OrderID"`
	Amount       float64                 `gorm:"column:amount"`
	Status       constants.PaymentStatus `gorm:"column:status"`
	PaymentToken string                  `gorm:"column:payment_token"`
	PaymentURL   string                  `gorm:"column:payment_url"`
	CreatedAt    time.Time               `gorm:"column:created_at"`
	UpdatedAt    time.Time               `gorm:"column:updated_at"`
	DeletedAt    sql.NullTime            `gorm:"column:deleted_at"`
}

func (p *Payment) TableName() string {
	return "payments"
}

func (p *Payment) BeforeCreate(tx *gorm.DB) error {
	p.Status = constants.PaymentStatusPending
	return nil
}
