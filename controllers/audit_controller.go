package controllers

import (

	// "UserManagement/utils"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/praleedsuvarna/shared-libs/utils"
	"go.mongodb.org/mongo-driver/bson"
)

// GetAuditLogs retrieves all audit logs (for super admin)
func GetAuditLogs(c *fiber.Ctx) error {
	logs, err := utils.GetAuditLogs(bson.M{})
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch audit logs",
		})
	}

	return c.JSON(logs)
}

// GetAdminAuditLogs retrieves audit logs for a specific admin
func GetAdminAuditLogs(c *fiber.Ctx) error {
	adminID := c.Params("adminId")
	if adminID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Admin ID is required",
		})
	}

	logs, err := utils.GetAuditLogs(bson.M{"admin_id": adminID})
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch admin audit logs",
		})
	}

	return c.JSON(logs)
}

// GetResourceAuditLogs retrieves audit logs for a specific resource/target
func GetResourceAuditLogs(c *fiber.Ctx) error {
	targetID := c.Params("targetId")
	if targetID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Target/Resource ID is required",
		})
	}

	logs, err := utils.GetAuditLogs(bson.M{"target_id": targetID})
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch resource audit logs",
		})
	}

	return c.JSON(logs)
}
