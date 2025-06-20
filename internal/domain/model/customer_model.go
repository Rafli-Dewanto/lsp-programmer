package model

import "cakestore/internal/domain/entity"

type CustomerResponse struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Address string `json:"address"`
}

type EmployeeResponse struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Address string `json:"address"`
	Role    string `json:"role"`
}

type CreateCustomerRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Address  string `json:"address" validate:"required"`
}

type UpdateUserRequest struct {
	Name    string `json:"name" validate:"required"`
	Email   string `json:"email" validate:"required,email"`
	Address string `json:"address" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func ToCustomerResponse(customer *entity.Customer) *CustomerResponse {
	return &CustomerResponse{
		ID:      customer.ID,
		Name:    customer.Name,
		Email:   customer.Email,
		Address: customer.Address,
	}
}

func ToEmployeeResponse(customer *entity.Customer) *EmployeeResponse {
	return &EmployeeResponse{
		ID:      customer.ID,
		Name:    customer.Name,
		Email:   customer.Email,
		Address: customer.Address,
		Role:    customer.Role,
	}
}

func ToLoginResponse(token string) *LoginResponse {
	return &LoginResponse{
		Token: token,
	}
}
