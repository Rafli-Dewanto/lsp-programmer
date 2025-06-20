package repository

import (
	"cakestore/internal/constants"
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type CustomerRepository interface {
	Create(customer *entity.Customer) error
	GetByID(id int64) (*entity.Customer, error)
	GetByEmail(email string) (*entity.Customer, error)
	Update(customer *entity.Customer) error
	Delete(id int64) error
	GetEmployees() ([]entity.Customer, error)
	GetEmployeeByID(id int64) (*entity.Customer, error)
	UpdateEmployee(id int64, request *model.UpdateUserRequest, role string) error
	DeleteEmployee(id int64) error
}

type customerRepository struct {
	db     *gorm.DB
	logger *logrus.Logger
}

func NewCustomerRepository(db *gorm.DB, logger *logrus.Logger) CustomerRepository {
	return &customerRepository{
		db:     db,
		logger: logger,
	}
}

func (r *customerRepository) GetEmployees() ([]entity.Customer, error) {
	var customer []entity.Customer
	if err := r.db.Where("role != ?", "customer").Find(&customer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("employee not found")
		}
		r.logger.Errorf("Error getting employee: %v", err)
		return nil, err
	}
	return customer, nil
}

func (r *customerRepository) GetEmployeeByID(id int64) (*entity.Customer, error) {
	var customer entity.Customer
	if err := r.db.
		Where("id = ? AND role != ?", id, "customer").
		First(&customer).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, constants.ErrNotFound
		}
		r.logger.Errorf("Error getting employee by ID: %v", err)
		return nil, err
	}
	return &customer, nil
}

func (r *customerRepository) UpdateEmployee(id int64, request *model.UpdateUserRequest, role string) error {
	customer, err := r.GetEmployeeByID(id)
	if err != nil {
		return err
	}
	customer.Name = request.Name
	customer.Address = request.Address
	customer.UpdatedAt = time.Now()
	customer.Email = request.Email

	// update role if provided
	if role != "" {
		customer.Role = role

	}
	if err := r.db.Save(customer).Error; err != nil {
		r.logger.Errorf("Error updating employee: %v", err)
		return err
	}
	return nil
}

func (r *customerRepository) DeleteEmployee(id int64) error {
	result := r.db.Delete(&entity.Customer{}, id)
	if result.Error != nil {
		r.logger.Errorf("Error deleting employee: %v", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("employee not found")
	}
	return nil
}

func (r *customerRepository) Create(customer *entity.Customer) error {
	if err := r.db.Create(customer).Error; err != nil {
		r.logger.Errorf("Error creating customer: %v", err)
		return err
	}
	return nil
}

func (r *customerRepository) GetByID(id int64) (*entity.Customer, error) {
	var customer entity.Customer
	if err := r.db.First(&customer, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("customer not found")
		}
		r.logger.Errorf("Error getting customer by ID: %v", err)
		return nil, err
	}
	return &customer, nil
}

func (r *customerRepository) GetByEmail(email string) (*entity.Customer, error) {
	var customer entity.Customer
	if err := r.db.Where("email = ?", email).First(&customer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("customer not found")
		}
		r.logger.Errorf("Error getting customer by email: %v", err)
		return nil, err
	}
	return &customer, nil
}

func (r *customerRepository) Update(customer *entity.Customer) error {
	if err := r.db.Save(customer).Error; err != nil {
		r.logger.Errorf("Error updating customer: %v", err)
		return err
	}
	return nil
}

func (r *customerRepository) Delete(id int64) error {
	result := r.db.Delete(&entity.Customer{}, id)
	if result.Error != nil {
		r.logger.Errorf("Error deleting customer: %v", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("customer not found")
	}
	return nil
}
