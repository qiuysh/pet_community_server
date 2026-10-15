package types

import (
	"pet-community/model"
	"strings"
	"time"
)

type SearchResult struct {
	Users  []UserSearchResult  `json:"users"`
	Posts  []PostSearchResult  `json:"posts"`
	Topics []TopicSearchResult `json:"topics"`
}

type UserSearchResult struct {
	ID             uint   `json:"id"`
	Nickname       string `json:"nickname"`
	Avatar         string `json:"avatar"`
	FollowersCount int    `json:"followersCount"`
}

type PostSearchResult struct {
	ID            uint     `json:"id"`
	UserID        uint     `json:"userId"`
	UserNickname  string   `json:"userNickname"`
	UserAvatar    string   `json:"userAvatar"`
	Content       string   `json:"content"`
	Images        []string `json:"images,omitempty"`
	Video         string   `json:"video,omitempty"`
	LikesCount    int      `json:"likesCount"`
	CommentsCount int      `json:"commentsCount"`
	CreatedAt     string   `json:"createdAt"`
}

type TopicSearchResult struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Cover       string `json:"cover"`
	PostsCount  int    `json:"postsCount"`
}

type SearchHistoryInfo struct {
	ID        uint   `json:"id"`
	UserID    uint   `json:"userId"`
	Keyword   string `json:"keyword"`
	CreatedAt string `json:"createdAt"`
}

type HotKeywordInfo struct {
	ID          uint   `json:"id"`
	Keyword     string `json:"keyword"`
	SearchCount int    `json:"searchCount"`
	Rank        int    `json:"rank"`
}

type TopicCategoryInfo struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	SortOrder   int    `json:"sortOrder"`
	Status      int    `json:"status"`
}

type TopicDetailInfo struct {
	ID                uint   `json:"id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	Cover             string `json:"cover"`
	PostsCount        int    `json:"postsCount"`
	Status            int    `json:"status"`
	ParticipantsCount int    `json:"participantsCount"`
	IsParticipated    bool   `json:"isParticipated"`
	CreatedAt         string `json:"createdAt"`
}

type RecommendPostInfo struct {
	ID            uint     `json:"id"`
	UserID        uint     `json:"userId"`
	UserNickname  string   `json:"userNickname"`
	UserAvatar    string   `json:"userAvatar"`
	UserLevel     int      `json:"userLevel"`
	IsVipName     bool     `json:"isVipName"`
	Content       string   `json:"content"`
	Images        []string `json:"images,omitempty"`
	Video         string   `json:"video,omitempty"`
	LikesCount    int      `json:"likesCount"`
	CommentsCount int      `json:"commentsCount"`
	SharesCount   int      `json:"sharesCount"`
	IsPinned      int      `json:"isPinned"`
	IsFeatured    int      `json:"isFeatured"`
	IsLiked       bool     `json:"isLiked"`
	IsFavorited   bool     `json:"isFavorited"`
	CreatedAt     string   `json:"createdAt"`
}

type RecommendResponse struct {
	List     []RecommendPostInfo `json:"list"`
	Total    int                 `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"pageSize"`
}

type RankingPost struct {
	ID            uint   `json:"id"`
	UserID        uint   `json:"userId"`
	UserNickname  string `json:"userNickname"`
	UserAvatar    string `json:"userAvatar"`
	Content       string `json:"content"`
	LikesCount    int    `json:"likesCount"`
	CommentsCount int    `json:"commentsCount"`
	SharesCount   int    `json:"sharesCount"`
	HotScore      int    `json:"hotScore"`
	CreatedAt     string `json:"createdAt"`
}

type RankingUser struct {
	ID             uint   `json:"id"`
	Nickname       string `json:"nickname"`
	Avatar         string `json:"avatar"`
	FollowersCount int    `json:"followersCount"`
	PostsCount     int    `json:"postsCount"`
	HotScore       int    `json:"hotScore"`
}

type RankingTopic struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Cover       string `json:"cover"`
	PostsCount  int    `json:"postsCount"`
	HotScore    int    `json:"hotScore"`
}

type RankingResponse struct {
	Posts  []RankingPost  `json:"posts"`
	Users  []RankingUser  `json:"users"`
	Topics []RankingTopic `json:"topics"`
}

type CategoryPostResponse struct {
	List     []RecommendPostInfo `json:"list"`
	Total    int                 `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"pageSize"`
}

type TagCloudItem struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
	Hot   int    `json:"hot"`
}

type TagCloudResponse struct {
	Tags []TagCloudItem `json:"tags"`
}

func ConvertUserToSearchResult(user *model.User) UserSearchResult {
	var followersCount int
	return UserSearchResult{
		ID:             user.ID,
		Nickname:       user.Nickname,
		Avatar:         user.Avatar,
		FollowersCount: followersCount,
	}
}

func ConvertPostToSearchResult(post *model.Post, user *model.User) PostSearchResult {
	images := []string{}
	if post.Images != "" {
		for _, img := range splitString(post.Images) {
			images = append(images, img)
		}
	}
	return PostSearchResult{
		ID:            post.ID,
		UserID:        post.UserID,
		UserNickname:  user.Nickname,
		UserAvatar:    user.Avatar,
		Content:       post.Content,
		Images:        images,
		Video:         post.Video,
		LikesCount:    post.LikesCount,
		CommentsCount: post.CommentsCount,
		CreatedAt:     post.CreatedAt.Format(time.RFC3339),
	}
}

func ConvertTopicToSearchResult(topic *model.Topic) TopicSearchResult {
	return TopicSearchResult{
		ID:          topic.ID,
		Name:        topic.Name,
		Description: topic.Description,
		Cover:       topic.Cover,
		PostsCount:  topic.PostsCount,
	}
}

func ConvertSearchHistoryToInfo(history *model.SearchHistory) SearchHistoryInfo {
	return SearchHistoryInfo{
		ID:        history.ID,
		UserID:    history.UserID,
		Keyword:   history.Keyword,
		CreatedAt: history.CreatedAt.Format(time.RFC3339),
	}
}

func ConvertHotKeywordToInfo(keyword *model.HotKeyword) HotKeywordInfo {
	return HotKeywordInfo{
		ID:          keyword.ID,
		Keyword:     keyword.Keyword,
		SearchCount: keyword.SearchCount,
		Rank:        keyword.Rank,
	}
}

func ConvertTopicCategoryToInfo(cat *model.TopicCategory) TopicCategoryInfo {
	return TopicCategoryInfo{
		ID:          cat.ID,
		Name:        cat.Name,
		Description: cat.Description,
		Icon:        cat.Icon,
		SortOrder:   cat.SortOrder,
		Status:      cat.Status,
	}
}

func splitString(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ",")
}
