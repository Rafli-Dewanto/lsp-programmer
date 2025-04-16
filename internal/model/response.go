package model

import (
	"cakestore/utils"

	"github.com/gofiber/fiber/v2"
)

type Response struct {
	Message string         `json:"message"`
	Data    any            `json:"data,omitempty"`
	Meta    *PaginatedMeta `json:"meta,omitempty"`
}

type FailedResponse struct {
	Message string `json:"message"`
	Errors  any    `json:"errors"`
}

func WriteResponse(c *fiber.Ctx, statusCode int, data interface{}, message string, meta *PaginatedMeta) error {
	if data != nil {
		if errors, ok := data.([]utils.ValidationError); ok {
			resp := FailedResponse{
				Message: message,
				Errors:  errors,
			}
			return c.Status(statusCode).JSON(resp)
		}
	}

	resp := Response{
		Data:    data,
		Message: message,
		Meta:    meta,
	}

	return c.Status(statusCode).JSON(resp)
}

func WriteErrorResponse(c *fiber.Ctx, statusCode int, message string) error {
	return WriteResponse(c, statusCode, nil, message, nil)
}

func WriteValidationErrorResponse(c *fiber.Ctx, errors []utils.ValidationError) error {
	return WriteResponse(c, fiber.StatusBadRequest, errors, "Validation failed", nil)
}
