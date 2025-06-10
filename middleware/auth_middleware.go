package middleware

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

// AuthMiddleware verifies the JWT token
func AuthMiddleware(c *fiber.Ctx) error {
	tokenString := c.Get("Authorization")

	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		fmt.Println(err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
	}

	userID := claims["user_id"].(string)
	organizationID := claims["organization_id"].(string)
	role, _ := claims["role"].(string)

	// Set user info in context
	c.Locals("user_id", userID)
	c.Locals("organization_id", organizationID)
	c.Locals("role", role)
	// c.Locals("user_id", claims["user_id"])
	return c.Next()
}

// AdminOnly ensures the user has admin role
func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get role from locals (set by AuthRequired middleware)
		role, ok := c.Locals("role").(string)
		if !ok || (role != "admin" && role != "super_admin") {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Admin privileges required",
			})
		}

		return c.Next()
	}
}

// SuperAdminOnly ensures the user has super_admin role
func SuperAdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get role from locals (set by AuthRequired middleware)
		role, ok := c.Locals("role").(string)
		if !ok || role != "super_admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Super admin privileges required",
			})
		}

		return c.Next()
	}
}

// AuthDebugger is a middleware that logs auth token details for debugging
func AuthDebugger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Log all headers for debugging
		fmt.Println("--- AUTH DEBUG INFO ---")
		fmt.Println("Method:", c.Method())
		fmt.Println("Path:", c.Path())

		// Log all headers
		fmt.Println("Headers:")
		c.Request().Header.VisitAll(func(key, value []byte) {
			fmt.Printf("%s: %s\n", string(key), string(value))
		})

		// Check specifically for Authorization header
		authHeader := c.Get("Authorization")
		fmt.Println("Authorization header:", authHeader)

		// Continue to next middleware/handler
		return c.Next()
	}
}
