package routes

import (
	"github.com/gofiber/fiber/v2"
	sharedControllers "github.com/praleedsuvarna/shared-libs/controllers"
	"github.com/praleedsuvarna/shared-libs/middleware"
)

// SetupAuditRoutes adds audit log endpoints to your application
func SetupAuditRoutes(app *fiber.App) {
	// Group routes with authentication and role checks
	auditGroup := app.Group("/audit",
		middleware.AuthMiddleware,
		middleware.AdminOnly(), // Ensure only admins can access audit logs
	)

	// Audit log endpoints
	auditGroup.Get("/logs", sharedControllers.GetAuditLogs)                       // All logs (super admin only)
	auditGroup.Get("/admin/:adminId", sharedControllers.GetAdminAuditLogs)        // Admin-specific logs
	auditGroup.Get("/resource/:targetId", sharedControllers.GetResourceAuditLogs) // Resource-specific logs
}
