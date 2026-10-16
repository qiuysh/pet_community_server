package repository

import (
	"fmt"
	"pet-community/model"
	"time"

	"gorm.io/gorm"
)

type PostRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Create(post *model.Post) error {
	return r.db.Create(post).Error
}

func (r *PostRepository) Update(post *model.Post) error {
	return r.db.Save(post).Error
}

func (r *PostRepository) GetByID(id uint) (*model.Post, error) {
	var post model.Post
	if err := r.db.First(&post, id).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *PostRepository) Delete(id uint) error {
	return r.db.Delete(&model.Post{}, id).Error
}

// List 列出动态。viewerID 为当前访客；authorID>0 时按作者筛选（个人主页）。
// 未指定作者时返回公开动态流（不因登录态误筛成「仅自己的帖」）。
func (r *PostRepository) List(viewerID, authorID, topicID, petID uint, postType, keyword string, page, pageSize int) ([]model.Post, int64, error) {
	var posts []model.Post
	var total int64
	query := r.db.Model(&model.Post{}).Where("status = 1")

	if authorID > 0 {
		query = query.Where("user_id = ?", authorID)
		if viewerID == authorID {
			// 本人：可见全部可见性
		} else if viewerID == 0 {
			query = query.Where("visibility = 0")
		} else {
			// 他人：公开 + 粉丝可见（粉丝关系在 service 再细滤）
			query = query.Where("visibility IN (0, 1)")
		}
	} else {
		query = query.Where("visibility = 0")
	}

	if topicID > 0 {
		query = query.Where("topic_ids LIKE ?", "%"+fmt.Sprintf("%d", topicID)+"%")
	}

	if petID > 0 {
		query = query.Where("pet_ids LIKE ?", "%"+fmt.Sprintf("%d", petID)+"%")
	}

	if postType == "image" {
		query = query.Where("images != '' AND video = ''")
	} else if postType == "video" {
		query = query.Where("video != ''")
	}

	if keyword != "" {
		query = query.Where("content LIKE ?", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("is_pinned DESC, id DESC").Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

func (r *PostRepository) IncrementLikes(id uint, delta int) error {
	return r.db.Model(&model.Post{}).Where("id = ?", id).UpdateColumn("likes_count", gorm.Expr("likes_count + ?", delta)).Error
}

func (r *PostRepository) IncrementComments(id uint, delta int) error {
	return r.db.Model(&model.Post{}).Where("id = ?", id).UpdateColumn("comments_count", gorm.Expr("comments_count + ?", delta)).Error
}

func (r *PostRepository) IncrementShares(id uint) error {
	return r.db.Model(&model.Post{}).Where("id = ?", id).UpdateColumn("shares_count", gorm.Expr("shares_count + 1")).Error
}

func (r *PostRepository) IncrementFavorites(id uint, delta int) error {
	return r.db.Model(&model.Post{}).Where("id = ?", id).UpdateColumn("favorites_count", gorm.Expr("favorites_count + ?", delta)).Error
}

func (r *PostRepository) SetPinned(id uint, pinned int) error {
	return r.db.Model(&model.Post{}).Where("id = ?", id).Update("is_pinned", pinned).Error
}

func (r *PostRepository) SetFeatured(id uint, featured int) error {
	return r.db.Model(&model.Post{}).Where("id = ?", id).Update("is_featured", featured).Error
}

// PublishScheduled 将到期定时帖更新为目标状态（1已发布 / 0待审），并清空 publish_at
func (r *PostRepository) PublishScheduled(toStatus int) error {
	return r.db.Model(&model.Post{}).
		Where("status = 2 AND publish_at <= ?", time.Now()).
		Updates(map[string]interface{}{
			"status":     toStatus,
			"publish_at": nil,
		}).Error
}

func (r *PostRepository) CountByUser(userID uint) (int64, error) {
	var count int64
	return count, r.db.Model(&model.Post{}).Where("user_id = ? AND status = 1", userID).Count(&count).Error
}

func (r *PostRepository) CountAll() (int64, error) {
	var count int64
	return count, r.db.Model(&model.Post{}).Count(&count).Error
}

func (r *PostRepository) CountByDate(date string) (int64, error) {
	var count int64
	return count, r.db.Model(&model.Post{}).Where("DATE(created_at) = ?", date).Count(&count).Error
}

func (r *PostRepository) Transaction(fn func(*gorm.DB) error) error {
	return r.db.Transaction(fn)
}

type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Create(comment *model.Comment) error {
	return r.db.Create(comment).Error
}

func (r *CommentRepository) Update(comment *model.Comment) error {
	return r.db.Save(comment).Error
}

func (r *CommentRepository) GetByID(id uint) (*model.Comment, error) {
	var comment model.Comment
	if err := r.db.First(&comment, id).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *CommentRepository) Delete(id uint) error {
	return r.db.Delete(&model.Comment{}, id).Error
}

func (r *CommentRepository) ListByPost(postID uint, sort string, page, pageSize int) ([]model.Comment, int64, error) {
	var comments []model.Comment
	var total int64
	query := r.db.Model(&model.Comment{}).Where("post_id = ? AND parent_id = 0 AND status = 1", postID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderStr := "id DESC"
	if sort == "hot" {
		orderStr = "likes_count DESC, id DESC"
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order(orderStr).Find(&comments).Error; err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

func (r *CommentRepository) GetReplies(parentID uint) ([]model.Comment, error) {
	var comments []model.Comment
	if err := r.db.Where("parent_id = ? AND status = 1", parentID).Order("id ASC").Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}

func (r *CommentRepository) CountAll() (int64, error) {
	var count int64
	return count, r.db.Model(&model.Comment{}).Count(&count).Error
}

func (r *CommentRepository) CountByDate(date string) (int64, error) {
	var count int64
	return count, r.db.Model(&model.Comment{}).Where("DATE(created_at) = ?", date).Count(&count).Error
}

func (r *CommentRepository) Transaction(fn func(*gorm.DB) error) error {
	return r.db.Transaction(fn)
}

func (r *CommentRepository) IncrementLikes(id uint, delta int) error {
	return r.db.Model(&model.Comment{}).Where("id = ?", id).UpdateColumn("likes_count", gorm.Expr("likes_count + ?", delta)).Error
}

type LikeRepository struct {
	db *gorm.DB
}

func NewLikeRepository(db *gorm.DB) *LikeRepository {
	return &LikeRepository{db: db}
}

func (r *LikeRepository) Create(like *model.Like) error {
	return r.db.Create(like).Error
}

func (r *LikeRepository) Delete(userID uint, targetType string, targetID uint) error {
	return r.db.Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).Delete(&model.Like{}).Error
}

func (r *LikeRepository) Exists(userID uint, targetType string, targetID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&model.Like{}).Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *LikeRepository) GetUserLikeMap(userID uint, targetType string, targetIDs []uint) (map[uint]bool, error) {
	var likes []model.Like
	result := make(map[uint]bool)
	if err := r.db.Where("user_id = ? AND target_type = ? AND target_id IN ?", userID, targetType, targetIDs).Find(&likes).Error; err != nil {
		return nil, err
	}
	for _, like := range likes {
		result[like.TargetID] = true
	}
	return result, nil
}

// ListPostLikesByUser 用户点赞过的动态 ID（按点赞时间倒序）
func (r *LikeRepository) ListPostLikesByUser(userID uint, page, pageSize int) ([]model.Like, int64, error) {
	var likes []model.Like
	var total int64
	query := r.db.Model(&model.Like{}).Where("user_id = ? AND target_type = ?", userID, "post")
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if err := query.Order("id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&likes).Error; err != nil {
		return nil, 0, err
	}
	return likes, total, nil
}

func (r *LikeRepository) ListByTarget(targetType string, targetID uint, page, pageSize int) ([]model.Like, int64, error) {
	var likes []model.Like
	var total int64
	query := r.db.Model(&model.Like{}).Where("target_type = ? AND target_id = ?", targetType, targetID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("id DESC").Find(&likes).Error; err != nil {
		return nil, 0, err
	}

	return likes, total, nil
}

func (r *LikeRepository) CountAll() (int64, error) {
	var count int64
	return count, r.db.Model(&model.Like{}).Count(&count).Error
}

func (r *LikeRepository) CountByDate(date string) (int64, error) {
	var count int64
	return count, r.db.Model(&model.Like{}).Where("DATE(created_at) = ?", date).Count(&count).Error
}

type PostStatusCount struct {
	Status int
	Count  int64
}

func (r *PostRepository) GetStatusDistribution() ([]PostStatusCount, error) {
	var results []PostStatusCount
	err := r.db.Model(&model.Post{}).
		Select("status, COUNT(*) as count").Group("status").Scan(&results).Error
	return results, err
}

type FavoriteRepository struct {
	db *gorm.DB
}

func NewFavoriteRepository(db *gorm.DB) *FavoriteRepository {
	return &FavoriteRepository{db: db}
}

func (r *FavoriteRepository) Create(favorite *model.Favorite) error {
	return r.db.Create(favorite).Error
}

func (r *FavoriteRepository) Delete(userID, postID uint) error {
	return r.db.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&model.Favorite{}).Error
}

func (r *FavoriteRepository) Exists(userID, postID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&model.Favorite{}).Where("user_id = ? AND post_id = ?", userID, postID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *FavoriteRepository) ListByUser(userID uint, page, pageSize int) ([]model.Favorite, int64, error) {
	var favorites []model.Favorite
	var total int64
	query := r.db.Model(&model.Favorite{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("id DESC").Find(&favorites).Error; err != nil {
		return nil, 0, err
	}

	return favorites, total, nil
}

func (r *FavoriteRepository) GetUserFavoriteMap(userID uint, postIDs []uint) (map[uint]bool, error) {
	var favorites []model.Favorite
	result := make(map[uint]bool)
	if err := r.db.Where("user_id = ? AND post_id IN ?", userID, postIDs).Find(&favorites).Error; err != nil {
		return nil, err
	}
	for _, fav := range favorites {
		result[fav.PostID] = true
	}
	return result, nil
}

type DraftRepository struct {
	db *gorm.DB
}

func NewDraftRepository(db *gorm.DB) *DraftRepository {
	return &DraftRepository{db: db}
}

func (r *DraftRepository) Create(draft *model.Draft) error {
	return r.db.Create(draft).Error
}

func (r *DraftRepository) Update(draft *model.Draft) error {
	return r.db.Save(draft).Error
}

func (r *DraftRepository) GetByID(id uint) (*model.Draft, error) {
	var draft model.Draft
	if err := r.db.First(&draft, id).Error; err != nil {
		return nil, err
	}
	return &draft, nil
}

func (r *DraftRepository) Delete(id uint) error {
	return r.db.Delete(&model.Draft{}, id).Error
}

func (r *DraftRepository) ListByUser(userID uint, page, pageSize int) ([]model.Draft, int64, error) {
	var drafts []model.Draft
	var total int64
	query := r.db.Model(&model.Draft{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("updated_at DESC").Find(&drafts).Error; err != nil {
		return nil, 0, err
	}

	return drafts, total, nil
}

type ReportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

func (r *ReportRepository) Create(report *model.Report) error {
	return r.db.Create(report).Error
}

func (r *ReportRepository) GetByID(id uint) (*model.Report, error) {
	var report model.Report
	if err := r.db.First(&report, id).Error; err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *ReportRepository) ListByUser(reporterID uint, page, pageSize int) ([]model.Report, int64, error) {
	var reports []model.Report
	var total int64
	query := r.db.Model(&model.Report{}).Where("reporter_id = ?", reporterID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("id DESC").Find(&reports).Error; err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

func (r *ReportRepository) Exists(reporterID uint, targetType string, targetID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&model.Report{}).Where("reporter_id = ? AND target_type = ? AND target_id = ?", reporterID, targetType, targetID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListAdmin 管理后台举报列表（按状态/原因类型/目标类型筛选）
func (r *ReportRepository) ListAdmin(status *int, reasonType, targetType string, page, pageSize int) ([]model.Report, int64, error) {
	var reports []model.Report
	var total int64
	query := r.db.Model(&model.Report{})

	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if reasonType != "" {
		query = query.Where("reason_type = ?", reasonType)
	}
	if targetType != "" {
		query = query.Where("target_type = ?", targetType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("id DESC").Find(&reports).Error; err != nil {
		return nil, 0, err
	}
	return reports, total, nil
}

// UpdateHandle 更新举报处理结果
func (r *ReportRepository) UpdateHandle(id uint, status int, result string, handledBy uint, handledAt time.Time) error {
	return r.db.Model(&model.Report{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        status,
		"handle_result": result,
		"handled_by":    handledBy,
		"handled_at":    handledAt,
	}).Error
}

// CountByStatus 按状态统计举报数
func (r *ReportRepository) CountByStatus(status int) (int64, error) {
	var count int64
	err := r.db.Model(&model.Report{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

// CountAll 举报总数
func (r *ReportRepository) CountAll() (int64, error) {
	var count int64
	err := r.db.Model(&model.Report{}).Count(&count).Error
	return count, err
}

// CountGroupByReasonType 按举报原因类型分组统计
func (r *ReportRepository) CountGroupByReasonType() ([]struct {
	ReasonType string
	Count      int64
}, error) {
	var rows []struct {
		ReasonType string
		Count      int64
	}
	err := r.db.Model(&model.Report{}).
		Select("reason_type as reason_type, COUNT(*) as count").
		Group("reason_type").
		Scan(&rows).Error
	return rows, err
}

// CountGroupByTargetType 按目标类型分组统计
func (r *ReportRepository) CountGroupByTargetType() ([]struct {
	TargetType string
	Count      int64
}, error) {
	var rows []struct {
		TargetType string
		Count      int64
	}
	err := r.db.Model(&model.Report{}).
		Select("target_type as target_type, COUNT(*) as count").
		Group("target_type").
		Scan(&rows).Error
	return rows, err
}

type TopicRepository struct {
	db *gorm.DB
}

func NewTopicRepository(db *gorm.DB) *TopicRepository {
	return &TopicRepository{db: db}
}

func (r *TopicRepository) Create(topic *model.Topic) error {
	// 显式 Select，避免 Status=0（待审核）被 GORM 零值省略后套用 default
	return r.db.Select("Name", "Description", "Cover", "CreatorID", "PostsCount", "Status").Create(topic).Error
}

func (r *TopicRepository) Update(topic *model.Topic) error {
	return r.db.Save(topic).Error
}

func (r *TopicRepository) GetByID(id uint) (*model.Topic, error) {
	var topic model.Topic
	if err := r.db.First(&topic, id).Error; err != nil {
		return nil, err
	}
	return &topic, nil
}

func (r *TopicRepository) GetByName(name string) (*model.Topic, error) {
	var topic model.Topic
	if err := r.db.Where("name = ?", name).First(&topic).Error; err != nil {
		return nil, err
	}
	return &topic, nil
}

func (r *TopicRepository) List(keyword string, page, pageSize int) ([]model.Topic, int64, error) {
	var topics []model.Topic
	var total int64
	query := r.db.Model(&model.Topic{}).Where("status = 1")

	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("posts_count DESC, id DESC").Find(&topics).Error; err != nil {
		return nil, 0, err
	}

	return topics, total, nil
}

func (r *TopicRepository) Search(keyword string, limit int) ([]model.Topic, error) {
	var topics []model.Topic
	if err := r.db.Where("status = 1 AND name LIKE ?", "%"+keyword+"%").Limit(limit).Order("posts_count DESC").Find(&topics).Error; err != nil {
		return nil, err
	}
	return topics, nil
}

func (r *TopicRepository) IncrementPostsCount(id uint, delta int) error {
	return r.db.Model(&model.Topic{}).Where("id = ?", id).UpdateColumn("posts_count", gorm.Expr("posts_count + ?", delta)).Error
}

func (r *TopicRepository) GetByIDs(ids []uint) ([]model.Topic, error) {
	var topics []model.Topic
	if err := r.db.Where("id IN ? AND status = 1", ids).Find(&topics).Error; err != nil {
		return nil, err
	}
	return topics, nil
}

func (r *TopicRepository) InitDefaultTopics() error {
	var count int64
	r.db.Model(&model.Topic{}).Count(&count)
	if count > 0 {
		return nil
	}

	topics := []model.Topic{
		{Name: "#萌宠日常", Description: "分享你家萌宠的日常生活", Status: 1},
		{Name: "#养宠知识", Description: "宠物养护、健康知识分享", Status: 1},
		{Name: "#宠物搞笑", Description: "宠物搞笑瞬间", Status: 1},
		{Name: "#宠物求助", Description: "宠物相关问题求助", Status: 1},
		{Name: "#领养送养", Description: "宠物领养、送养信息", Status: 1},
		{Name: "#宠物健康", Description: "宠物健康、医疗话题", Status: 1},
		{Name: "#晒宠大赛", Description: "晒出你家最可爱的宠物", Status: 1},
	}

	return r.db.Create(&topics).Error
}

type ShareRepository struct {
	db *gorm.DB
}

func NewShareRepository(db *gorm.DB) *ShareRepository {
	return &ShareRepository{db: db}
}

func (r *ShareRepository) Create(share *model.ShareRecord) error {
	return r.db.Create(share).Error
}
