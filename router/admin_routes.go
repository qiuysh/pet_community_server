package router

import (
	"pet-community/controller"
	"pet-community/middleware"

	"github.com/gin-gonic/gin"
)

// AdminControllers 所有按域拆分后的 admin 子 controller 聚合结构。
// 用于替换原单一 *controller.AdminController。
type AdminControllers struct {
	Auth         *controller.AdminAuthController
	Users        *controller.AdminUsersController
	FrontUsers   *controller.AdminFrontUsersController
	Settings     *controller.AdminSettingsController
	Logs         *controller.AdminLogsController
	Review       *controller.AdminReviewController
	Report       *controller.AdminReportController
	Stats        *controller.AdminStatsController
	Dashboard    *controller.AdminDashboardController
	Notification *controller.AdminNotificationController
	Announcement *controller.AdminAnnouncementController
	Ad           *controller.AdminAdController
}

// RegisterAdminRoutes 注册所有 admin 路由,并按域挂载 RBAC RequireRole 中间件。
// 原:全部路由只挂 AuthMiddleware() 只认证不鉴权,B3 RBAC 是死代码。
// 现:AuthMiddleware + 按域 RequireRole(...),role.go 从死代码变活代码。
func RegisterAdminRoutes(router *gin.Engine, ac AdminControllers) {
	// ===== 公开: 登录 + 健康检查 =====
	adminPublic := router.Group("/api/v1/admin")
	{
		adminPublic.POST("/login", ac.Auth.Login)
	}

	// ===== 受保护基类: 所有 admin 必须先过 AuthMiddleware(认证) =====
	protected := router.Group("/api/v1/admin")
	protected.Use(middleware.AuthMiddleware())
	{
		// ===== 运营驾驶舱: 全部角色（Ad5；service 侧重待办） =====
		protected.GET("/dashboard/overview", ac.Dashboard.GetDashboardOverview)
		protected.GET("/dashboard/todo", ac.Dashboard.GetTodoAggregate)

		// ===== 日志只读: super_admin | data_admin | service（Ad8；content_admin 不可见） =====
		adminLogsRead := protected.Group("")
		adminLogsRead.Use(middleware.RequireRole(
			middleware.RoleSuperAdmin, middleware.RoleDataAdmin, middleware.RoleService,
		))
		{
			adminLogsRead.GET("/logs/operations", ac.Logs.GetOperationLogs)
			adminLogsRead.GET("/logs/activities", ac.Logs.GetActivityLogs)
			adminLogsRead.GET("/logs/login", ac.Logs.GetLoginLogs)
			adminLogsRead.GET("/logs/behavior", ac.Logs.GetUserBehaviorLogs)
			adminLogsRead.GET("/logs/operations-v2", ac.Logs.GetOperationLogsV2)
			adminLogsRead.GET("/logs/security", ac.Logs.GetSecurityAuditLogs)
		}
		// ===== 日志导出: super_admin | data_admin =====
		adminLogsWrite := protected.Group("")
		adminLogsWrite.Use(middleware.RequireRole(middleware.RoleSuperAdmin, middleware.RoleDataAdmin))
		{
			adminLogsWrite.GET("/logs/export", ac.Logs.ExportLogs)
		}

		// ===== Admin 账号管理: super_admin only =====
		adminUsers := protected.Group("")
		adminUsers.Use(middleware.RequireRole(middleware.RoleSuperAdmin))
		{
			adminUsers.GET("/users", ac.Users.ListAdmins)
			adminUsers.GET("/users/:id", ac.Users.GetAdmin)
			adminUsers.POST("/users", ac.Users.CreateAdmin)
			adminUsers.PUT("/users/:id", ac.Users.UpdateAdmin)
			adminUsers.PUT("/users/:id/password", ac.Users.UpdateAdminPassword)
			adminUsers.DELETE("/users/:id", ac.Users.DeleteAdmin)
			adminUsers.PUT("/users/:id/status", ac.Users.UpdateAdminStatus)
		}

		// ===== 站点/审核配置: super_admin | content_admin（Ad0-3） =====
		adminConfig := protected.Group("")
		adminConfig.Use(middleware.RequireRole(middleware.RoleSuperAdmin, middleware.RoleContentAdmin))
		{
			adminConfig.GET("/settings/site", ac.Settings.GetSiteConfig)
			adminConfig.PUT("/settings/site", ac.Settings.UpdateSiteConfig)
			adminConfig.GET("/settings/audit", ac.Settings.GetAuditConfig)
			adminConfig.PUT("/settings/audit", ac.Settings.UpdateAuditConfig)
		}

		// ===== 敏感词: super_admin | content_admin =====
		adminSensitive := protected.Group("")
		adminSensitive.Use(middleware.RequireRole(middleware.RoleSuperAdmin, middleware.RoleContentAdmin))
		{
			adminSensitive.GET("/sensitive-words", ac.Settings.GetSensitiveWords)
			adminSensitive.POST("/sensitive-words", ac.Settings.AddSensitiveWord)
			adminSensitive.PUT("/sensitive-words/:id", ac.Settings.UpdateSensitiveWord)
			adminSensitive.DELETE("/sensitive-words/:id", ac.Settings.DeleteSensitiveWord)
		}

		// ===== 前台用户管理(封禁等): super_admin | service =====
		adminFront := protected.Group("")
		adminFront.Use(middleware.RequireRole(middleware.RoleSuperAdmin, middleware.RoleService))
		{
			adminFront.GET("/front-users", ac.FrontUsers.ListFrontUsers)
			adminFront.GET("/front-users/:id", ac.FrontUsers.GetFrontUser)
			adminFront.PUT("/front-users/:id/ban", ac.FrontUsers.BanFrontUser)
			adminFront.PUT("/front-users/:id/unban", ac.FrontUsers.UnbanFrontUser)
			adminFront.POST("/front-users/batch-ban", ac.FrontUsers.BatchBanFrontUsers)
			adminFront.GET("/front-users/logs", ac.FrontUsers.GetFrontUserOperationLogs)
		}

		// ===== 内容审核只读: super_admin | content_admin | service（Ad1 service 只读） =====
		adminReviewRead := protected.Group("")
		adminReviewRead.Use(middleware.RequireRole(
			middleware.RoleSuperAdmin, middleware.RoleContentAdmin, middleware.RoleService,
		))
		{
			adminReviewRead.GET("/reviews/pending", ac.Review.GetPendingReviews)
			adminReviewRead.GET("/reviews/history", ac.Review.GetReviewHistory)
		}
		// ===== 内容审核写: super_admin | content_admin =====
		adminReviewWrite := protected.Group("")
		adminReviewWrite.Use(middleware.RequireRole(middleware.RoleSuperAdmin, middleware.RoleContentAdmin))
		{
			adminReviewWrite.POST("/reviews/approve", ac.Review.ApproveReview)
			adminReviewWrite.POST("/reviews/reject", ac.Review.RejectReview)
			adminReviewWrite.POST("/reviews/batch", ac.Review.BatchReview)
		}

		// ===== 举报管理 Ad4: super_admin | content_admin | service =====
		adminReports := protected.Group("")
		adminReports.Use(middleware.RequireRole(
			middleware.RoleSuperAdmin, middleware.RoleContentAdmin, middleware.RoleService,
		))
		{
			adminReports.GET("/reports", ac.Report.ListReports)
			adminReports.GET("/reports/stats", ac.Report.GetReportStats)
			adminReports.POST("/reports/batch", ac.Report.BatchHandleReports)
			adminReports.POST("/reports/:id/handle", ac.Report.HandleReport)
		}

		// ===== 数据统计/报表: super_admin | data_admin =====
		adminStats := protected.Group("")
		adminStats.Use(middleware.RequireRole(middleware.RoleSuperAdmin, middleware.RoleDataAdmin))
		{
			adminStats.GET("/stats/overview", ac.Stats.GetOverviewStats)
			adminStats.GET("/stats/user-growth", ac.Stats.GetUserGrowthStats)
			adminStats.GET("/stats/content", ac.Stats.GetContentStats)
			adminStats.GET("/stats/interaction", ac.Stats.GetInteractionStats)
			adminStats.GET("/stats/trend", ac.Stats.GetTrendData)
			adminStats.GET("/stats/comparison", ac.Stats.GetComparisonStats)
			adminStats.GET("/stats/distribution", ac.Stats.GetDistributionStats)
			adminStats.POST("/stats/export", ac.Stats.ExportStats)
		}

		// ===== 通知管理: super_admin | content_admin | service（Ad7） =====
		adminNotify := protected.Group("")
		adminNotify.Use(middleware.RequireRole(
			middleware.RoleSuperAdmin, middleware.RoleContentAdmin, middleware.RoleService,
		))
		{
			adminNotify.POST("/notifications", ac.Notification.CreateNotification)
			adminNotify.PUT("/notifications/:id", ac.Notification.UpdateNotification)
			adminNotify.DELETE("/notifications/:id", ac.Notification.DeleteNotification)
			adminNotify.GET("/notifications/:id", ac.Notification.GetNotification)
			adminNotify.GET("/notifications", ac.Notification.ListNotifications)
			adminNotify.POST("/notifications/:id/send", ac.Notification.SendNotification)
		}

		// ===== 公告管理: super_admin | content_admin =====
		adminAnnounce := protected.Group("")
		adminAnnounce.Use(middleware.RequireRole(middleware.RoleSuperAdmin, middleware.RoleContentAdmin))
		{
			adminAnnounce.POST("/announcements", ac.Announcement.CreateAnnouncement)
			adminAnnounce.PUT("/announcements/:id", ac.Announcement.UpdateAnnouncement)
			adminAnnounce.DELETE("/announcements/:id", ac.Announcement.DeleteAnnouncement)
			adminAnnounce.GET("/announcements/:id", ac.Announcement.GetAnnouncement)
			adminAnnounce.GET("/announcements", ac.Announcement.ListAnnouncements)
			adminAnnounce.PUT("/announcements/:id/status", ac.Announcement.UpdateAnnouncementStatus)
			adminAnnounce.POST("/announcements/:id/pin", ac.Announcement.PinAnnouncement)
		}

		// ===== 广告位管理: super_admin only =====
		adminAdPositions := protected.Group("")
		adminAdPositions.Use(middleware.RequireRole(middleware.RoleSuperAdmin))
		{
			adminAdPositions.POST("/ad-positions", ac.Ad.CreateAdPosition)
			adminAdPositions.PUT("/ad-positions/:id", ac.Ad.UpdateAdPosition)
			adminAdPositions.DELETE("/ad-positions/:id", ac.Ad.DeleteAdPosition)
		}
		// 广告位查询: 所有 admin 可看
		protected.GET("/ad-positions", ac.Ad.ListAdPositions)
		protected.GET("/ad-positions/:id", ac.Ad.GetAdPosition)

		// ===== 广告 CRUD + 审核: super_admin | content_admin =====
		adminAds := protected.Group("")
		adminAds.Use(middleware.RequireRole(middleware.RoleSuperAdmin, middleware.RoleContentAdmin))
		{
			adminAds.POST("/ads", ac.Ad.CreateAd)
			adminAds.PUT("/ads/:id", ac.Ad.UpdateAd)
			adminAds.DELETE("/ads/:id", ac.Ad.DeleteAd)
			adminAds.POST("/ads/review", ac.Ad.ReviewAd)
		}
		// 广告查询 + 统计 + view/click: 所有 admin 可访问
		protected.GET("/ads", ac.Ad.ListAds)
		protected.GET("/ads/:id", ac.Ad.GetAd)
		protected.POST("/ads/:id/view", ac.Ad.IncrementAdView)
		protected.POST("/ads/:id/click", ac.Ad.IncrementAdClick)
		protected.GET("/ads/stats", ac.Ad.GetAdStats)
	}
}
