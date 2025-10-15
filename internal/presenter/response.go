// Add missing presenter functions
package presenter

import (
	"github.com/gofiber/fiber/v2"
)

type PageMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

type API struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data,omitempty"`
	Message    string      `json:"message,omitempty"`
	Error      string      `json:"error,omitempty"`
	Pagination *PageMeta   `json:"pagination,omitempty"`
}

func OK(c *fiber.Ctx, data interface{}, pg *PageMeta) error {
	return c.Status(fiber.StatusOK).JSON(API{Success: true, Data: data, Pagination: pg})
}

// BadRequest sends a 400 Bad Request response
func BadRequest(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"success": false,
		"message": message,
		"data":    nil,
	})
}

// InternalServerError sends a 500 Internal Server Error response
func InternalServerError(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"success": false,
		"message": message,
		"data":    nil,
	})
}

// ErrorResponse sends an error response
func ErrorResponse(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"success": false,
		"message": message,
		"data":    nil,
	})
}

// Created sends a 201 Created response
func Created(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Resource created successfully",
		"data":    data,
	})
}

// Success sends a 200 OK response
func Success(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
	})
}

func SuccessResponse(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// NotImplemented sends a 501 Not Implemented response
func NotImplemented(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"success": false,
		"message": message,
		"data":    nil,
	})
}

func Err(c *fiber.Ctx, code int, msg string) error {
	return c.Status(code).JSON(API{Success: false, Error: msg})
}

func NotFound(c *fiber.Ctx, message string) error {
	return c.Status(404).JSON(fiber.Map{
		"success": false,
		"message": message,
	})
}

func ValidationErrorResponse(c *fiber.Ctx, err error) error {
	return c.Status(400).JSON(fiber.Map{
		"success": false,
		"message": "Validation failed",
		"error":   err.Error(),
	})
}

func Unauthorized(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"success": false,
		"message": message,
		"data":    nil,
	})
}

func Conflict(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusConflict).JSON(fiber.Map{
		"success": false,
		"message": message,
		"data":    nil,
	})
}
