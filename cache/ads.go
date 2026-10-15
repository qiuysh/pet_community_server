package cache

import (
	"time"

	"pet-community/model"
)

// GetActiveAds 读取有效广告列表缓存
func GetActiveAds(location string) ([]model.Advertisement, bool) {
	var ads []model.Advertisement
	if !GetJSON(keyActiveAds(location), &ads) {
		return nil, false
	}
	return ads, true
}

// SetActiveAds 写入有效广告列表缓存
func SetActiveAds(location string, ads []model.Advertisement) {
	if ads == nil {
		ads = []model.Advertisement{}
	}
	SetJSON(keyActiveAds(location), ads, time.Duration(activeAdsTTLSeconds)*time.Second)
}

// InvalidateActiveAds 广告变更后清空所有位置缓存
func InvalidateActiveAds() {
	DeleteByPattern(keyActiveAdsPattern)
}
