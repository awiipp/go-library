package response

import "github.com/gofiber/fiber/v2"

type JSONResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func Success(c *fiber.Ctx, status int, data any) error {
	return c.Status(status).JSON(JSONResponse{
		Status: "success",
		Data:   data,
	})
}

func Error(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(JSONResponse{
		Status:  "error",
		Message: message,
	})
}
