// server/service/admin_review_service.go
package service

import (
	"errors"
	"strconv"
	"time"

	"pet-community/cache"
	"pet-community/model"
	"pet-community/repository"
	"pet-community/types"
)

// AdminReviewService 负责内容审核:待审核列表/通过/拒绝/批量/历史。
type AdminReviewService struct {
	adminRepo *repository.AdminRepository
	userRepo  *repository.UserRepository
	postRepo  *repository.PostRepository
}

// NewAdminReviewService 构造 AdminReviewService。
func NewAdminReviewService(adminRepo *repository.AdminRepository, userRepo *repository.UserRepository) *AdminReviewService {
	return &AdminReviewService{adminRepo: adminRepo, userRepo: userRepo}
}

// SetPostRepository 注入帖子仓储，用于动态置顶/精选（I7）。
func (s *AdminReviewService) SetPostRepository(postRepo *repository.PostRepository) {
	s.postRepo = postRepo
}

// PinPost 置顶/取消置顶动态（I7-01 管理员），并写审计日志。
//
// 仅对已发布动态生效；置顶会改变信息流排序（is_pinned 优先），需失效游客缓存与榜单缓存。
func (s *AdminReviewService) PinPost(postID uint, pinned int, adminID uint, ip string) error {
	if s.postRepo == nil {
		return errors.New("post repository not configured")
	}
	post, err := s.postRepo.GetByID(postID)
	if err != nil {
		return errors.New("动态不存在")
	}
	if post.Status != PostStatusPublished {
		return errors.New("仅可置顶已发布的动态")
	}
	if err := s.postRepo.SetPinned(postID, pinned); err != nil {
		return err
	}
	cache.InvalidateGuestFeed()
	cache.DeleteByPattern("rank:all:*")

	action := "unpin_post"
	detail := "取消置顶动态: post#" + strconv.FormatUint(uint64(postID), 10)
	if pinned == 1 {
		action = "pin_post"
		detail = "置顶动态: post#" + strconv.FormatUint(uint64(postID), 10)
	}
	log := &model.AdminOperationLog{
		AdminID:    adminID,
		Action:     action,
		TargetType: "post",
		TargetID:   postID,
		Detail:     detail,
		IP:         ip,
	}
	return s.adminRepo.CreateOperationLog(log)
}

// FeaturePost 设置/取消精选动态（I7-02 管理员），并写审计日志。
//
// 仅对已发布动态生效；精选影响信息流排序（is_featured 优先），需失效游客缓存与榜单缓存。
func (s *AdminReviewService) FeaturePost(postID uint, featured int, adminID uint, ip string) error {
	if s.postRepo == nil {
		return errors.New("post repository not configured")
	}
	post, err := s.postRepo.GetByID(postID)
	if err != nil {
		return errors.New("动态不存在")
	}
	if post.Status != PostStatusPublished {
		return errors.New("仅可精选已发布的动态")
	}
	if err := s.postRepo.SetFeatured(postID, featured); err != nil {
		return err
	}
	cache.InvalidateGuestFeed()
	cache.DeleteByPattern("rank:all:*")

	action := "unfeature_post"
	detail := "取消精选动态: post#" + strconv.FormatUint(uint64(postID), 10)
	if featured == 1 {
		action = "feature_post"
		detail = "精选动态: post#" + strconv.FormatUint(uint64(postID), 10)
	}
	log := &model.AdminOperationLog{
		AdminID:    adminID,
		Action:     action,
		TargetType: "post",
		TargetID:   postID,
		Detail:     detail,
		IP:         ip,
	}
	return s.adminRepo.CreateOperationLog(log)
}

func (s *AdminReviewService) GetPendingReviews(reviewType string, page, pageSize int) (*types.ReviewListResponse, error) {
	contents := make([]types.ReviewContent, 0)
	var total int64

	if reviewType == "post" || reviewType == "" {
		posts, postTotal, err := s.adminRepo.ListPendingPosts(page, pageSize)
		if err != nil {
			return nil, err
		}
		total = postTotal

		for _, post := range posts {
			user, _ := s.userRepo.GetByID(post.UserID)
			if user == nil {
				user = &model.User{}
			}
			contents = append(contents, types.ReviewContent{
				ID:         post.ID,
				Type:       "post",
				Content:    post.Content,
				Images:     post.Images,
				Video:      post.Video,
				UserID:     post.UserID,
				Username:   user.Nickname,
				Avatar:     user.Avatar,
				Status:     post.Status,
				IsPinned:   post.IsPinned,
				IsFeatured: post.IsFeatured,
				CreatedAt:  post.CreatedAt.Format(time.RFC3339),
			})
		}
	}

	if reviewType == "comment" || reviewType == "" {
		comments, commentTotal, err := s.adminRepo.ListPendingComments(page, pageSize)
		if err != nil {
			return nil, err
		}
		if reviewType == "" {
			total += commentTotal
		} else {
			total = commentTotal
		}

		for _, comment := range comments {
			user, _ := s.userRepo.GetByID(comment.UserID)
			if user == nil {
				user = &model.User{}
			}
			contents = append(contents, types.ReviewContent{
				ID:        comment.ID,
				Type:      "comment",
				Content:   comment.Content,
				PostID:    comment.PostID,
				UserID:    comment.UserID,
				Username:  user.Nickname,
				Avatar:    user.Avatar,
				Status:    comment.Status,
				CreatedAt: comment.CreatedAt.Format(time.RFC3339),
			})
		}
	}

	if reviewType == "topic" || reviewType == "" {
		topics, topicTotal, err := s.adminRepo.ListPendingTopics(page, pageSize)
		if err != nil {
			return nil, err
		}
		if reviewType == "" {
			total += topicTotal
		} else {
			total = topicTotal
		}

		for _, topic := range topics {
			user, _ := s.userRepo.GetByID(topic.CreatorID)
			if user == nil {
				user = &model.User{}
			}
			contents = append(contents, types.ReviewContent{
				ID:        topic.ID,
				Type:      "topic",
				Content:   topic.Name + " — " + topic.Description,
				UserID:    topic.CreatorID,
				Username:  user.Nickname,
				Avatar:    user.Avatar,
				Status:    topic.Status,
				CreatedAt: topic.CreatedAt.Format(time.RFC3339),
			})
		}
	}

	return &types.ReviewListResponse{
		List:  contents,
		Total: total,
		Page:  page,
		Size:  pageSize,
	}, nil
}

func (s *AdminReviewService) ApproveReview(targetType string, targetID uint, adminID uint, ip string) error {
	if targetType == "post" {
		if err := s.adminRepo.UpdatePostStatus(targetID, PostStatusPublished); err != nil {
			return err
		}
		cache.InvalidateGuestFeed()
		cache.DeleteByPattern("rank:all:*")
	} else if targetType == "comment" {
		if err := s.adminRepo.UpdateCommentStatus(targetID, 1); err != nil {
			return err
		}
	} else if targetType == "topic" {
		if err := s.adminRepo.UpdateTopicStatus(targetID, 1); err != nil {
			return err
		}
	} else {
		return errors.New("无效的审核类型")
	}

	log := &model.AdminOperationLog{
		AdminID:    adminID,
		Action:     "approve_review",
		TargetType: targetType,
		TargetID:   targetID,
		Detail:     "审核通过: " + targetType + "#" + strconv.FormatUint(uint64(targetID), 10),
		IP:         ip,
	}
	return s.adminRepo.CreateOperationLog(log)
}

func (s *AdminReviewService) RejectReview(targetType string, targetID uint, reason string, adminID uint, ip string) error {
	if targetType == "post" {
		// 3=已拒绝（2 留给定时待发）
		if err := s.adminRepo.UpdatePostStatus(targetID, PostStatusRejected); err != nil {
			return err
		}
	} else if targetType == "comment" {
		if err := s.adminRepo.UpdateCommentStatus(targetID, 2); err != nil {
			return err
		}
	} else if targetType == "topic" {
		if err := s.adminRepo.UpdateTopicStatus(targetID, 3); err != nil {
			return err
		}
	} else {
		return errors.New("无效的审核类型")
	}

	log := &model.AdminOperationLog{
		AdminID:    adminID,
		Action:     "reject_review",
		TargetType: targetType,
		TargetID:   targetID,
		Detail:     "审核拒绝: " + targetType + "#" + strconv.FormatUint(uint64(targetID), 10) + ", 原因: " + reason,
		IP:         ip,
	}
	return s.adminRepo.CreateOperationLog(log)
}

func (s *AdminReviewService) BatchReview(targetType string, ids []uint, status int, reason string, adminID uint, ip string) error {
	for _, id := range ids {
		if status == 1 {
			if err := s.ApproveReview(targetType, id, adminID, ip); err != nil {
				return err
			}
		} else if status == 2 {
			if err := s.RejectReview(targetType, id, reason, adminID, ip); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *AdminReviewService) GetReviewHistory(targetType string, page, pageSize int) (*types.ReviewHistoryResponse, error) {
	histories := make([]types.ReviewHistory, 0)

	logs, total, err := s.adminRepo.GetReviewHistory(targetType, page, pageSize)
	if err != nil {
		return nil, err
	}

	for _, log := range logs {
		admin, _ := s.adminRepo.FindByID(log.AdminID)
		if admin == nil {
			admin = &model.AdminUser{}
		}
		histories = append(histories, types.ReviewHistory{
			ID:         log.ID,
			TargetType: log.TargetType,
			TargetID:   log.TargetID,
			Action:     log.Action,
			Reason:     log.Detail,
			AdminID:    log.AdminID,
			AdminName:  admin.Username,
			CreatedAt:  log.CreatedAt.Format(time.RFC3339),
		})
	}

	return &types.ReviewHistoryResponse{
		List:  histories,
		Total: total,
		Page:  page,
		Size:  pageSize,
	}, nil
}
