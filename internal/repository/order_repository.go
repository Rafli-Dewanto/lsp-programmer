package repository

import (
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"cakestore/utils"
	"errors"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type OrderRepository interface {
	Create(order *entity.Order) error
	GetByID(id int64) (*entity.Order, error)
	GetAll(params *model.PaginationQuery) ([]entity.Order, *model.PaginatedMeta, error)
	GetByCustomerID(customerID int64) ([]entity.Order, error)
	Update(order *entity.Order) error
	Delete(id int64) error
	UpdateStatus(id int64, status entity.OrderStatus) error
	// GetPendingOrder retrieves the first pending order from the database for testing purposes
	GetPendingOrder() (int64, error)
}

type orderRepository struct {
	db     *gorm.DB
	logger *logrus.Logger
}

func NewOrderRepository(db *gorm.DB, logger *logrus.Logger) OrderRepository {
	return &orderRepository{
		db:     db,
		logger: logger,
	}
}

func (r *orderRepository) GetAll(params *model.PaginationQuery) ([]entity.Order, *model.PaginatedMeta, error) {
	var orders []entity.Order
	var total int64
	var meta *model.PaginatedMeta

	if params == nil {
		params = &model.PaginationQuery{}
	}
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Offset <= 0 {
		params.Offset = params.Page - 1*params.Limit
	}

	if err := r.db.Model(&entity.Order{}).Count(&total).Error; err != nil {
		r.logger.Errorf("Error getting total orders: %v", err)
		return nil, nil, err
	}

	meta = utils.CreatePaginationMeta(params.Page, params.Limit, total)

	if err := r.db.Preload("Items.Cake").
		Preload("Customer").
		Limit(int(params.Limit)).
		Offset(int((params.Page - 1) * params.Limit)).
		Find(&orders).Error; err != nil {
		r.logger.Errorf("Error getting orders: %v", err)
		return nil, nil, err
	}
	return orders, meta, nil
}

func (r *orderRepository) Create(order *entity.Order) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			r.logger.Errorf("Error creating order: %v", err)
			return err
		}
		return nil
	})
}

func (r *orderRepository) GetByID(id int64) (*entity.Order, error) {
	var order entity.Order
	if err := r.db.Preload("Items.Cake").Preload("Customer").First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		r.logger.Errorf("Error getting order by ID: %v", err)
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) GetByCustomerID(customerID int64) ([]entity.Order, error) {
	var orders []entity.Order
	if err := r.db.Preload("Customer").Preload("Items.Cake").Where("customer_id = ?", customerID).Find(&orders).Error; err != nil {
		r.logger.Errorf("Error getting orders by customer ID: %v", err)
		return nil, err
	}
	return orders, nil
}

func (r *orderRepository) Update(order *entity.Order) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(order).Error; err != nil {
			r.logger.Errorf("Error updating order: %v", err)
			return err
		}

		for _, item := range order.Items {
			if err := tx.Save(&item).Error; err != nil {
				r.logger.Errorf("Error updating order item: %v", err)
				return err
			}
		}
		return nil
	})
}

func (r *orderRepository) Delete(id int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("order_id = ?", id).Delete(&entity.OrderItem{}).Error; err != nil {
			r.logger.Errorf("Error deleting order items: %v", err)
			return err
		}

		result := tx.Delete(&entity.Order{}, id)
		if result.Error != nil {
			r.logger.Errorf("Error deleting order: %v", result.Error)
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("order not found")
		}
		return nil
	})
}

func (r *orderRepository) UpdateStatus(id int64, status entity.OrderStatus) error {
	result := r.db.Model(&entity.Order{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		r.logger.Errorf("UpdateStatus repository ~ Error updating order status: %v", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("order not found")
	}
	return nil
}

// GetPendingOrder retrieves the first pending order from the database for testing purposes
func (r *orderRepository) GetPendingOrder() (int64, error) {
	var order entity.Order
	if err := r.db.
		Preload("Items.Cake").
		Preload("Customer").
		Where("status = ?", entity.OrderStatusPending).
		Order("created_at DESC").
		First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("order not found")
		}
		r.logger.Errorf("GetPendingOrder repository ~ Error getting order: %v", err)
		return 0, err
	}
	return order.ID, nil
}
