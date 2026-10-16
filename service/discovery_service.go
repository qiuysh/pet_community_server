package service

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"pet-community/cache"
	"pet-community/config"
	"pet-community/model"
	"pet-community/repository"
	"pet-community/types"

	"gorm.io/gorm"
)

type DiscoveryService struct {
	postRepo             *repository.PostRepository
	userRepo             *repository.UserRepository
	topicRepo            *repository.TopicRepository
	likeRepo             *repository.LikeRepository
	favoriteRepo         *repository.FavoriteRepository
	petRepo              *repository.PetRepository
	searchHistoryRepo    *repository.SearchHistoryRepository
	userInterestRepo     *repository.UserInterestRepository
	topicCategoryRepo    *repository.TopicCategoryRepository
	topicParticipantRepo *repository.TopicParticipantRepository
	hotKeywordRepo       *repository.HotKeywordRepository
	followRepo           *repository.FollowRepository
	blockRepo            *repository.BlockRepository
}

func NewDiscoveryService(
	postRepo *repository.PostRepository,
	userRepo *repository.UserRepository,
	topicRepo *repository.TopicRepository,
	likeRepo *repository.LikeRepository,
	favoriteRepo *repository.FavoriteRepository,
	petRepo *repository.PetRepository,
	searchHistoryRepo *repository.SearchHistoryRepository,
	userInterestRepo *repository.UserInterestRepository,
	topicCategoryRepo *repository.TopicCategoryRepository,
	topicParticipantRepo *repository.TopicParticipantRepository,
	hotKeywordRepo *repository.HotKeywordRepository,
	followRepo *repository.FollowRepository,
) *DiscoveryService {
	return &DiscoveryService{
		postRepo:             postRepo,
		userRepo:             userRepo,
		topicRepo:            topicRepo,
		likeRepo:             likeRepo,
		favoriteRepo:         favoriteRepo,
		petRepo:              petRepo,
		searchHistoryRepo:    searchHistoryRepo,
		userInterestRepo:     userInterestRepo,
		topicCategoryRepo:    topicCategoryRepo,
		topicParticipantRepo: topicParticipantRepo,
		hotKeywordRepo:       hotKeywordRepo,
		followRepo:           followRepo,
	}
}

func (s *DiscoveryService) SetBlockRepository(blockRepo *repository.BlockRepository) {
	s.blockRepo = blockRepo
}

// applyVisibilityFilter 公开 / 仅粉丝（普通关注）/ 自己的私密；悄悄关注不计粉丝可见
func (s *DiscoveryService) applyVisibilityFilter(query *gorm.DB, userID uint) *gorm.DB {
	if userID == 0 {
		return query.Where("visibility = ?", VisibilityPublic)
	}
	if s.followRepo == nil {
		return query.Where("visibility = ? OR user_id = ?", VisibilityPublic, userID)
	}
	followingIDs, err := s.followRepo.GetPublicFollowingIDs(userID)
	if err != nil || len(followingIDs) == 0 {
		return query.Where("visibility = ? OR user_id = ?", VisibilityPublic, userID)
	}
	return query.Where(
		"visibility = ? OR user_id = ? OR (visibility = ? AND user_id IN ?)",
		VisibilityPublic, userID, VisibilityFollowers, followingIDs,
	)
}

const postHotOrderSQL = "(likes_count + comments_count * 2 + shares_count * 3) DESC, created_at DESC"

func (s *DiscoveryService) Search(keyword string, userID uint) (*types.SearchResult, error) {
	if keyword == "" {
		return nil, errors.New("搜索关键词不能为空")
	}

	s.recordSearch(keyword, userID)

	var wg sync.WaitGroup
	result := &types.SearchResult{}

	wg.Add(3)

	go func() {
		defer wg.Done()
		users, _ := s.searchUsers(keyword, 10)
		result.Users = users
	}()

	go func() {
		defer wg.Done()
		posts, _ := s.searchPosts(keyword, userID, 10)
		result.Posts = posts
	}()

	go func() {
		defer wg.Done()
		topics, _ := s.searchTopics(keyword, 10)
		result.Topics = topics
	}()

	wg.Wait()

	return result, nil
}

func (s *DiscoveryService) searchUsers(keyword string, limit int) ([]types.UserSearchResult, error) {
	users, err := s.userRepo.SearchByNickname(keyword, limit)
	if err != nil {
		return nil, err
	}

	result := make([]types.UserSearchResult, 0, len(users))
	for _, user := range users {
		followersCount, _ := s.followRepo.GetFollowersCount(user.ID)
		result = append(result, types.UserSearchResult{
			ID:             user.ID,
			Nickname:       user.Nickname,
			Avatar:         user.Avatar,
			FollowersCount: int(followersCount),
		})
	}
	return result, nil
}

func (s *DiscoveryService) searchPosts(keyword string, userID uint, limit int) ([]types.PostSearchResult, error) {
	var posts []model.Post
	if err := config.DB.Where("status = 1 AND visibility = 0 AND content LIKE ?", "%"+keyword+"%").
		Limit(limit).Order("created_at DESC").Find(&posts).Error; err != nil {
		return nil, err
	}

	result := make([]types.PostSearchResult, 0, len(posts))
	for _, post := range posts {
		user, _ := s.userRepo.GetByID(post.UserID)
		if user == nil {
			continue
		}
		result = append(result, types.ConvertPostToSearchResult(&post, user))
	}
	return result, nil
}

func (s *DiscoveryService) searchTopics(keyword string, limit int) ([]types.TopicSearchResult, error) {
	topics, err := s.topicRepo.Search(keyword, limit)
	if err != nil {
		return nil, err
	}

	result := make([]types.TopicSearchResult, 0, len(topics))
	for _, topic := range topics {
		result = append(result, types.ConvertTopicToSearchResult(&topic))
	}
	return result, nil
}

func (s *DiscoveryService) recordSearch(keyword string, userID uint) {
	if userID > 0 {
		exists, _ := s.searchHistoryRepo.Exists(userID, keyword)
		if !exists {
			s.searchHistoryRepo.Create(&model.SearchHistory{
				UserID:  userID,
				Keyword: keyword,
			})
		}
	}

	_, err := s.hotKeywordRepo.GetByKeyword(keyword)
	if err == nil {
		s.hotKeywordRepo.IncrementSearchCount(keyword)
	} else {
		s.hotKeywordRepo.Create(&model.HotKeyword{
			Keyword:     keyword,
			SearchCount: 1,
		})
	}
	cache.IncrHotKeyword(keyword)
}

func (s *DiscoveryService) GetSearchHistory(userID uint, limit int) ([]types.SearchHistoryInfo, error) {
	histories, err := s.searchHistoryRepo.ListByUser(userID, limit)
	if err != nil {
		return nil, err
	}

	result := make([]types.SearchHistoryInfo, 0, len(histories))
	for _, h := range histories {
		result = append(result, types.ConvertSearchHistoryToInfo(&h))
	}
	return result, nil
}

func (s *DiscoveryService) DeleteSearchHistory(userID, id uint) error {
	return s.searchHistoryRepo.Delete(userID, id)
}

func (s *DiscoveryService) ClearSearchHistory(userID uint) error {
	return s.searchHistoryRepo.Clear(userID)
}

func (s *DiscoveryService) GetHotKeywords(limit int) ([]types.HotKeywordInfo, error) {
	if limit <= 0 || limit > 20 {
		limit = 10
	}

	if cached, ok := cache.GetHotKeywords(limit); ok {
		return cached, nil
	}
	if fromZ := cache.TopHotKeywordsFromZSet(limit); len(fromZ) > 0 {
		cache.SetHotKeywords(limit, fromZ)
		return fromZ, nil
	}

	keywords, err := s.hotKeywordRepo.List(limit)
	if err != nil {
		return nil, err
	}

	result := make([]types.HotKeywordInfo, 0, len(keywords))
	for _, kw := range keywords {
		result = append(result, types.ConvertHotKeywordToInfo(&kw))
	}
	cache.SetHotKeywords(limit, result)
	return result, nil
}

func (s *DiscoveryService) GetRecommendPosts(userID uint, page, pageSize int, postType string) (*types.RecommendResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 20
	}

	// 游客热门流可缓存；登录兴趣流不做整页缓存
	if userID == 0 {
		if cached, ok := cache.GetGuestFeed(page, pageSize, postType); ok {
			return cached, nil
		}
	}

	var posts []model.Post
	var total int64

	query := config.DB.Model(&model.Post{}).Where("status = 1")
	query = s.applyVisibilityFilter(query, userID)
	if userID > 0 && s.blockRepo != nil {
		if blockedIDs, err := s.blockRepo.GetBlockedRelationIDs(userID); err == nil && len(blockedIDs) > 0 {
			query = query.Where("user_id NOT IN ?", blockedIDs)
		}
	}

	if postType == "image" {
		query = query.Where("images != '' AND video = ''")
	} else if postType == "video" {
		query = query.Where("video != ''")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 登录用户：热度 + 兴趣加权重排（R1-01）；游客：纯热度分页
	interestWeights := map[string]float64{}
	useInterest := false
	if userID > 0 && s.userInterestRepo != nil {
		if interests, err := s.userInterestRepo.ListByUser(userID); err == nil && len(interests) > 0 {
			useInterest = true
			for _, it := range interests {
				interestWeights[it.Interest] = it.Weight
			}
		}
	}

	if useInterest {
		candidateLimit := page * pageSize * 3
		if candidateLimit < 60 {
			candidateLimit = 60
		}
		if candidateLimit > 300 {
			candidateLimit = 300
		}
		base := query.Session(&gorm.Session{})
		if err := base.Order("is_pinned DESC, is_featured DESC, " + postHotOrderSQL).
			Limit(candidateLimit).Offset(0).Find(&posts).Error; err != nil {
			return nil, err
		}
		posts = s.rankPostsByInterest(posts, interestWeights)
		start := (page - 1) * pageSize
		if start >= len(posts) {
			posts = []model.Post{}
		} else {
			end := start + pageSize
			if end > len(posts) {
				end = len(posts)
			}
			posts = posts[start:end]
		}
	} else {
		if err := query.Order("is_pinned DESC, is_featured DESC, " + postHotOrderSQL).
			Limit(pageSize).Offset((page - 1) * pageSize).Find(&posts).Error; err != nil {
			return nil, err
		}
	}

	result := s.convertPostsToRecommend(posts, userID)
	resp := &types.RecommendResponse{
		List:     result,
		Total:    int(total),
		Page:     page,
		PageSize: pageSize,
	}
	if userID == 0 {
		cache.SetGuestFeed(page, pageSize, postType, resp)
	}
	return resp, nil
}

func (s *DiscoveryService) rankPostsByInterest(posts []model.Post, weights map[string]float64) []model.Post {
	type scored struct {
		post  model.Post
		score float64
	}
	items := make([]scored, 0, len(posts))
	topicCache := map[uint]string{}
	for _, post := range posts {
		heat := float64(post.LikesCount + post.CommentsCount*2 + post.SharesCount*3)
		boost := 0.0
		if post.TopicIDs != "" {
			for _, idStr := range strings.Split(post.TopicIDs, ",") {
				id, err := strconv.ParseUint(strings.TrimSpace(idStr), 10, 32)
				if err != nil {
					continue
				}
				name, ok := topicCache[uint(id)]
				if !ok {
					if topic, tErr := s.topicRepo.GetByID(uint(id)); tErr == nil && topic != nil {
						name = topic.Name
						topicCache[uint(id)] = name
					}
				}
				if w, ok := weights[name]; ok {
					boost += w
				}
			}
		}
		// 置顶/精选保持绝对优先；兴趣加权叠加于热度
		score := heat + boost*10
		if post.IsPinned == 1 {
			score += 1e9
		}
		if post.IsFeatured == 1 {
			score += 1e6
		}
		items = append(items, scored{post: post, score: score})
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].score > items[j].score
	})
	out := make([]model.Post, len(items))
	for i, it := range items {
		out[i] = it.post
	}
	return out
}

func (s *DiscoveryService) convertPostsToRecommend(posts []model.Post, userID uint) []types.RecommendPostInfo {
	result := make([]types.RecommendPostInfo, 0, len(posts))

	var postIDs []uint
	for _, p := range posts {
		postIDs = append(postIDs, p.ID)
	}

	likeMap := make(map[uint]bool)
	favMap := make(map[uint]bool)
	if userID > 0 && len(postIDs) > 0 {
		likeMap, _ = s.likeRepo.GetUserLikeMap(userID, "post", postIDs)
		favMap, _ = s.favoriteRepo.GetUserFavoriteMap(userID, postIDs)
	}

	for _, post := range posts {
		user, _ := s.userRepo.GetByID(post.UserID)
		if user == nil {
			continue
		}

		images := []string{}
		if post.Images != "" {
			images = strings.Split(post.Images, ",")
		}

		lv := userLevelOf(user.ID)
		result = append(result, types.RecommendPostInfo{
			ID:            post.ID,
			UserID:        post.UserID,
			UserNickname:  user.Nickname,
			UserAvatar:    user.Avatar,
			UserLevel:     lv,
			IsVipName:     lv >= 10,
			Content:       post.Content,
			Images:        images,
			Video:         post.Video,
			LikesCount:    post.LikesCount,
			CommentsCount: post.CommentsCount,
			SharesCount:   post.SharesCount,
			IsPinned:      post.IsPinned,
			IsFeatured:    post.IsFeatured,
			IsLiked:       likeMap[post.ID],
			IsFavorited:   favMap[post.ID],
			CreatedAt:     post.CreatedAt.Format(time.RFC3339),
		})
	}

	return result
}

func (s *DiscoveryService) GetCategoryPosts(userID, categoryID uint, petType, postType, sort string, page, pageSize int) (*types.CategoryPostResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 20
	}

	query := config.DB.Model(&model.Post{}).Where("status = 1")
	query = s.applyVisibilityFilter(query, userID)
	if userID > 0 && s.blockRepo != nil {
		if blockedIDs, err := s.blockRepo.GetBlockedRelationIDs(userID); err == nil && len(blockedIDs) > 0 {
			query = query.Where("user_id NOT IN ?", blockedIDs)
		}
	}

	// 按话题分类：仅当该分类下已有话题时收窄帖子范围
	if categoryID > 0 {
		var topicIDs []uint
		config.DB.Model(&model.Topic{}).Where("category_id = ? AND status = 1", categoryID).Pluck("id", &topicIDs)
		if len(topicIDs) > 0 {
			conds := make([]string, 0, len(topicIDs))
			args := make([]interface{}, 0, len(topicIDs))
			for _, tid := range topicIDs {
				conds = append(conds, "topic_ids LIKE ?")
				args = append(args, "%"+fmt.Sprintf("%d", tid)+"%")
			}
			query = query.Where("("+strings.Join(conds, " OR ")+")", args...)
		}
	}

	if petType != "" {
		petIDs := s.getPetIDsByType(petType)
		if len(petIDs) > 0 {
			petIDStrs := make([]string, 0, len(petIDs))
			for _, id := range petIDs {
				petIDStrs = append(petIDStrs, fmt.Sprintf("%d", id))
			}
			query = query.Where("pet_ids REGEXP ?", "("+strings.Join(petIDStrs, "|")+")")
		} else {
			// 指定了宠物品类但无匹配宠物 → 空结果
			return &types.CategoryPostResponse{List: []types.RecommendPostInfo{}, Total: 0, Page: page, PageSize: pageSize}, nil
		}
	}

	if postType == "image" {
		query = query.Where("images != '' AND video = ''")
	} else if postType == "video" {
		query = query.Where("video != ''")
	} else if postType == "text" {
		query = query.Where("(images = '' OR images IS NULL) AND (video = '' OR video IS NULL)")
	}

	orderStr := "created_at DESC"
	if sort == "hot" {
		orderStr = postHotOrderSQL
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var posts []model.Post
	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order(orderStr).Find(&posts).Error; err != nil {
		return nil, err
	}

	result := s.convertPostsToRecommend(posts, userID)

	return &types.CategoryPostResponse{
		List:     result,
		Total:    int(total),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *DiscoveryService) getPetIDsByType(petType string) []uint {
	var pets []model.Pet
	config.DB.Where("type = ?", petType).Find(&pets)
	result := make([]uint, 0, len(pets))
	for _, p := range pets {
		result = append(result, p.ID)
	}
	return result
}

func (s *DiscoveryService) GetTagCloud() (*types.TagCloudResponse, error) {
	topics, _, err := s.topicRepo.List("", 1, 50)
	if err != nil {
		return nil, err
	}

	tags := make([]types.TagCloudItem, 0, len(topics))
	maxCount := 0
	for _, topic := range topics {
		if topic.PostsCount > maxCount {
			maxCount = topic.PostsCount
		}
	}

	for _, topic := range topics {
		hot := 1
		if maxCount > 0 {
			ratio := float64(topic.PostsCount) / float64(maxCount)
			if ratio > 0.7 {
				hot = 3
			} else if ratio > 0.3 {
				hot = 2
			}
		}
		tags = append(tags, types.TagCloudItem{
			ID:    topic.ID,
			Name:  topic.Name,
			Count: topic.PostsCount,
			Hot:   hot,
		})
	}

	return &types.TagCloudResponse{Tags: tags}, nil
}

func (s *DiscoveryService) GetHotRanking(period string) (*types.RankingResponse, error) {
	switch period {
	case "week", "month", "day":
	default:
		period = "day"
	}

	if cached, ok := cache.GetRanking(period); ok {
		return cached, nil
	}

	var wg sync.WaitGroup
	result := &types.RankingResponse{
		Posts:  make([]types.RankingPost, 0),
		Users:  make([]types.RankingUser, 0),
		Topics: make([]types.RankingTopic, 0),
	}

	wg.Add(3)

	go func() {
		defer wg.Done()
		posts, _ := s.getPostRanking(period, 20)
		result.Posts = posts
	}()

	go func() {
		defer wg.Done()
		users, _ := s.getUserRanking(period, 20)
		result.Users = users
	}()

	go func() {
		defer wg.Done()
		topics, _ := s.getTopicRanking(period, 20)
		result.Topics = topics
	}()

	wg.Wait()
	cache.SetRanking(period, result)
	return result, nil
}

func (s *DiscoveryService) getPostRanking(period string, limit int) ([]types.RankingPost, error) {
	var posts []model.Post
	var duration time.Duration

	switch period {
	case "week":
		duration = 7 * 24 * time.Hour
	case "month":
		duration = 30 * 24 * time.Hour
	default:
		duration = 24 * time.Hour
	}

	since := time.Now().Add(-duration)

	if err := config.DB.Where("status = 1 AND visibility = 0 AND created_at >= ?", since).
		Order("(likes_count + comments_count * 2 + shares_count * 3) DESC").
		Limit(limit).Find(&posts).Error; err != nil {
		return nil, err
	}

	result := make([]types.RankingPost, 0, len(posts))
	for _, post := range posts {
		user, _ := s.userRepo.GetByID(post.UserID)
		if user == nil {
			continue
		}
		hotScore := post.LikesCount + post.CommentsCount*2 + post.SharesCount*3
		result = append(result, types.RankingPost{
			ID:            post.ID,
			UserID:        post.UserID,
			UserNickname:  user.Nickname,
			UserAvatar:    user.Avatar,
			Content:       post.Content,
			LikesCount:    post.LikesCount,
			CommentsCount: post.CommentsCount,
			SharesCount:   post.SharesCount,
			HotScore:      hotScore,
			CreatedAt:     post.CreatedAt.Format(time.RFC3339),
		})
	}
	return result, nil
}

func (s *DiscoveryService) getUserRanking(period string, limit int) ([]types.RankingUser, error) {
	users, err := s.userRepo.FindActiveUsers(limit)
	if err != nil {
		return nil, err
	}

	result := make([]types.RankingUser, 0, len(users))
	for _, user := range users {
		followersCount, _ := s.followRepo.GetFollowersCount(user.ID)

		postsCount, _ := s.postRepo.CountByUser(user.ID)

		hotScore := int(followersCount) + int(postsCount)*5
		result = append(result, types.RankingUser{
			ID:             user.ID,
			Nickname:       user.Nickname,
			Avatar:         user.Avatar,
			FollowersCount: int(followersCount),
			PostsCount:     int(postsCount),
			HotScore:       hotScore,
		})
	}
	return result, nil
}

func (s *DiscoveryService) getTopicRanking(period string, limit int) ([]types.RankingTopic, error) {
	topics, _, err := s.topicRepo.List("", 1, limit)
	if err != nil {
		return nil, err
	}

	result := make([]types.RankingTopic, 0, len(topics))
	for _, topic := range topics {
		hotScore := topic.PostsCount * 10
		result = append(result, types.RankingTopic{
			ID:          topic.ID,
			Name:        topic.Name,
			Description: topic.Description,
			Cover:       topic.Cover,
			PostsCount:  topic.PostsCount,
			HotScore:    hotScore,
		})
	}
	return result, nil
}

func (s *DiscoveryService) GetTopicCategories() ([]types.TopicCategoryInfo, error) {
	categories, err := s.topicCategoryRepo.List(1)
	if err != nil {
		return nil, err
	}

	result := make([]types.TopicCategoryInfo, 0, len(categories))
	for _, cat := range categories {
		result = append(result, types.ConvertTopicCategoryToInfo(&cat))
	}
	return result, nil
}

func (s *DiscoveryService) GetTopicDetail(topicID, userID uint) (*types.TopicDetailInfo, error) {
	topic, err := s.topicRepo.GetByID(topicID)
	if err != nil {
		return nil, errors.New("话题不存在")
	}
	if topic.Status != 1 {
		return nil, errors.New("话题不存在")
	}

	participantsCount, _ := s.topicParticipantRepo.CountByTopic(topicID)
	isParticipated, _ := s.topicParticipantRepo.Exists(topicID, userID)

	return &types.TopicDetailInfo{
		ID:                topic.ID,
		Name:              topic.Name,
		Description:       topic.Description,
		Cover:             topic.Cover,
		PostsCount:        topic.PostsCount,
		Status:            topic.Status,
		ParticipantsCount: int(participantsCount),
		IsParticipated:    isParticipated,
		CreatedAt:         topic.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *DiscoveryService) GetTopicPosts(topicID, userID uint, page, pageSize int) (*types.CategoryPostResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 20
	}

	query := config.DB.Model(&model.Post{}).Where("status = 1")
	query = s.applyVisibilityFilter(query, userID)

	query = query.Where("topic_ids LIKE ?", "%"+fmt.Sprintf("%d", topicID)+"%")

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var posts []model.Post
	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("created_at DESC").Find(&posts).Error; err != nil {
		return nil, err
	}

	result := s.convertPostsToRecommend(posts, userID)

	return &types.CategoryPostResponse{
		List:     result,
		Total:    int(total),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *DiscoveryService) ParticipateTopic(topicID, userID uint) error {
	topic, err := s.topicRepo.GetByID(topicID)
	if err != nil {
		return errors.New("话题不存在")
	}
	if topic.Status != 1 {
		return errors.New("话题不存在")
	}

	exists, err := s.topicParticipantRepo.Exists(topicID, userID)
	if err != nil {
		return err
	}
	if exists {
		return s.topicParticipantRepo.Delete(topicID, userID)
	}

	return s.topicParticipantRepo.Create(&model.TopicParticipant{
		TopicID: topicID,
		UserID:  userID,
	})
}

// ShareTopic 生成话题分享链接（前端可据此复制/出二维码）
func (s *DiscoveryService) ShareTopic(topicID uint, platform string) (*types.TopicShareResponse, error) {
	topic, err := s.topicRepo.GetByID(topicID)
	if err != nil {
		return nil, errors.New("话题不存在")
	}
	if topic.Status != 1 {
		return nil, errors.New("话题不存在")
	}
	_ = platform
	return &types.TopicShareResponse{
		ShareURL: fmt.Sprintf("/topics/%d", topic.ID),
	}, nil
}

func (s *DiscoveryService) CreateTopic(userID uint, req *types.TopicCreateRequest) (*types.TopicDetailInfo, error) {
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
		Status:      0, // 待审核；运营通过后置 1
	}

	if err := s.topicRepo.Create(topic); err != nil {
		return nil, err
	}

	// 待审核话题不可走公开详情；返回创建结果摘要
	return &types.TopicDetailInfo{
		ID:          topic.ID,
		Name:        topic.Name,
		Description: topic.Description,
		Cover:       topic.Cover,
		PostsCount:  0,
		Status:      topic.Status,
		CreatedAt:   topic.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *DiscoveryService) TrackUserInterest(userID uint, actionType string, targetID uint) {
	if userID == 0 {
		return
	}

	switch actionType {
	case "view_post":
		s.trackPostInterest(userID, targetID)
	case "like_post":
		s.trackPostInterest(userID, targetID)
	case "favorite_post":
		s.trackPostInterest(userID, targetID)
	case "view_topic":
		s.trackTopicInterest(userID, targetID)
	}
}

func (s *DiscoveryService) trackPostInterest(userID, postID uint) {
	post, err := s.postRepo.GetByID(postID)
	if err != nil {
		return
	}

	if post.TopicIDs != "" {
		for _, idStr := range strings.Split(post.TopicIDs, ",") {
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				topic, _ := s.topicRepo.GetByID(uint(id))
				if topic != nil {
					s.updateInterest(userID, topic.Name)
				}
			}
		}
	}
}

func (s *DiscoveryService) trackTopicInterest(userID, topicID uint) {
	topic, err := s.topicRepo.GetByID(topicID)
	if err != nil {
		return
	}
	s.updateInterest(userID, topic.Name)
}

func (s *DiscoveryService) updateInterest(userID uint, interest string) {
	existing, err := s.userInterestRepo.GetByUserAndInterest(userID, interest)
	if err != nil {
		s.userInterestRepo.Create(&model.UserInterest{
			UserID:     userID,
			Interest:   interest,
			Weight:     1.0,
			LastActive: time.Now(),
		})
	} else {
		existing.Weight += 0.5
		existing.LastActive = time.Now()
		s.userInterestRepo.Update(existing)
	}
}

func (s *DiscoveryService) getDB() *gorm.DB {
	return config.DB
}
