package repository

import (
	"pet-community/model"

	"gorm.io/gorm"
)

// AdoptRepository 领养信息仓储
type AdoptRepository struct {
	db *gorm.DB
}

func NewAdoptRepository(db *gorm.DB) *AdoptRepository {
	return &AdoptRepository{db: db}
}

func (r *AdoptRepository) Create(adopt *model.Adopt) error {
	return r.db.Create(adopt).Error
}

func (r *AdoptRepository) Update(adopt *model.Adopt) error {
	return r.db.Save(adopt).Error
}

func (r *AdoptRepository) GetByID(id uint) (*model.Adopt, error) {
	var adopt model.Adopt
	if err := r.db.First(&adopt, id).Error; err != nil {
		return nil, err
	}
	return &adopt, nil
}

func (r *AdoptRepository) Delete(id uint) error {
	return r.db.Delete(&model.Adopt{}, id).Error
}

func (r *AdoptRepository) List(petType string, status int, location, keyword string, page, pageSize int) ([]model.Adopt, int64, error) {
	var adopts []model.Adopt
	var total int64
	query := r.db.Model(&model.Adopt{})

	if petType != "" {
		query = query.Where("pet_type = ?", petType)
	}

	if status >= 0 {
		query = query.Where("status = ?", status)
	} else {
		query = query.Where("status != ?", 2) // 默认不显示已取消的
	}

	if location != "" {
		query = query.Where("location LIKE ?", "%"+location+"%")
	}

	if keyword != "" {
		query = query.Where("pet_name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("id DESC").Find(&adopts).Error; err != nil {
		return nil, 0, err
	}

	return adopts, total, nil
}

func (r *AdoptRepository) ListByUser(userID uint, page, pageSize int) ([]model.Adopt, int64, error) {
	var adopts []model.Adopt
	var total int64
	query := r.db.Model(&model.Adopt{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("id DESC").Find(&adopts).Error; err != nil {
		return nil, 0, err
	}

	return adopts, total, nil
}

func (r *AdoptRepository) CountApplications(adoptionID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.AdoptionApplication{}).Where("adoption_id = ?", adoptionID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// Count 统计领养信息总数
func (r *AdoptRepository) Count() (int64, error) {
	var count int64
	if err := r.db.Model(&model.Adopt{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CountByStatus 按状态统计领养信息数量
func (r *AdoptRepository) CountByStatus(status int) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Adopt{}).Where("status = ?", status).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// AdoptionApplicationRepository 领养申请仓储
type AdoptionApplicationRepository struct {
	db *gorm.DB
}

func NewAdoptionApplicationRepository(db *gorm.DB) *AdoptionApplicationRepository {
	return &AdoptionApplicationRepository{db: db}
}

func (r *AdoptionApplicationRepository) Create(application *model.AdoptionApplication) error {
	return r.db.Create(application).Error
}

func (r *AdoptionApplicationRepository) Update(application *model.AdoptionApplication) error {
	return r.db.Save(application).Error
}

func (r *AdoptionApplicationRepository) GetByID(id uint) (*model.AdoptionApplication, error) {
	var application model.AdoptionApplication
	if err := r.db.First(&application, id).Error; err != nil {
		return nil, err
	}
	return &application, nil
}

func (r *AdoptionApplicationRepository) GetByAdoptionAndUser(adoptionID, userID uint) (*model.AdoptionApplication, error) {
	var application model.AdoptionApplication
	if err := r.db.Where("adoption_id = ? AND user_id = ?", adoptionID, userID).First(&application).Error; err != nil {
		return nil, err
	}
	return &application, nil
}

func (r *AdoptionApplicationRepository) ListByAdoption(adoptionID uint, status int, page, pageSize int) ([]model.AdoptionApplication, int64, error) {
	var applications []model.AdoptionApplication
	var total int64
	query := r.db.Model(&model.AdoptionApplication{}).Where("adoption_id = ?", adoptionID)

	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("id DESC").Find(&applications).Error; err != nil {
		return nil, 0, err
	}

	return applications, total, nil
}

func (r *AdoptionApplicationRepository) ListByUser(userID uint, status int, page, pageSize int) ([]model.AdoptionApplication, int64, error) {
	var applications []model.AdoptionApplication
	var total int64
	query := r.db.Model(&model.AdoptionApplication{}).Where("user_id = ?", userID)

	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("id DESC").Find(&applications).Error; err != nil {
		return nil, 0, err
	}

	return applications, total, nil
}

func (r *AdoptionApplicationRepository) Exists(adoptionID, userID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&model.AdoptionApplication{}).Where("adoption_id = ? AND user_id = ?", adoptionID, userID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *AdoptionApplicationRepository) UpdateStatus(id uint, status int, reviewedBy uint) error {
	return r.db.Model(&model.AdoptionApplication{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      status,
		"reviewed_by": reviewedBy,
	}).Error
}

func (r *AdoptionApplicationRepository) SignAgreement(id uint) error {
	return r.db.Model(&model.AdoptionApplication{}).Where("id = ?", id).Update("agreement_signed", 1).Error
}

func (r *AdoptionApplicationRepository) CountByAdoption(adoptionID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.AdoptionApplication{}).Where("adoption_id = ?", adoptionID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *AdoptionApplicationRepository) CountByAdoptionAndStatus(adoptionID uint, status int) (int64, error) {
	var count int64
	if err := r.db.Model(&model.AdoptionApplication{}).Where("adoption_id = ? AND status = ?", adoptionID, status).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// Count 统计领养申请总数
func (r *AdoptionApplicationRepository) Count() (int64, error) {
	var count int64
	if err := r.db.Model(&model.AdoptionApplication{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// AdoptionFollowUpRepository 领养回访记录仓储
type AdoptionFollowUpRepository struct {
	db *gorm.DB
}

func NewAdoptionFollowUpRepository(db *gorm.DB) *AdoptionFollowUpRepository {
	return &AdoptionFollowUpRepository{db: db}
}

func (r *AdoptionFollowUpRepository) Create(followUp *model.AdoptionFollowUp) error {
	return r.db.Create(followUp).Error
}

func (r *AdoptionFollowUpRepository) GetByID(id uint) (*model.AdoptionFollowUp, error) {
	var followUp model.AdoptionFollowUp
	if err := r.db.First(&followUp, id).Error; err != nil {
		return nil, err
	}
	return &followUp, nil
}

func (r *AdoptionFollowUpRepository) ListByAdoption(adoptionID uint, page, pageSize int) ([]model.AdoptionFollowUp, int64, error) {
	var followUps []model.AdoptionFollowUp
	var total int64
	query := r.db.Model(&model.AdoptionFollowUp{}).Where("adoption_id = ?", adoptionID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("id DESC").Find(&followUps).Error; err != nil {
		return nil, 0, err
	}

	return followUps, total, nil
}

func (r *AdoptionFollowUpRepository) ListByApplication(applicationID uint, page, pageSize int) ([]model.AdoptionFollowUp, int64, error) {
	var followUps []model.AdoptionFollowUp
	var total int64
	query := r.db.Model(&model.AdoptionFollowUp{}).Where("application_id = ?", applicationID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("id DESC").Find(&followUps).Error; err != nil {
		return nil, 0, err
	}

	return followUps, total, nil
}


