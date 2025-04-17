package usecase

import (
	"bytes"
	"cakestore/internal/constants"
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"cakestore/internal/repository"
	"cakestore/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/midtrans/midtrans-go"
	"github.com/sirupsen/logrus"
)

type PaymentUseCase interface {
	CreatePaymentURL(order *entity.Order) (*model.PaymentResponse, error)
	GetOrderStatus(orderID string) (string, error)
	UpdateOrderStatus(id string, status constants.PaymentStatus) error
}

type paymentUseCase struct {
	paymentRepository repository.PaymentRepository
	endpoint          string
	log               *logrus.Logger
	env               string
}

func NewPaymentUseCase(
	endpoint string,
	paymentRepository repository.PaymentRepository,
	log *logrus.Logger,
	env string,
) PaymentUseCase {
	return &paymentUseCase{
		endpoint:          endpoint,
		paymentRepository: paymentRepository,
		log:               log,
		env:               env,
	}
}

func (uc *paymentUseCase) CreatePaymentURL(order *entity.Order) (*model.PaymentResponse, error) {
	var req model.CreatePaymentRequest

	req.TransactionDetails = midtrans.TransactionDetails{
		OrderID:  strconv.Itoa(order.ID),
		GrossAmt: int64(order.TotalPrice),
	}

	headers := utils.GenerateRequestHeader()

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", uc.endpoint+"/snap/v1/transactions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create payment URL, status code: %d", resp.StatusCode)
	}

	var paymentResponse model.PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResponse); err != nil {
		return nil, err
	}

	// insert payment to db
	payment := &entity.Payment{
		OrderID:      order.ID,
		Amount:       order.TotalPrice,
		Status:       constants.PaymentStatusPending,
		PaymentToken: paymentResponse.Token,
		PaymentURL:   paymentResponse.RedirectURL,
	}
	if err := uc.paymentRepository.CreatePayment(payment); err != nil {
		return nil, err
	}

	return &paymentResponse, nil
}

func (uc *paymentUseCase) GetOrderStatus(orderID string) (string, error) {
	endpoint := fmt.Sprintf("%s/v2/%s/status", uc.endpoint, orderID)
	headers := utils.GenerateRequestHeader()

	httpReq, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get order status, status code: %d", resp.StatusCode)
	}

	var orderStatus model.GetOrderStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&orderStatus); err != nil {
		return "", err
	}

	if orderStatus.StatusCode != "200" {
		return "", fmt.Errorf("failed to get order status, status code: %s", orderStatus.StatusCode)
	}

	return orderStatus.TransactionStatus, nil
}

func (uc *paymentUseCase) UpdateOrderStatus(id string, status constants.PaymentStatus) error {
	if uc.env == "development" {
		orderId, err := uc.paymentRepository.GetPendingPayment()
		if err != nil {
			uc.log.Errorf("Error getting pending payment: %v", err)
			return err
		}
		payment := model.ToPaymentEntity(&model.PaymentModel{
			OrderID: orderId,
			Status:  status,
		})
		if err := uc.paymentRepository.UpdatePayment(payment); err != nil {
			uc.log.Errorf("Error updating order status: %v", err)
			return err
		}
		return nil
	}

	uc.log.Info("Running in production mode, updating payment status")
	orderID, err := strconv.Atoi(id)
	if err != nil {
		return err
	}

	payment := model.ToPaymentEntity(&model.PaymentModel{
		OrderID: orderID,
		Status:  status,
	})

	if err := uc.paymentRepository.UpdatePayment(payment); err != nil {
		uc.log.Errorf("Error updating order status: %v", err)
		return err
	}
	return nil
}
