package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"pet-community/cache"
	"pet-community/config"
	"pet-community/model"
	"pet-community/repository"
	"pet-community/types"

	"gorm.io/gorm"
)

type PostService struct {
	postRepo            *repository.PostRepository
	likeRepo            *repository.LikeRepository
	favoriteRepo        *repository.FavoriteRepository
	topicRepo           *repository.TopicRepository
	userRepo            *repository.UserRepository
	petRepo             *repository.PetRepository
	blockRepo           *repository.BlockRepository
	followRepo          *repository.FollowRepository
	draftRepo           *repository.DraftRepository
	notificationService *NotificationService
	pointsService       *PointsService
}

func NewPostService(
	postRepo *repository.PostRepository,
	likeRepo *repository.LikeRepository,
	favoriteRepo *repository.FavoriteRepository,
	topicRepo *repository.TopicRepository,
	userRepo *repository.UserRepository,
	petRepo *repository.PetRepository,
) *PostService {
	return &PostService{
		postRepo:     postRepo,
		likeRepo:     likeRepo,
		favoriteRepo: favoriteRepo,
		topicRepo:    topicRepo,
		userRepo:     userRepo,
		petRepo:      petRepo,
	}
}

func (s *PostService) SetNotificationService(ns *NotificationService) {
	s.notificationService = ns
}

func (s *PostService) SetPointsService(ps *PointsService) {
	s.pointsService = ps
}

func (s *PostService) SetBlockRepository(blockRepo *repository.BlockRepository) {
	s.blockRepo = blockRepo
}

func (s *PostService) SetFollowRepository(followRepo *repository.FollowRepository) {
	s.followRepo = followRepo
}

func (s *PostService) SetDraftRepository(draftRepo *repository.DraftRepository) {
	s.draftRepo = draftRepo
}

// parsePublishAt 解析并校验定时发布时间。
// requireFuture 为 true 时要求时间晚于当前时间；始终校验最多延迟 7 天。
// 空字符串返回 nil,表示不设置定时发布（立即发布）。
func parsePublishAt(raw string, requireFuture bool) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	pt, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, errors.New("定时发布时间格式错误")
	}
	now := time.Now()
	if requireFuture && pt.Before(now) {
		return nil, errors.New("定时发布时间不能早于当前时间")
	}
	if pt.After(now.Add(7 * 24 * time.Hour)) {
		return nil, errors.New("定时发布时间最多延迟7天")
	}
	return &pt, nil
}

func (s *PostService) CreatePost(userID uint, req *types.PostCreateRequest) (*types.PostInfo, error) {
	if req.Content == "" && len(req.Images) == 0 && req.Video == "" {
		return nil, errors.New("请添加内容")
	}
	if len(req.Images) > 0 && req.Video != "" {
		return nil, errors.New("图文和视频不能同时发布")
	}
	if len(req.TopicIDs) > 3 {
		return nil, errors.New("最多添加3个话题")
	}
	req.Content = FilterSensitiveText(req.Content)

	imagesStr := ""
	if len(req.Images) > 0 {
		imagesStr = strings.Join(req.Images, ",")
	}

	petIDsStr := ""
	if len(req.PetIDs) > 0 {
		var petIDStrs []string
		for _, id := range req.PetIDs {
			petIDStrs = append(petIDStrs, fmt.Sprintf("%d", id))
		}
		petIDsStr = strings.Join(petIDStrs, ",")
	}

	topicIDsStr := ""
	if len(req.TopicIDs) > 0 {
		var topicIDStrs []string
		for _, id := range req.TopicIDs {
			topicIDStrs = append(topicIDStrs, fmt.Sprintf("%d", id))
		}
		topicIDsStr = strings.Join(topicIDStrs, ",")
	}

	// 状态：0待审 1已发布 2定时待发 3已拒绝
	status := PostStatusPending
	var publishAt *time.Time
	if req.PublishAt != "" {
		pt, err := parsePublishAt(req.PublishAt, true)
		if err != nil {
			return nil, err
		}
		publishAt = pt
		status = PostStatusScheduled
	} else if isAutoAuditEnabled() {
		status = PostStatusPublished
	}

	var author *model.User
	if s.userRepo != nil {
		author, _ = s.userRepo.GetByID(userID)
	}
	visibility := ResolvePostVisibility(req.Visibility, author)

	post := &model.Post{
		UserID:     userID,
		Content:    req.Content,
		Images:     imagesStr,
		Video:      req.Video,
		PetIDs:     petIDsStr,
		TopicIDs:   topicIDsStr,
		Location:   req.Location,
		Visibility: visibility,
		PublishAt:  publishAt,
		Status:     status,
	}

	if err := s.postRepo.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(post).Error; err != nil {
			return err
		}

		for _, topicID := range req.TopicIDs {
			if err := tx.Model(&model.Topic{}).Where("id = ?", topicID).UpdateColumn("posts_count", gorm.Expr("posts_count + 1")).Error; err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	if s.notificationService != nil && req.Content != "" {
		atUserIDs := s.notificationService.ParseAtUsers(req.Content)
		if len(atUserIDs) > 0 {
			go s.notificationService.SendAtNotifications(userID, "post", post.ID, req.Content, atUserIDs)
		}
	}

	if s.pointsService != nil && status == PostStatusPublished {
		go s.pointsService.AddPoints(userID, "post_create", post.ID, "发布动态")
	}
	if status == PostStatusPublished {
		cache.InvalidateGuestFeed()
		cache.DeleteByPattern("rank:all:*")
	}

	// 作者可读待审/定时帖，不走公开 GetPostByID 的 status=1 限制
	info := s.convertToPostInfo(post)
	s.enrichPostInfo(userID, info)
	return info, nil
}

func (s *PostService) UpdatePost(userID uint, postID uint, req *types.PostUpdateRequest) (*types.PostInfo, error) {
	post, err := s.postRepo.GetByID(postID)
	if err != nil {
		return nil, errors.New("动态不存在")
	}
	if post.UserID != userID {
		return nil, errors.New("无权限修改")
	}
	if post.Video != "" && len(req.Images) > 0 {
		return nil, errors.New("视频动态不能修改为图文")
	}

	if req.Content != "" {
		post.Content = FilterSensitiveText(req.Content)
	}
	if req.Images != nil {
		var imagesStr string
		if len(req.Images) > 0 {
			imagesStr = strings.Join(req.Images, ",")
		}
		post.Images = imagesStr
	}
	if req.PetIDs != nil {
		var petIDStrs []string
		for _, id := range req.PetIDs {
			petIDStrs = append(petIDStrs, fmt.Sprintf("%d", id))
		}
		post.PetIDs = strings.Join(petIDStrs, ",")
	}
	if req.TopicIDs != nil {
		if len(req.TopicIDs) > 3 {
			return nil, errors.New("最多添加3个话题")
		}
		var topicIDStrs []string
		for _, id := range req.TopicIDs {
			topicIDStrs = append(topicIDStrs, fmt.Sprintf("%d", id))
		}
		post.TopicIDs = strings.Join(topicIDStrs, ",")
	}
	if req.Location != "" {
		post.Location = req.Location
	}
	if req.Visibility != post.Visibility {
		post.Visibility = req.Visibility
	}
	// 定时发布：传入新的 publishAt 即设置/修改定时时间（并标记为定时状态）。
	// 空字符串表示不修改原有定时设置。取消定时请走专门的"转为草稿"流程。
	if req.PublishAt != "" {
		pt, err := parsePublishAt(req.PublishAt, true)
		if err != nil {
			return nil, err
		}
		post.PublishAt = pt
		post.Status = 2
	}

	if err := s.postRepo.Update(post); err != nil {
		return nil, err
	}

	// 直接构造返回,避免 GetPostByID 对 status=2(定时发布)的过滤导致作者编辑后无法读取。
	info := s.convertToPostInfo(post)
	s.enrichPostInfo(userID, info)
	return info, nil
}

func (s *PostService) DeletePost(userID uint, postID uint) error {
	post, err := s.postRepo.GetByID(postID)
	if err != nil {
		return errors.New("动态不存在")
	}
	if post.UserID != userID {
		return errors.New("无权限删除")
	}

	return s.postRepo.Delete(postID)
}

func (s *PostService) GetPostByID(userID uint, postID uint) (*types.PostInfo, error) {
	post, err := s.postRepo.GetByID(postID)
	if err != nil {
		return nil, errors.New("动态不存在")
	}
	// 已发布：公开可见规则；待审/定时/拒绝：仅作者可见
	switch post.Status {
	case PostStatusPublished:
		// ok
	case PostStatusPending, PostStatusScheduled, PostStatusRejected:
		if userID == 0 || post.UserID != userID {
			return nil, errors.New("动态不存在")
		}
	default:
		return nil, errors.New("动态不存在")
	}

	isPublicFollower := false
	if userID > 0 && post.UserID != userID && s.followRepo != nil {
		isPublicFollower, _ = s.followRepo.ExistsPublic(userID, post.UserID)
	}
	if !CanViewPost(userID, post.UserID, post.Visibility, isPublicFollower) {
		return nil, errors.New("无权限查看")
	}

	if userID > 0 && s.blockRepo != nil && post.UserID != userID {
		blocked, err := s.blockRepo.ExistsEither(userID, post.UserID)
		if err != nil {
			return nil, err
		}
		if blocked {
			return nil, errors.New("无法互动")
		}
	}

	info := s.convertToPostInfo(post)
	s.enrichPostInfo(userID, info)
	return info, nil
}

func (s *PostService) ListPosts(userID, topicID, petID uint, postType, keyword string, page, pageSize int) (*types.PostListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 10
	}

	posts, total, err := s.postRepo.List(userID, topicID, petID, postType, keyword, page, pageSize)
	if err != nil {
		return nil, err
	}

	postInfos := make([]types.PostInfo, 0, len(posts))
	var postIDs []uint
	for _, post := range posts {
		postIDs = append(postIDs, post.ID)
		postInfos = append(postInfos, *s.convertToPostInfo(&post))
	}

	likeMap := make(map[uint]bool)
	favMap := make(map[uint]bool)
	if userID > 0 && len(postIDs) > 0 {
		likeMap, _ = s.likeRepo.GetUserLikeMap(userID, "post", postIDs)
		favMap, _ = s.favoriteRepo.GetUserFavoriteMap(userID, postIDs)
	}

	for i := range postInfos {
		postInfos[i].IsLiked = likeMap[postInfos[i].ID]
		postInfos[i].IsFavorited = favMap[postInfos[i].ID]
		s.enrichPostInfoBasic(&postInfos[i])
	}

	return &types.PostListResponse{
		List:     postInfos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *PostService) ToggleLike(userID, postID uint) (*types.LikeResult, error) {
	post, err := s.postRepo.GetByID(postID)
	if err != nil {
		return nil, errors.New("动态不存在")
	}

	if s.blockRepo != nil && post.UserID != userID {
		blocked, bErr := s.blockRepo.ExistsEither(userID, post.UserID)
		if bErr != nil {
			return nil, bErr
		}
		if blocked {
			return nil, errors.New("无法互动")
		}
	}

	exists, err := s.likeRepo.Exists(userID, "post", postID)
	if err != nil {
		return nil, err
	}

	if exists {
		if err := s.postRepo.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("user_id = ? AND target_type = ? AND target_id = ?", userID, "post", postID).Delete(&model.Like{}).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.Post{}).Where("id = ?", postID).UpdateColumn("likes_count", gorm.Expr("likes_count - 1")).Error; err != nil {
				return err
			}
			return nil
		}); err != nil {
			return nil, err
		}
		return &types.LikeResult{Liked: false, LikesCount: post.LikesCount - 1}, nil
	}

	like := &model.Like{
		UserID:     userID,
		TargetType: "post",
		TargetID:   postID,
	}
	if err := s.postRepo.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(like).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Post{}).Where("id = ?", postID).UpdateColumn("likes_count", gorm.Expr("likes_count + 1")).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if s.notificationService != nil && post.UserID != userID {
		go s.notificationService.CreateLikeNotification(userID, post.UserID, "post", postID)
	}

	if s.pointsService != nil {
		go s.pointsService.AddPoints(post.UserID, "post_liked", postID, "动态被点赞")
	}

	return &types.LikeResult{Liked: true, LikesCount: post.LikesCount + 1}, nil
}

func (s *PostService) ToggleFavorite(userID, postID uint) (*types.FavoriteResult, error) {
	post, err := s.postRepo.GetByID(postID)
	if err != nil {
		return nil, errors.New("动态不存在")
	}

	exists, err := s.favoriteRepo.Exists(userID, postID)
	if err != nil {
		return nil, err
	}

	if exists {
		if err := s.postRepo.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&model.Favorite{}).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.Post{}).Where("id = ?", postID).UpdateColumn("favorites_count", gorm.Expr("favorites_count - 1")).Error; err != nil {
				return err
			}
			return nil
		}); err != nil {
			return nil, err
		}
		return &types.FavoriteResult{Favorited: false, FavoritesCount: post.FavoritesCount - 1}, nil
	}

	favorite := &model.Favorite{
		UserID: userID,
		PostID: postID,
	}
	if err := s.postRepo.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(favorite).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Post{}).Where("id = ?", postID).UpdateColumn("favorites_count", gorm.Expr("favorites_count + 1")).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &types.FavoriteResult{Favorited: true, FavoritesCount: post.FavoritesCount + 1}, nil
}

func (s *PostService) RecordShare(userID, postID uint, platform string) error {
	_, err := s.postRepo.GetByID(postID)
	if err != nil {
		return errors.New("动态不存在")
	}

	share := &model.ShareRecord{
		UserID:   userID,
		PostID:   postID,
		Platform: platform,
	}

	return s.postRepo.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(share).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.Post{}).Where("id = ?", postID).UpdateColumn("shares_count", gorm.Expr("shares_count + 1")).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *PostService) GetFavorites(userID uint, page, pageSize int) (*types.PostListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 10
	}

	favorites, total, err := s.favoriteRepo.ListByUser(userID, page, pageSize)
	if err != nil {
		return nil, err
	}

	var postIDs []uint
	for _, fav := range favorites {
		postIDs = append(postIDs, fav.PostID)
	}

	postInfos := make([]types.PostInfo, 0)
	if len(postIDs) > 0 {
		for _, postID := range postIDs {
			post, err := s.postRepo.GetByID(postID)
			if err == nil && post.Status == 1 {
				info := s.convertToPostInfo(post)
				info.IsFavorited = true
				s.enrichPostInfoBasic(info)
				postInfos = append(postInfos, *info)
			}
		}
	}

	return &types.PostListResponse{
		List:     postInfos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *PostService) SetPinned(postID uint, pinned int) error {
	return s.postRepo.SetPinned(postID, pinned)
}

// PublishScheduledPosts 发布到期的定时动态；关闭自动审核时进入待审
func (s *PostService) PublishScheduledPosts() error {
	toStatus := PostStatusPublished
	if !isAutoAuditEnabled() {
		toStatus = PostStatusPending
	}
	if err := s.postRepo.PublishScheduled(toStatus); err != nil {
		return err
	}
	if toStatus == PostStatusPublished {
		cache.InvalidateGuestFeed()
		cache.DeleteByPattern("rank:all:*")
	}
	return nil
}

// CancelScheduleToDraft 取消定时发布并转为草稿（D7-02）
func (s *PostService) CancelScheduleToDraft(userID, postID uint) (*types.DraftInfo, error) {
	if s.draftRepo == nil {
		return nil, errors.New("草稿服务未就绪")
	}
	post, err := s.postRepo.GetByID(postID)
	if err != nil {
		return nil, errors.New("动态不存在")
	}
	if post.UserID != userID {
		return nil, errors.New("无权限操作")
	}
	if post.Status != PostStatusScheduled {
		return nil, errors.New("仅定时待发动态可取消")
	}

	draft := &model.Draft{
		UserID:     post.UserID,
		Content:    post.Content,
		Images:     post.Images,
		Video:      post.Video,
		PetIDs:     post.PetIDs,
		TopicIDs:   post.TopicIDs,
		Location:   post.Location,
		Visibility: post.Visibility,
	}
	if err := s.draftRepo.Create(draft); err != nil {
		return nil, err
	}
	if err := s.postRepo.Delete(postID); err != nil {
		_ = s.draftRepo.Delete(draft.ID)
		return nil, err
	}

	// 复用草稿转换逻辑
	ds := &DraftService{draftRepo: s.draftRepo}
	return ds.convertToDraftInfo(draft), nil
}

// GetLikeList 获取点赞列表
func (s *PostService) GetLikeList(postID uint, page, pageSize int) (*types.LikeListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 20
	}

	likes, total, err := s.likeRepo.ListByTarget("post", postID, page, pageSize)
	if err != nil {
		return nil, err
	}

	list := make([]types.LikeInfo, 0, len(likes))
	for _, like := range likes {
		user, err := s.userRepo.GetByID(like.UserID)
		if err != nil {
			continue
		}
		list = append(list, types.LikeInfo{
			ID:        like.ID,
			UserID:    like.UserID,
			Nickname:  user.Nickname,
			Avatar:    user.Avatar,
			CreatedAt: like.CreatedAt.Format(time.RFC3339),
		})
	}

	return &types.LikeListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *PostService) SetFeatured(postID uint, featured int) error {
	return s.postRepo.SetFeatured(postID, featured)
}

func (s *PostService) convertToPostInfo(post *model.Post) *types.PostInfo {
	images := []string{}
	if post.Images != "" {
		images = strings.Split(post.Images, ",")
	}

	petIDs := []uint{}
	if post.PetIDs != "" {
		for _, idStr := range strings.Split(post.PetIDs, ",") {
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				petIDs = append(petIDs, uint(id))
			}
		}
	}

	topicIDs := []uint{}
	if post.TopicIDs != "" {
		for _, idStr := range strings.Split(post.TopicIDs, ",") {
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				topicIDs = append(topicIDs, uint(id))
			}
		}
	}

	publishAt := ""
	if post.PublishAt != nil {
		publishAt = post.PublishAt.Format(time.RFC3339)
	}

	return &types.PostInfo{
		ID:             post.ID,
		UserID:         post.UserID,
		Content:        post.Content,
		Images:         images,
		Video:          post.Video,
		PetIDs:         petIDs,
		TopicIDs:       topicIDs,
		Location:       post.Location,
		Visibility:     post.Visibility,
		LikesCount:     post.LikesCount,
		CommentsCount:  post.CommentsCount,
		SharesCount:    post.SharesCount,
		FavoritesCount: post.FavoritesCount,
		IsPinned:       post.IsPinned,
		IsFeatured:     post.IsFeatured,
		Status:         post.Status,
		PublishAt:      publishAt,
		CreatedAt:      post.CreatedAt.Format(time.RFC3339),
	}
}

func (s *PostService) enrichPostInfo(userID uint, info *types.PostInfo) {
	s.enrichPostInfoBasic(info)

	if userID > 0 {
		isLiked, _ := s.likeRepo.Exists(userID, "post", info.ID)
		isFav, _ := s.favoriteRepo.Exists(userID, info.ID)
		info.IsLiked = isLiked
		info.IsFavorited = isFav
	}
}

func (s *PostService) enrichPostInfoBasic(info *types.PostInfo) {
	user, err := s.userRepo.GetByID(info.UserID)
	if err == nil {
		info.UserNickname = user.Nickname
		info.UserAvatar = user.Avatar
	}

	if len(info.TopicIDs) > 0 {
		topics, err := s.topicRepo.GetByIDs(info.TopicIDs)
		if err == nil {
			topicInfos := make([]types.TopicInfo, 0, len(topics))
			for _, t := range topics {
				topicInfos = append(topicInfos, types.TopicInfo{
					ID:          t.ID,
					Name:        t.Name,
					Description: t.Description,
					Cover:       t.Cover,
					PostsCount:  t.PostsCount,
					Status:      t.Status,
				})
			}
			info.Topics = topicInfos
		}
	}

	if len(info.PetIDs) > 0 {
		var petInfos []types.PetInfo
		for _, petID := range info.PetIDs {
			pet, err := s.petRepo.GetByID(petID)
			if err == nil {
				petInfos = append(petInfos, types.PetInfo{
					ID:     pet.ID,
					UserID: pet.UserID,
					Name:   pet.Name,
					Avatar: pet.Avatar,
				})
			}
		}
		info.Pets = petInfos
	}
}

type DraftService struct {
	draftRepo *repository.DraftRepository
}

func NewDraftService(draftRepo *repository.DraftRepository) *DraftService {
	return &DraftService{draftRepo: draftRepo}
}

func (s *DraftService) CreateDraft(userID uint, req *types.DraftCreateRequest) (*types.DraftInfo, error) {
	imagesStr := ""
	if len(req.Images) > 0 {
		imagesStr = strings.Join(req.Images, ",")
	}

	petIDsStr := ""
	if len(req.PetIDs) > 0 {
		var petIDStrs []string
		for _, id := range req.PetIDs {
			petIDStrs = append(petIDStrs, fmt.Sprintf("%d", id))
		}
		petIDsStr = strings.Join(petIDStrs, ",")
	}

	topicIDsStr := ""
	if len(req.TopicIDs) > 0 {
		var topicIDStrs []string
		for _, id := range req.TopicIDs {
			topicIDStrs = append(topicIDStrs, fmt.Sprintf("%d", id))
		}
		topicIDsStr = strings.Join(topicIDStrs, ",")
	}

	draft := &model.Draft{
		UserID:     userID,
		Content:    req.Content,
		Images:     imagesStr,
		Video:      req.Video,
		PetIDs:     petIDsStr,
		TopicIDs:   topicIDsStr,
		Location:   req.Location,
		Visibility: req.Visibility,
	}

	if err := s.draftRepo.Create(draft); err != nil {
		return nil, err
	}

	return s.convertToDraftInfo(draft), nil
}

func (s *DraftService) UpdateDraft(userID uint, draftID uint, req *types.DraftCreateRequest) (*types.DraftInfo, error) {
	draft, err := s.draftRepo.GetByID(draftID)
	if err != nil {
		return nil, errors.New("草稿不存在")
	}
	if draft.UserID != userID {
		return nil, errors.New("无权限修改")
	}

	if req.Content != "" || req.Images != nil {
		draft.Content = req.Content
		if req.Images != nil {
			imagesStr := ""
			if len(req.Images) > 0 {
				imagesStr = strings.Join(req.Images, ",")
			}
			draft.Images = imagesStr
		}
	}
	if req.Video != "" {
		draft.Video = req.Video
	}
	if req.PetIDs != nil {
		var petIDStrs []string
		for _, id := range req.PetIDs {
			petIDStrs = append(petIDStrs, fmt.Sprintf("%d", id))
		}
		draft.PetIDs = strings.Join(petIDStrs, ",")
	}
	if req.TopicIDs != nil {
		var topicIDStrs []string
		for _, id := range req.TopicIDs {
			topicIDStrs = append(topicIDStrs, fmt.Sprintf("%d", id))
		}
		draft.TopicIDs = strings.Join(topicIDStrs, ",")
	}
	if req.Location != "" {
		draft.Location = req.Location
	}
	if req.Visibility != draft.Visibility {
		draft.Visibility = req.Visibility
	}

	if err := s.draftRepo.Update(draft); err != nil {
		return nil, err
	}

	return s.convertToDraftInfo(draft), nil
}

func (s *DraftService) GetDraft(userID, draftID uint) (*types.DraftInfo, error) {
	draft, err := s.draftRepo.GetByID(draftID)
	if err != nil {
		return nil, errors.New("草稿不存在")
	}
	if draft.UserID != userID {
		return nil, errors.New("无权限查看")
	}
	return s.convertToDraftInfo(draft), nil
}

func (s *DraftService) ListDrafts(userID uint, page, pageSize int) (*types.DraftListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 10
	}

	drafts, total, err := s.draftRepo.ListByUser(userID, page, pageSize)
	if err != nil {
		return nil, err
	}

	draftInfos := make([]types.DraftInfo, 0, len(drafts))
	for _, draft := range drafts {
		draftInfos = append(draftInfos, *s.convertToDraftInfo(&draft))
	}

	return &types.DraftListResponse{
		List:     draftInfos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *DraftService) DeleteDraft(userID, draftID uint) error {
	draft, err := s.draftRepo.GetByID(draftID)
	if err != nil {
		return errors.New("草稿不存在")
	}
	if draft.UserID != userID {
		return errors.New("无权限删除")
	}
	return s.draftRepo.Delete(draftID)
}

func (s *DraftService) convertToDraftInfo(draft *model.Draft) *types.DraftInfo {
	images := []string{}
	if draft.Images != "" {
		images = strings.Split(draft.Images, ",")
	}

	petIDs := []uint{}
	if draft.PetIDs != "" {
		for _, idStr := range strings.Split(draft.PetIDs, ",") {
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				petIDs = append(petIDs, uint(id))
			}
		}
	}

	topicIDs := []uint{}
	if draft.TopicIDs != "" {
		for _, idStr := range strings.Split(draft.TopicIDs, ",") {
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				topicIDs = append(topicIDs, uint(id))
			}
		}
	}

	return &types.DraftInfo{
		ID:         draft.ID,
		UserID:     draft.UserID,
		Content:    draft.Content,
		Images:     images,
		Video:      draft.Video,
		PetIDs:     petIDs,
		TopicIDs:   topicIDs,
		Location:   draft.Location,
		Visibility: draft.Visibility,
		CreatedAt:  draft.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  draft.UpdatedAt.Format(time.RFC3339),
	}
}

type TopicService struct {
	topicRepo *repository.TopicRepository
}

func NewTopicService(topicRepo *repository.TopicRepository) *TopicService {
	return &TopicService{topicRepo: topicRepo}
}

func (s *TopicService) CreateTopic(userID uint, req *types.TopicCreateRequest) (*types.TopicInfo, error) {
	var level model.UserLevel
	if err := config.DB.Where("user_id = ?", userID).First(&level).Error; err != nil || level.Level < 5 {
		return nil, errors.New("需达到 Lv5 才可申请创建话题")
	}

	var count int64
	today := time.Now().Format("2006-01-02")
	config.DB.Model(&model.Topic{}).Where("creator_id = ? AND DATE(created_at) = ?", userID, today).Count(&count)
	if count >= 3 {
		return nil, errors.New("每天最多创建3个话题")
	}

	existing, _ := s.topicRepo.GetByName(req.Name)
	if existing != nil {
		return nil, errors.New("话题已存在")
	}

	topic := &model.Topic{
		Name:        req.Name,
		Description: req.Description,
		Cover:       req.Cover,
		CreatorID:   userID,
		Status:      0, // 待审核
	}

	if err := s.topicRepo.Create(topic); err != nil {
		return nil, err
	}

	return s.convertToTopicInfo(topic), nil
}

func (s *TopicService) GetTopic(id uint) (*types.TopicInfo, error) {
	topic, err := s.topicRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("话题不存在")
	}
	return s.convertToTopicInfo(topic), nil
}

func (s *TopicService) ListTopics(keyword string, page, pageSize int) (*types.TopicListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 20
	}

	topics, total, err := s.topicRepo.List(keyword, page, pageSize)
	if err != nil {
		return nil, err
	}

	topicInfos := make([]types.TopicInfo, 0, len(topics))
	for _, topic := range topics {
		topicInfos = append(topicInfos, *s.convertToTopicInfo(&topic))
	}

	return &types.TopicListResponse{
		List:     topicInfos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *TopicService) SearchTopics(keyword string, limit int) ([]types.TopicInfo, error) {
	if limit <= 0 || limit > 20 {
		limit = 10
	}

	topics, err := s.topicRepo.Search(keyword, limit)
	if err != nil {
		return nil, err
	}

	topicInfos := make([]types.TopicInfo, 0, len(topics))
	for _, topic := range topics {
		topicInfos = append(topicInfos, *s.convertToTopicInfo(&topic))
	}

	return topicInfos, nil
}

func (s *TopicService) convertToTopicInfo(topic *model.Topic) *types.TopicInfo {
	return &types.TopicInfo{
		ID:          topic.ID,
		Name:        topic.Name,
		Description: topic.Description,
		Cover:       topic.Cover,
		PostsCount:  topic.PostsCount,
		Status:      topic.Status,
	}
}

type ReportService struct {
	reportRepo *repository.ReportRepository
}

func NewReportService(reportRepo *repository.ReportRepository) *ReportService {
	return &ReportService{reportRepo: reportRepo}
}

func (s *ReportService) CreateReport(userID uint, req *types.ReportCreateRequest) (*types.ReportInfo, error) {
	exists, _ := s.reportRepo.Exists(userID, req.TargetType, req.TargetID)
	if exists {
		return nil, errors.New("已经举报过了")
	}

	report := &model.Report{
		ReporterID: userID,
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		Reason:     req.Reason,
		ReasonType: req.ReasonType,
		Status:     0,
	}

	if err := s.reportRepo.Create(report); err != nil {
		return nil, err
	}

	return s.convertToReportInfo(report), nil
}

func (s *ReportService) GetReport(userID, reportID uint) (*types.ReportInfo, error) {
	report, err := s.reportRepo.GetByID(reportID)
	if err != nil {
		return nil, errors.New("举报不存在")
	}
	if report.ReporterID != userID {
		return nil, errors.New("无权限查看")
	}
	return s.convertToReportInfo(report), nil
}

func (s *ReportService) ListReports(userID uint, page, pageSize int) (*types.ReportListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 10
	}

	reports, total, err := s.reportRepo.ListByUser(userID, page, pageSize)
	if err != nil {
		return nil, err
	}

	reportInfos := make([]types.ReportInfo, 0, len(reports))
	for _, report := range reports {
		reportInfos = append(reportInfos, *s.convertToReportInfo(&report))
	}

	return &types.ReportListResponse{
		List:     reportInfos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *ReportService) convertToReportInfo(report *model.Report) *types.ReportInfo {
	handledAt := ""
	if report.HandledAt != nil {
		handledAt = report.HandledAt.Format(time.RFC3339)
	}

	return &types.ReportInfo{
		ID:           report.ID,
		ReporterID:   report.ReporterID,
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

type CommentService struct {
	commentRepo         *repository.CommentRepository
	likeRepo            *repository.LikeRepository
	postRepo            *repository.PostRepository
	userRepo            *repository.UserRepository
	blockRepo           *repository.BlockRepository
	followRepo          *repository.FollowRepository
	notificationService *NotificationService
	pointsService       *PointsService
}

func NewCommentService(
	commentRepo *repository.CommentRepository,
	likeRepo *repository.LikeRepository,
	postRepo *repository.PostRepository,
	userRepo *repository.UserRepository,
) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		likeRepo:    likeRepo,
		postRepo:    postRepo,
		userRepo:    userRepo,
	}
}

func (s *CommentService) SetNotificationService(ns *NotificationService) {
	s.notificationService = ns
}

func (s *CommentService) SetPointsService(ps *PointsService) {
	s.pointsService = ps
}

func (s *CommentService) SetBlockRepository(blockRepo *repository.BlockRepository) {
	s.blockRepo = blockRepo
}

func (s *CommentService) SetFollowRepository(followRepo *repository.FollowRepository) {
	s.followRepo = followRepo
}

func (s *CommentService) CreateComment(userID uint, req *types.CommentCreateRequest) (*types.CommentInfo, error) {
	postForCheck, err := s.postRepo.GetByID(req.PostID)
	if err != nil {
		return nil, errors.New("动态不存在")
	}
	if s.blockRepo != nil && postForCheck.UserID != userID {
		blocked, bErr := s.blockRepo.ExistsEither(userID, postForCheck.UserID)
		if bErr != nil {
			return nil, bErr
		}
		if blocked {
			return nil, errors.New("无法互动")
		}
	}
	// U8 评论权限：仅粉丝 = 普通关注作者（不含悄悄关注）
	if postForCheck.UserID != userID && s.userRepo != nil {
		author, aErr := s.userRepo.GetByID(postForCheck.UserID)
		if aErr == nil && author != nil {
			_, commentPerm, _ := ParsePrivacySettings(author.PrivacySettings)
			switch commentPerm {
			case PermissionDenied:
				return nil, errors.New("作者已关闭评论")
			case PermissionFollowers:
				if s.followRepo == nil {
					return nil, errors.New("仅粉丝可评论")
				}
				ok, fErr := s.followRepo.ExistsPublic(userID, postForCheck.UserID)
				if fErr != nil {
					return nil, fErr
				}
				if !ok {
					return nil, errors.New("仅粉丝可评论")
				}
			}
		}
	}

	var parent *model.Comment
	if req.ParentID > 0 {
		parent, err = s.commentRepo.GetByID(req.ParentID)
		if err != nil {
			return nil, errors.New("父评论不存在")
		}
		if parent.ParentID > 0 {
			return nil, errors.New("最多两级评论")
		}
	}

	atUsersStr := ""
	if len(req.AtUsers) > 0 {
		var userIDStrs []string
		for _, id := range req.AtUsers {
			userIDStrs = append(userIDStrs, fmt.Sprintf("%d", id))
		}
		atUsersStr = strings.Join(userIDStrs, ",")
	}

	comment := &model.Comment{
		PostID:    req.PostID,
		UserID:    userID,
		ParentID:  req.ParentID,
		ReplyToID: req.ReplyToID,
		Content:   FilterSensitiveText(req.Content),
		AtUsers:   atUsersStr,
		Status:    0,
	}

	if err := s.commentRepo.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(comment).Error; err != nil {
			return err
		}

		if parent == nil {
			if err := tx.Model(&model.Post{}).Where("id = ?", req.PostID).UpdateColumn("comments_count", gorm.Expr("comments_count + 1")).Error; err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	post, _ := s.postRepo.GetByID(req.PostID)
	if post != nil {
		if s.notificationService != nil && userID != post.UserID {
			go s.notificationService.CreateCommentNotification(userID, post.UserID, req.PostID, req.Content)
		}

		if req.ReplyToID > 0 && s.notificationService != nil {
			replyToUser, _ := s.userRepo.GetByID(req.ReplyToID)
			if replyToUser != nil && userID != replyToUser.ID {
				go s.notificationService.CreateReplyNotification(userID, req.ReplyToID, req.PostID, comment.ID, req.Content)
			}
		}

		atUserIDs := ParseAtUsersFromString(atUsersStr)
		if s.notificationService != nil && len(atUserIDs) > 0 {
			go s.notificationService.SendAtNotifications(userID, "comment", comment.ID, req.Content, atUserIDs)
		}

		if s.pointsService != nil {
			go s.pointsService.AddPoints(userID, "comment", comment.ID, "发表评论")
			if userID != post.UserID {
				go s.pointsService.AddPoints(post.UserID, "comment_received", comment.ID, "动态被评论")
			}
		}
	}

	return s.GetCommentByID(userID, comment.ID)
}

func (s *CommentService) UpdateComment(userID, commentID uint, req *types.CommentUpdateRequest) (*types.CommentInfo, error) {
	comment, err := s.commentRepo.GetByID(commentID)
	if err != nil {
		return nil, errors.New("评论不存在")
	}
	if comment.UserID != userID {
		return nil, errors.New("无权限修改")
	}

	comment.Content = req.Content

	if err := s.commentRepo.Update(comment); err != nil {
		return nil, err
	}

	return s.GetCommentByID(userID, commentID)
}

func (s *CommentService) DeleteComment(userID, commentID uint) error {
	comment, err := s.commentRepo.GetByID(commentID)
	if err != nil {
		return errors.New("评论不存在")
	}
	if comment.UserID != userID {
		return errors.New("无权限删除")
	}

	return s.commentRepo.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(comment).Error; err != nil {
			return err
		}

		if comment.ParentID == 0 {
			if err := tx.Model(&model.Post{}).Where("id = ?", comment.PostID).UpdateColumn("comments_count", gorm.Expr("comments_count - 1")).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *CommentService) GetCommentByID(userID, commentID uint) (*types.CommentInfo, error) {
	comment, err := s.commentRepo.GetByID(commentID)
	if err != nil {
		return nil, errors.New("评论不存在")
	}

	info := s.convertToCommentInfo(comment)
	s.enrichCommentInfo(userID, info)
	return info, nil
}

func (s *CommentService) ListComments(userID, postID uint, sort string, page, pageSize int) (*types.CommentListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 20
	}

	comments, total, err := s.commentRepo.ListByPost(postID, sort, page, pageSize)
	if err != nil {
		return nil, err
	}

	commentInfos := make([]types.CommentInfo, 0, len(comments))
	var commentIDs []uint
	for _, c := range comments {
		commentIDs = append(commentIDs, c.ID)
		commentInfos = append(commentInfos, *s.convertToCommentInfo(&c))
	}

	likeMap := make(map[uint]bool)
	if userID > 0 && len(commentIDs) > 0 {
		likeMap, _ = s.likeRepo.GetUserLikeMap(userID, "comment", commentIDs)
	}

	for i := range commentInfos {
		commentInfos[i].IsLiked = likeMap[commentInfos[i].ID]
		s.enrichCommentInfoBasic(&commentInfos[i])

		replies, err := s.commentRepo.GetReplies(commentInfos[i].ID)
		if err == nil && len(replies) > 0 {
			children := make([]types.CommentInfo, 0, len(replies))
			for _, reply := range replies {
				childInfo := s.convertToCommentInfo(&reply)
				s.enrichCommentInfoBasic(childInfo)
				if userID > 0 {
					isLiked, _ := s.likeRepo.Exists(userID, "comment", reply.ID)
					childInfo.IsLiked = isLiked
				}
				children = append(children, *childInfo)
			}
			commentInfos[i].Children = children
		}
	}

	return &types.CommentListResponse{
		List:     commentInfos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *CommentService) ToggleLike(userID, commentID uint) (*types.LikeResult, error) {
	comment, err := s.commentRepo.GetByID(commentID)
	if err != nil {
		return nil, errors.New("评论不存在")
	}

	exists, err := s.likeRepo.Exists(userID, "comment", commentID)
	if err != nil {
		return nil, err
	}

	if exists {
		if err := s.commentRepo.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("user_id = ? AND target_type = ? AND target_id = ?", userID, "comment", commentID).Delete(&model.Like{}).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.Comment{}).Where("id = ?", commentID).UpdateColumn("likes_count", gorm.Expr("likes_count - 1")).Error; err != nil {
				return err
			}
			return nil
		}); err != nil {
			return nil, err
		}
		return &types.LikeResult{Liked: false, LikesCount: comment.LikesCount - 1}, nil
	}

	like := &model.Like{
		UserID:     userID,
		TargetType: "comment",
		TargetID:   commentID,
	}
	if err := s.commentRepo.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(like).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Comment{}).Where("id = ?", commentID).UpdateColumn("likes_count", gorm.Expr("likes_count + 1")).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &types.LikeResult{Liked: true, LikesCount: comment.LikesCount + 1}, nil
}

func (s *CommentService) convertToCommentInfo(comment *model.Comment) *types.CommentInfo {
	atUsers := []uint{}
	if comment.AtUsers != "" {
		for _, idStr := range strings.Split(comment.AtUsers, ",") {
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				atUsers = append(atUsers, uint(id))
			}
		}
	}

	return &types.CommentInfo{
		ID:         comment.ID,
		PostID:     comment.PostID,
		UserID:     comment.UserID,
		ParentID:   comment.ParentID,
		ReplyToID:  comment.ReplyToID,
		Content:    comment.Content,
		AtUsers:    atUsers,
		LikesCount: comment.LikesCount,
		Status:     comment.Status,
		CreatedAt:  comment.CreatedAt.Format(time.RFC3339),
	}
}

func (s *CommentService) enrichCommentInfo(userID uint, info *types.CommentInfo) {
	s.enrichCommentInfoBasic(info)

	if userID > 0 {
		isLiked, _ := s.likeRepo.Exists(userID, "comment", info.ID)
		info.IsLiked = isLiked
	}
}

func (s *CommentService) enrichCommentInfoBasic(info *types.CommentInfo) {
	user, err := s.userRepo.GetByID(info.UserID)
	if err == nil {
		info.UserNickname = user.Nickname
		info.UserAvatar = user.Avatar
	}

	if info.ReplyToID > 0 {
		replyUser, err := s.userRepo.GetByID(info.ReplyToID)
		if err == nil {
			info.ReplyToName = replyUser.Nickname
		}
	}
}
