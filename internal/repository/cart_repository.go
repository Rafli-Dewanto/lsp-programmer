package repository

import (
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type CartRepository interface {
	Create(cart *entity.Cart) error
	GetByID(id int) (*entity.Cart, error)
	GetByCustomerID(customerID int, params *model.PaginationQuery) (*model.PaginationResponse, error)
	GetByCustomerIDAndCakeID(customerID int, cakeID int) (*entity.Cart, error)
	Update(cart *entity.Cart) error
	Delete(cartID int) error
	RemoveItem(customerID int, cartID int) error
	ClearCustomerCart(customerID int) error
}

type cartRepository struct {
	db     *gorm.DB
	logger *logrus.Logger
}

func NewCartRepository(db *gorm.DB, logger *logrus.Logger) CartRepository {
	return &cartRepository{
		db:     db,
		logger: logger,
	}
}

func (r *cartRepository) Create(cart *entity.Cart) error {
	return r.db.Create(cart).Error
}

func (r *cartRepository) GetByID(id int) (*entity.Cart, error) {
	var cart entity.Cart
	if err := r.db.First(&cart, id).Error; err != nil {
		return nil, err
	}
	return &cart, nil
}

func (r *cartRepository) GetByCustomerID(customerID int, params *model.PaginationQuery) (*model.PaginationResponse, error) {
	var carts []*entity.Cart
	var total int64
	var perPage int
	var page int
	if params.Page < 1 {
		page = 1
	}

	query := r.db.Model(&entity.Cart{}).Where("customer_id = ?", customerID)
	query.Count(&total)

	if params.Limit > 0 {
		perPage = int(params.Limit)
	} else {
		perPage = 10
	}

	offSet := (params.Page - 1) * perPage

	if err := query.Offset(offSet).Limit(perPage).Find(&carts).Error; err != nil {
		return nil, err
	}

	return &model.PaginationResponse{
		Total:      total,
		Data:       carts,
		Page:       page,
		PageSize:   perPage,
		TotalPages: int(total) / perPage,
	}, nil
}

func (r *cartRepository) GetByCustomerIDAndCakeID(customerID int, cakeID int) (*entity.Cart, error) {
	var cart entity.Cart
	if err := r.db.Where("customer_id = ? AND cake_id = ?", customerID, cakeID).First(&cart).Error; err != nil {
		return nil, err
	}
	return &cart, nil
}

func (r *cartRepository) Update(cart *entity.Cart) error {
	return r.db.Save(cart).Error
}

func (r *cartRepository) Delete(cartID int) error {
	return r.db.Delete(&entity.Cart{}, cartID).Error
}

func (r *cartRepository) RemoveItem(customerID int, cartID int) error {
	type result struct {
		Quantity int
		Subtotal float64
	}

	var res result

	// Retrieve quantity and subtotal
	if err := r.db.
		Model(&entity.Cart{}).
		Where("id = ? AND customer_id = ?", cartID, customerID).
		Select("quantity", "subtotal").
		Scan(&res).Error; err != nil {
		return err
	}

	// If only 1 item, delete the cart item
	if res.Quantity <= 1 {
		if err := r.db.
			Where("id = ? AND customer_id = ?", cartID, customerID).
			Delete(&entity.Cart{}).Error; err != nil {
			return err
		}
	} else {
		// Update: decrement quantity and subtotal
		newSubtotal := res.Subtotal / float64(res.Quantity)
		if err := r.db.
			Model(&entity.Cart{}).
			Where("id = ? AND customer_id = ?", cartID, customerID).
			Updates(map[string]interface{}{
				"quantity": gorm.Expr("quantity - ?", 1),
				"subtotal": gorm.Expr("subtotal - ?", newSubtotal),
			}).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *cartRepository) ClearCustomerCart(customerID int) error {
	result := r.db.Where("customer_id = ?", customerID).Delete(&entity.Cart{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
