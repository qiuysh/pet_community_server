package main

import (
	"crypto/rand"
	"encoding/hex"
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

	// 表结构唯一来源见 config.Models()；新增模型记得同步 sql/init/create.sql
	if err := config.DB.AutoMigrate(config.Models()...); err != nil {
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
	// I7 置顶/精选走内容审核域，注入帖子仓储
	adminReviewSvc.SetPostRepository(postRepo)
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
	topicCategoryRepo := repository.NewTopicCategoryRepository(config.DB)
	topicService := service.NewTopicService(topicRepo, topicCategoryRepo)
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
	topicParticipantRepo := repository.NewTopicParticipantRepository(config.DB)
	hotKeywordRepo := repository.NewHotKeywordRepository(config.DB)
	adminReviewSvc.SetTopicCategoryRepository(topicCategoryRepo)

	blockService := service.NewBlockService(blockRepo, userRepo)
	blockService.SetFollowRepository(followRepo)
	messageService := service.NewMessageService(conversationRepo, messageRepo, userRepo, blockRepo)
	pointsService := service.NewPointsService(pointsRepo, userLevelRepo, pointsRuleRepo, levelBenefitRepo)
	notificationSettingsService := service.NewNotificationSettingsService(notificationSettingsRepo)
	// N8：通知创建侧统一读设置开关过滤
	notificationService.SetSettingsRepository(notificationSettingsRepo)
	followService.SetNotificationSettingsRepository(notificationSettingsRepo)
	adminNotifySvc.SetNotificationSettingsRepository(notificationSettingsRepo)
	// R4：登录登记未知设备时写安全通知
	userService.SetNotificationService(notificationService)

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
	messageService.SetNotificationSettingsRepository(notificationSettingsRepo)
	followService.SetBlockRepository(blockRepo)
	followService.SetPointsService(pointsService)

	blockController := controller.NewBlockController(blockService)
	messageController := controller.NewMessageController(messageService, userService)
	pointsController := controller.NewPointsController(pointsService)
	notificationSettingsController := controller.NewNotificationSettingsController(notificationSettingsService)
	discoveryController := controller.NewDiscoveryController(discoveryService)

	wsService := service.NewWebSocketService()
	messageService.SetWebSocketService(wsService)
	// 在线状态以 WebSocket 连接态为准（见 FollowService.GetOnlineStatus）
	followService.SetPresence(wsService)

	// 种子数据失败不阻断启动（缺数据不等于服务不可用），但必须留痕：
	// 静默失败会让「话题广场空着 / 积分规则缺失」变成无从排查的现象。
	logSeedErr("初始化默认话题", topicRepo.InitDefaultTopics())
	logSeedErr("初始化话题分类", topicCategoryRepo.InitDefaultCategories())
	logSeedErr("初始化积分规则", pointsRuleRepo.InitDefaultRules())
	logSeedErr("初始化等级权益", levelBenefitRepo.InitDefaultBenefits())

	activityRepo := repository.NewActivityRepository(config.DB)
	registrationRepo := repository.NewRegistrationRepository(config.DB)
	submissionRepo := repository.NewSubmissionRepository(config.DB)
	voteRepo := repository.NewVoteRepository(config.DB)
	signInRepo := repository.NewSignInRepository(config.DB)
	albumRepo := repository.NewActivityAlbumRepository(config.DB)
	reviewRepo := repository.NewActivityReviewRepository(config.DB)

	activityService := service.NewActivityService(activityRepo, registrationRepo, submissionRepo, voteRepo, userRepo)
	registrationService := service.NewRegistrationService(registrationRepo, activityRepo, userRepo)
	submissionService := service.NewSubmissionService(service.SubmissionServiceDeps{
		SubmissionRepo:   submissionRepo,
		ActivityRepo:     activityRepo,
		RegistrationRepo: registrationRepo,
		UserRepo:         userRepo,
		VoteRepo:         voteRepo,
		NotificationRepo: notificationRepo,
	})
	voteService := service.NewVoteService(voteRepo, submissionRepo, activityRepo, userRepo)
	signInService := service.NewSignInService(signInRepo, activityRepo, registrationRepo, userRepo)
	albumService := service.NewAlbumService(albumRepo, activityRepo, registrationRepo, userRepo)
	reviewService := service.NewActivityReviewService(reviewRepo, activityRepo, userRepo)

	activityController := controller.NewActivityController(activityService, registrationService, submissionService, voteService, signInService, albumService, reviewService)

	adoptRepo := repository.NewAdoptRepository(config.DB)
	applicationRepo := repository.NewAdoptionApplicationRepository(config.DB)
	followUpRepo := repository.NewAdoptionFollowUpRepository(config.DB)
	adoptService := service.NewAdoptService(adoptRepo, applicationRepo, followUpRepo, userRepo)
	// R6：回访提交后通知对方
	adoptService.SetNotificationRepository(notificationRepo)
	adoptController := controller.NewAdoptController(adoptService)

	frontAdController := controller.NewFrontAdController(adService)
	siteController := controller.NewSiteController(adminConfigSvc)

	uploadService := service.NewUploadService()
	uploadController := controller.NewUploadController(uploadService)

	r := gin.Default()
	mountUploadRoutes(r)

	middleware.SetUploadPrefix(config.Cfg.Storage.Local.BaseURL)
	middleware.SetLogService(logService)
	r.Use(middleware.UserBehaviorLogMiddleware())
	r.Use(middleware.CSRFMiddleware())
	r.Use(middleware.RateLimitMiddleware())

	// 路由表统一由 router.RegisterRoutes 注册，契约测试复用同一入口
	router.RegisterRoutes(r, router.Handlers{
		Admin:                adminCtrls,
		User:                 userController,
		Pet:                  petController,
		Post:                 postController,
		Notification:         notificationController,
		Follow:               followController,
		Block:                blockController,
		Message:              messageController,
		Points:               pointsController,
		NotificationSettings: notificationSettingsController,
		Discovery:            discoveryController,
		Activity:             activityController,
		Adopt:                adoptController,
		FrontAd:              frontAdController,
		Site:                 siteController,
		Upload:               uploadController,
	})

	// WebSocket 入口带运行时状态，单独挂载
	r.GET("/ws", middleware.WebSocketAuthMiddleware(), func(c *gin.Context) {
		wsService.ServeWS(c.Writer, c.Request)
	})

	startSchedulers(statsService, adminNotifySvc, postService, adService)

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

	username := config.Cfg.Admin.InitialUsername
	if username == "" {
		username = "admin"
	}
	// 口令不再硬编码：配置留空时随机生成，并在日志里打印一次供首次登录后修改
	password := config.Cfg.Admin.InitialPassword
	generated := false
	if password == "" {
		buf := make([]byte, 9)
		if _, err := rand.Read(buf); err != nil {
			log.Fatalf("generate admin password: %v", err)
		}
		password = hex.EncodeToString(buf)
		generated = true
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		log.Fatalf("hash default admin password: %v", err)
	}
	if err := config.DB.Create(&model.AdminUser{
		Username: username,
		Password: hashedPassword,
		Email:    "admin@petcommunity.com",
		Role:     "super_admin",
		Status:   1,
	}).Error; err != nil {
		log.Fatalf("create default admin: %v", err)
	}
	if generated {
		log.Printf("已创建初始超级管理员 %s，随机口令：%s（请登录后立即修改）", username, password)
	} else {
		log.Printf("已按 config.yaml 的 admin 配置创建初始超级管理员 %s", username)
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

// logSeedErr 记录种子数据初始化失败，不阻断启动。
func logSeedErr(action string, err error) {
	if err != nil {
		log.Printf("[warn] %s失败（服务继续启动，相关功能可能不可用）: %v", action, err)
	}
}
