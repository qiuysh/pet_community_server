package service

import (
	"errors"

	"pet-community/repository"
)

// resolveTopicCategoryID 校验话题分类写入口径（R7）：
// - 0 / 未传 → 未分类
// - >0 → 分类须存在且启用，否则返回业务错误
func resolveTopicCategoryID(repo *repository.TopicCategoryRepository, categoryID uint) (uint, error) {
	if categoryID == 0 {
		return 0, nil
	}
	if repo == nil {
		return 0, errors.New("话题分类服务未配置")
	}
	cat, err := repo.GetByID(categoryID)
	if err != nil {
		return 0, errors.New("话题分类不存在")
	}
	if cat.Status != 1 {
		return 0, errors.New("话题分类未启用")
	}
	return cat.ID, nil
}
