package service

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"pet-community/model"
	"pet-community/repository"
	"pet-community/types"
)

// AdminReportService 管理后台举报处理：列表/处理/批量/统计。
type AdminReportService struct {
	reportRepo *repository.ReportRepository
	adminRepo  *repository.AdminRepository
	userRepo   *repository.UserRepository
}

func NewAdminReportService(
	reportRepo *repository.ReportRepository,
	adminRepo *repository.AdminRepository,
	userRepo *repository.UserRepository,
) *AdminReportService {
	return &AdminReportService{
		reportRepo: reportRepo,
		adminRepo:  adminRepo,
		userRepo:   userRepo,
	}
}

func (s *AdminReportService) ListReports(status *int, reasonType, targetType string, page, pageSize int) (*types.ReportListResponse, error) {
	reports, total, err := s.reportRepo.ListAdmin(status, reasonType, targetType, page, pageSize)
	if err != nil {
		return nil, err
	}

	list := make([]types.ReportInfo, 0, len(reports))
	for i := range reports {
		info := s.toAdminReportInfo(&reports[i])
		list = append(list, *info)
	}

	return &types.ReportListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *AdminReportService) HandleReport(reportID uint, action, result string, adminID uint, ip string) error {
	report, err := s.reportRepo.GetByID(reportID)
	if err != nil || report == nil {
		return errors.New("举报不存在")
	}
	if report.Status != 0 {
		return errors.New("该举报已处理")
	}

	status, handleResult, err := mapReportAction(action, result)
	if err != nil {
		return err
	}

	now := time.Now()
	if err := s.reportRepo.UpdateHandle(reportID, status, handleResult, adminID, now); err != nil {
		return err
	}

	if err := s.applyActionSideEffects(report, action); err != nil {
		return err
	}

	s.notifyReporter(report, action, handleResult)
	if action == "warn" {
		s.notifyTargetAuthor(report, handleResult)
	}

	log := &model.AdminOperationLog{
		AdminID:    adminID,
		Action:     "handle_report",
		TargetType: "report",
		TargetID:   reportID,
		Detail:     fmt.Sprintf("处理举报#%d action=%s result=%s", reportID, action, handleResult),
		IP:         ip,
	}
	_ = s.adminRepo.CreateOperationLog(log)
	return nil
}

func (s *AdminReportService) BatchHandleReports(ids []uint, action, result string, adminID uint, ip string) error {
	if len(ids) == 0 {
		return errors.New("请选择要处理的举报")
	}
	var lastErr error
	success := 0
	for _, id := range ids {
		if err := s.HandleReport(id, action, result, adminID, ip); err != nil {
			lastErr = err
			continue
		}
		success++
	}
	if success == 0 && lastErr != nil {
		return lastErr
	}
	return nil
}

func (s *AdminReportService) GetStats() (*types.ReportStatsResponse, error) {
	total, err := s.reportRepo.CountAll()
	if err != nil {
		return nil, err
	}
	pending, _ := s.reportRepo.CountByStatus(0)
	handled, _ := s.reportRepo.CountByStatus(1)
	ignored, _ := s.reportRepo.CountByStatus(2)

	var handleRate float64
	if total > 0 {
		handleRate = float64(handled+ignored) / float64(total) * 100
	}

	reasonRows, _ := s.reportRepo.CountGroupByReasonType()
	byReason := make([]types.ReportTypeCount, 0, len(reasonRows))
	for _, row := range reasonRows {
		t := row.ReasonType
		if t == "" {
			t = "other"
		}
		byReason = append(byReason, types.ReportTypeCount{Type: t, Count: row.Count})
	}

	targetRows, _ := s.reportRepo.CountGroupByTargetType()
	byTarget := make([]types.ReportTypeCount, 0, len(targetRows))
	for _, row := range targetRows {
		t := row.TargetType
		if t == "" {
			t = "unknown"
		}
		byTarget = append(byTarget, types.ReportTypeCount{Type: t, Count: row.Count})
	}

	return &types.ReportStatsResponse{
		Total:        total,
		Pending:      pending,
		Handled:      handled,
		Ignored:      ignored,
		HandleRate:   handleRate,
		ByReasonType: byReason,
		ByTargetType: byTarget,
	}, nil
}

func mapReportAction(action, result string) (status int, handleResult string, err error) {
	switch action {
	case "approve":
		status = 1
		handleResult = "通过"
	case "warn":
		status = 1
		handleResult = "警告"
	case "reject":
		status = 2
		handleResult = "忽略"
	default:
		return 0, "", errors.New("无效的处理动作")
	}
	if result != "" {
		handleResult = handleResult + ": " + result
		if len(handleResult) > 200 {
			handleResult = handleResult[:200]
		}
	}
	return status, handleResult, nil
}

func (s *AdminReportService) applyActionSideEffects(report *model.Report, action string) error {
	if action != "approve" {
		return nil
	}
	switch report.TargetType {
	case "post":
		// PostStatusRejected=3（2 为定时待发，不可用作下架）
		return s.adminRepo.UpdatePostStatus(report.TargetID, PostStatusRejected)
	case "comment":
		// 评论：1=通过，2=拒绝（与审核服务一致）
		return s.adminRepo.UpdateCommentStatus(report.TargetID, 2)
	}
	return nil
}

func (s *AdminReportService) notifyReporter(report *model.Report, action, handleResult string) {
	title := "举报处理结果"
	content := fmt.Sprintf("您的举报（#%d）已处理：%s", report.ID, handleResult)
	if action == "reject" {
		content = fmt.Sprintf("您的举报（#%d）经核查后已忽略：%s", report.ID, handleResult)
	}
	_ = s.adminRepo.CreateUserNotification(&model.UserNotification{
		UserID:      report.ReporterID,
		Type:        "system",
		Title:       title,
		Content:     content,
		TriggerType: "report",
		TriggerID:   report.ID,
	})
}

func (s *AdminReportService) notifyTargetAuthor(report *model.Report, handleResult string) {
	var authorID uint
	switch report.TargetType {
	case "post":
		post, err := s.adminRepo.GetPostByID(report.TargetID)
		if err != nil || post == nil {
			return
		}
		authorID = post.UserID
	case "comment":
		comment, err := s.adminRepo.GetCommentByID(report.TargetID)
		if err != nil || comment == nil {
			return
		}
		authorID = comment.UserID
	default:
		return
	}
	if authorID == 0 {
		return
	}
	_ = s.adminRepo.CreateUserNotification(&model.UserNotification{
		UserID:      authorID,
		Type:        "system",
		Title:       "内容违规警告",
		Content:     fmt.Sprintf("您的%s内容因被举报受到警告：%s", report.TargetType, handleResult),
		TriggerType: report.TargetType,
		TriggerID:   report.TargetID,
	})
}

func (s *AdminReportService) toAdminReportInfo(report *model.Report) *types.ReportInfo {
	handledAt := ""
	if report.HandledAt != nil {
		handledAt = report.HandledAt.Format(time.RFC3339)
	}
	reporterName := ""
	if user, err := s.userRepo.GetByID(report.ReporterID); err == nil && user != nil {
		reporterName = user.Nickname
		if reporterName == "" {
			reporterName = "用户" + strconv.FormatUint(uint64(user.ID), 10)
		}
	}
	return &types.ReportInfo{
		ID:           report.ID,
		ReporterID:   report.ReporterID,
		ReporterName: reporterName,
		TargetType:   report.TargetType,
		TargetID:     report.TargetID,
		Reason:       report.Reason,
		ReasonType:   report.ReasonType,
		Status:       report.Status,
		HandleResult: report.HandleResult,
		HandledBy:    report.HandledBy,
		HandledAt:    handledAt,
		CreatedAt:    report.CreatedAt.Format(time.RFC3339),
	}
}
