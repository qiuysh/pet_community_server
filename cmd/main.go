package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"

	"pet-community/config"
	"pet-community/controller"
	"pet-community/middleware"
	"pet-community/model"
	"pet-community/repository"
	"pet-community/router"
	"pet-community/service"
	"pet-community/types"
	"pet-community/utils"
)

func main() {
	config.Load()
	config.InitDB()
	config.InitRedis()
	utils.InitRSA(2048)

	// 初始化统一存储：本地磁盘 / 对象存储（OSS），由 storage.driver 决定
	store, err := utils.InitStorage(&config.Cfg.Storage)
	if err != nil {
		log.Fatalf("init storage failed: %v", err)
	}
	log.Printf("storage driver=%s", store.Driver())

	if err := config.DB.AutoMigrate(
		&model.AdminUser{},
		&model.AdminOperationLog{},
		&model.AdminLoginLog{},
		&model.SystemConfig{},
		&model.SensitiveWord{},
		&model.UserActivityLog{},
		&model.UserBehaviorLog{},
		&model.User{},
		&model.UserDevice{},
		&model.VerificationCode{},
		&model.UserThirdAccount{},
		&model.SystemNotification{},
		&model.Announcement{},
		&model.Pet{},
		&model.PetBreed{},
		&model.PetAlbum{},
		&model.PetGrowthRecord{},
		&model.LostPet{},
		&model.Post{},
		&model.Comment{},
		&model.Like{},
		&model.Favorite{},
		&model.Draft{},
		&model.Report{},
		&model.Topic{},
		&model.ShareRecord{},
		&model.UserNotification{},
		&model.Follow{},
		&model.Block{},
		&model.Conversation{},
		&model.Message{},
		&model.UserLevel{},
		&model.PointsRecord{},
		&model.PointsRule{},
		&model.LevelBenefit{},
		&model.NotificationSettings{},
		&model.SearchHistory{},
		&model.UserInterest{},
		&model.TopicCategory{},
		&model.TopicParticipant{},
		&model.HotKeyword{},
		&model.Activity{},
		&model.Registration{},
		&model.Submission{},
		&model.Vote{},
		&model.SignIn{},
		&model.ActivityAlbum{},
		&model.ActivityReview{},
		&model.Adopt{},
		&model.AdoptionApplication{},
		&model.AdoptionFollowUp{},
		&model.Statistics{},
		&model.AdPosition{},
		&model.Advertisement{},
	); err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}

	ensureDefaultAdmin()

	adminRepo := repository.NewAdminRepository(config.DB)
	userRepo := repository.NewUserRepository(config.DB)

	adminAuthSvc := service.NewAdminAuthService(adminRepo)
	adminConfigSvc := service.NewAdminConfigService(adminRepo)
	adminLogSvc := service.NewAdminLogService(adminRepo)
	adminFrontUserSvc := service.NewAdminFrontUserService(adminRepo, userRepo)
	adminReviewSvc := service.NewAdminReviewService(adminRepo, userRepo)
	adminNotifySvc := service.NewAdminNotificationService(adminRepo)
	adminAnnounceSvc := service.NewAdminAnnouncementService(adminRepo)

	statsRepo := repository.NewStatsRepository(config.DB)
	postRepo := repository.NewPostRepository(config.DB)
	likeRepo := repository.NewLikeRepository(config.DB)
	favoriteRepo := repository.NewFavoriteRepository(config.DB)
	topicRepo := repository.NewTopicRepository(config.DB)
	commentRepo := repository.NewCommentRepository(config.DB)
	shareRepo := repository.NewShareRepository(config.DB)

	statsService := service.NewStatsService(statsRepo, userRepo, postRepo, commentRepo, likeRepo, favoriteRepo, shareRepo, topicRepo)

	petRepo := repository.NewPetRepository(config.DB)
	reportRepo := repository.NewReportRepository(config.DB)
	adminReportSvc := service.NewAdminReportService(reportRepo, adminRepo, userRepo)
	dashboardService := service.NewDashboardService(adminRepo, userRepo, postRepo, petRepo, statsRepo, reportRepo)

	adRepo := repository.NewAdRepository(config.DB)
	adService := service.NewAdService(adRepo, adminRepo)

	behaviorLogRepo := repository.NewBehaviorLogRepository(config.DB)
	logService := service.NewLogService(behaviorLogRepo)

	adminAuthCtrl := controller.NewAdminAuthController(adminAuthSvc)
	adminUsersCtrl := controller.NewAdminUsersController(adminAuthSvc)
	adminFrontUsersCtrl := controller.NewAdminFrontUsersController(adminFrontUserSvc, adminLogSvc)
	adminSettingsCtrl := controller.NewAdminSettingsController(adminConfigSvc)
	adminLogsCtrl := controller.NewAdminLogsController(adminLogSvc, adminAuthSvc, logService)
	adminReviewCtrl := controller.NewAdminReviewController(adminReviewSvc)
	adminReportCtrl := controller.NewAdminReportController(adminReportSvc)
	adminStatsCtrl := controller.NewAdminStatsController(statsService)
	adminDashboardCtrl := controller.NewAdminDashboardController(dashboardService)
	adminNotifyCtrl := controller.NewAdminNotificationController(adminNotifySvc)
	adminAnnounceCtrl := controller.NewAdminAnnouncementController(adminAnnounceSvc)
	adminAdCtrl := controller.NewAdminAdController(adService)
	adminCtrls := router.AdminControllers{
		Auth:         adminAuthCtrl,
		Users:        adminUsersCtrl,
		FrontUsers:   adminFrontUsersCtrl,
		Settings:     adminSettingsCtrl,
		Logs:         adminLogsCtrl,
		Review:       adminReviewCtrl,
		Report:       adminReportCtrl,
		Stats:        adminStatsCtrl,
		Dashboard:    adminDashboardCtrl,
		Notification: adminNotifyCtrl,
		Announcement: adminAnnounceCtrl,
		Ad:           adminAdCtrl,
	}
	userService := service.NewUserService(userRepo)
	userController := controller.NewUserController(userService)
	petService := service.NewPetService(petRepo)
	petController := controller.NewPetController(petService)

	petRepo.InitBreeds()

	draftRepo := repository.NewDraftRepository(config.DB)
	notificationRepo := repository.NewNotificationRepository(config.DB)
	followRepo := repository.NewFollowRepository(config.DB)
	userService.SetProfileRepos(followRepo, postRepo, petRepo)

	notificationService := service.NewNotificationService(notificationRepo, followRepo, userRepo)
	followService := service.NewFollowService(followRepo, userRepo, notificationRepo)

	postService := service.NewPostService(postRepo, likeRepo, favoriteRepo, topicRepo, userRepo, petRepo)
	draftService := service.NewDraftService(draftRepo)
	topicService := service.NewTopicService(topicRepo)
	commentService := service.NewCommentService(commentRepo, likeRepo, postRepo, userRepo)
	reportService := service.NewReportService(reportRepo)

	postService.SetNotificationService(notificationService)
	commentService.SetNotificationService(notificationService)

	postController := controller.NewPostController(postService, draftService, topicService, commentService, reportService)

	notificationController := controller.NewNotificationController(notificationService, notificationRepo)
	followController := controller.NewFollowController(followService)

	blockRepo := repository.NewBlockRepository(config.DB)
	conversationRepo := repository.NewConversationRepository(config.DB)
	messageRepo := repository.NewMessageRepository(config.DB)
	pointsRepo := repository.NewPointsRepository(config.DB)
	pointsRuleRepo := repository.NewPointsRuleRepository(config.DB)
	userLevelRepo := repository.NewUserLevelRepository(config.DB)
	levelBenefitRepo := repository.NewLevelBenefitRepository(config.DB)
	notificationSettingsRepo := repository.NewNotificationSettingsRepository(config.DB)

	searchHistoryRepo := repository.NewSearchHistoryRepository(config.DB)
	userInterestRepo := repository.NewUserInterestRepository(config.DB)
	topicCategoryRepo := repository.NewTopicCategoryRepository(config.DB)
	topicParticipantRepo := repository.NewTopicParticipantRepository(config.DB)
	hotKeywordRepo := repository.NewHotKeywordRepository(config.DB)

	blockService := service.NewBlockService(blockRepo, userRepo)
	blockService.SetFollowRepository(followRepo)
	messageService := service.NewMessageService(conversationRepo, messageRepo, userRepo, blockRepo)
	pointsService := service.NewPointsService(pointsRepo, userLevelRepo, pointsRuleRepo, levelBenefitRepo)
	notificationSettingsService := service.NewNotificationSettingsService(notificationSettingsRepo)

	discoveryService := service.NewDiscoveryService(
		postRepo, userRepo, topicRepo, likeRepo, favoriteRepo, petRepo,
		searchHistoryRepo, userInterestRepo, topicCategoryRepo, topicParticipantRepo,
		hotKeywordRepo, followRepo,
	)
	discoveryService.SetBlockRepository(blockRepo)

	postService.SetPointsService(pointsService)
	postService.SetBlockRepository(blockRepo)
	postService.SetFollowRepository(followRepo)
	postService.SetDraftRepository(draftRepo)
	userService.SetBlockRepository(blockRepo)
	commentService.SetPointsService(pointsService)
	commentService.SetBlockRepository(blockRepo)
	commentService.SetFollowRepository(followRepo)
	messageService.SetFollowRepository(followRepo)
	followService.SetBlockRepository(blockRepo)
	followService.SetPointsService(pointsService)

	blockController := controller.NewBlockController(blockService)
	messageController := controller.NewMessageController(messageService, userService)
	pointsController := controller.NewPointsController(pointsService)
	notificationSettingsController := controller.NewNotificationSettingsController(notificationSettingsService)
	discoveryController := controller.NewDiscoveryController(discoveryService)

	wsService := service.NewWebSocketService()
	messageService.SetWebSocketService(wsService)

	topicRepo.InitDefaultTopics()
	topicCategoryRepo.InitDefaultCategories()
	_ = pointsRuleRepo.InitDefaultRules()
	_ = levelBenefitRepo.InitDefaultBenefits()

	activityRepo := repository.NewActivityRepository(config.DB)
	registrationRepo := repository.NewRegistrationRepository(config.DB)
	submissionRepo := repository.NewSubmissionRepository(config.DB)
	voteRepo := repository.NewVoteRepository(config.DB)
	signInRepo := repository.NewSignInRepository(config.DB)
	albumRepo := repository.NewActivityAlbumRepository(config.DB)
	reviewRepo := repository.NewActivityReviewRepository(config.DB)

	activityService := service.NewActivityService(activityRepo, registrationRepo, submissionRepo, voteRepo, userRepo)
	registrationService := service.NewRegistrationService(registrationRepo, activityRepo, userRepo)
	submissionService := service.NewSubmissionService(submissionRepo, activityRepo, registrationRepo, userRepo, voteRepo)
	voteService := service.NewVoteService(voteRepo, submissionRepo, activityRepo, userRepo)
	signInService := service.NewSignInService(signInRepo, activityRepo, registrationRepo, userRepo)
	albumService := service.NewAlbumService(albumRepo, activityRepo, registrationRepo, userRepo)
	reviewService := service.NewActivityReviewService(reviewRepo, activityRepo, userRepo)

	activityController := controller.NewActivityController(activityService, registrationService, submissionService, voteService, signInService, albumService, reviewService)

	adoptRepo := repository.NewAdoptRepository(config.DB)
	applicationRepo := repository.NewAdoptionApplicationRepository(config.DB)
	followUpRepo := repository.NewAdoptionFollowUpRepository(config.DB)
	adoptService := service.NewAdoptService(adoptRepo, applicationRepo, followUpRepo, userRepo)
	adoptController := controller.NewAdoptController(adoptService)

	frontAdController := controller.NewFrontAdController(adService)

	r := gin.Default()
	mountUploadRoutes(r)

	middleware.SetUploadPrefix(config.Cfg.Storage.Local.BaseURL)
	middleware.SetLogService(logService)
	r.Use(middleware.UserBehaviorLogMiddleware())
	r.Use(middleware.CSRFMiddleware())
	r.Use(middleware.RateLimitMiddleware())

	r.GET("/health", adminCtrls.Auth.HealthCheck)
	r.GET("/api/v1/csrf-token", middleware.IssueCSRFToken)
	r.GET("/api/v1/public-key", func(c *gin.Context) {
		c.JSON(200, types.SuccessResponse(gin.H{
			"publicKey": utils.GetPublicKey(),
		}))
	})
	r.GET("/ws", middleware.WebSocketAuthMiddleware(), func(c *gin.Context) {
		wsService.ServeWS(c.Writer, c.Request)
	})

	startSchedulers(statsService, adminNotifySvc, postService, adService)

	apiV1 := r.Group("/api/v1")
	router.RegisterAdminRoutes(r, adminCtrls)
	router.SetupUserRoutes(apiV1, userController)
	router.SetupPetRoutes(apiV1, petController)
	router.SetupPostRoutes(apiV1, postController)
	router.SetupNotificationRoutes(apiV1, notificationController, followController)
	router.SetupSocialRoutes(apiV1, blockController, messageController, pointsController, notificationSettingsController, followController)
	router.SetupDiscoveryRoutes(apiV1, discoveryController)
	router.SetupActivityRoutes(apiV1, activityController)
	router.SetupAdoptRoutes(apiV1, adoptController)
	router.SetupAdRoutes(apiV1, frontAdController)

	addr := config.Cfg.Addr()
	log.Printf("server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

// mountUploadRoutes 挂载上传文件的静态访问路由。
// 仅本地驱动需要挂载（对象存储由 bucket 域名直接提供访问）；
// 挂载路径取自 storage.local.base_url（默认 /api/v1/uploads），需与生成的 URL 前缀一致。
func mountUploadRoutes(r *gin.Engine) {
	if config.Cfg.Storage.Driver != config.StorageDriverLocal {
		return
	}
	r.Static(config.Cfg.Storage.Local.BaseURL, config.Cfg.Storage.Local.Root)
}

func ensureDefaultAdmin() {
	var count int64
	config.DB.Model(&model.AdminUser{}).Count(&count)
	if count > 0 {
		return
	}
	hashedPassword, err := utils.HashPassword("admin123")
	if err != nil {
		log.Fatalf("hash default admin password: %v", err)
	}
	if err := config.DB.Create(&model.AdminUser{
		Username: "admin",
		Password: hashedPassword,
		Email:    "admin@petcommunity.com",
		Role:     "super_admin",
		Status:   1,
	}).Error; err != nil {
		log.Fatalf("create default admin: %v", err)
	}
}

func startSchedulers(
	statsService *service.StatsService,
	adminNotifySvc *service.AdminNotificationService,
	postService *service.PostService,
	adService *service.AdService,
) {
	sched := config.Cfg.Scheduler
	c := cron.New(cron.WithSeconds())

	mustAddCron(c, "scheduler.stats_cron", sched.StatsCron, func() {
		statsService.CalculateStats()
	})
	mustAddCron(c, "scheduler.notification_cron", sched.NotificationCron, func() {
		adminNotifySvc.ProcessScheduledNotifications()
	})
	mustAddCron(c, "scheduler.scheduled_post_cron", sched.ScheduledPostCron, func() {
		postService.PublishScheduledPosts()
	})
	// 每 30 秒刷广告曝光/点击缓冲
	mustAddCron(c, "scheduler.ad_counter_flush", "0 */30 * * * *", func() {
		adService.FlushBufferedCounters()
	})
	c.Start()

	go func() {
		if delay := time.Duration(sched.StatsStartupDelaySec) * time.Second; delay > 0 {
			time.Sleep(delay)
		}
		statsService.CalculateStats()
	}()
}

func mustAddCron(c *cron.Cron, name, expr string, fn func()) {
	if _, err := c.AddFunc(expr, fn); err != nil {
		log.Fatalf("invalid %s %q: %v", name, expr, err)
	}
}
