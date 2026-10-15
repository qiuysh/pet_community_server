package cache

import (
	"fmt"
	"time"

	"pet-community/types"
)

const guestFeedTTL = 3 * time.Minute

func keyGuestFeed(page, pageSize int, postType string) string {
	if postType == "" {
		postType = "all"
	}
	return fmt.Sprintf("feed:hot:p%d:s%d:t%s", page, pageSize, postType)
}

// GetGuestFeed 游客推荐流缓存
func GetGuestFeed(page, pageSize int, postType string) (*types.RecommendResponse, bool) {
	var resp types.RecommendResponse
	if !GetJSON(keyGuestFeed(page, pageSize, postType), &resp) {
		return nil, false
	}
	return &resp, true
}

// SetGuestFeed 写入游客推荐流
func SetGuestFeed(page, pageSize int, postType string, resp *types.RecommendResponse) {
	if resp == nil {
		return
	}
	SetJSON(keyGuestFeed(page, pageSize, postType), resp, guestFeedTTL)
}

// InvalidateGuestFeed 新内容发布后失效游客 Feed
func InvalidateGuestFeed() {
	DeleteByPattern("feed:hot:*")
}
