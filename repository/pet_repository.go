package repository

import (
	"pet-community/model"

	"gorm.io/gorm"
)

type PetRepository struct {
	db *gorm.DB
}

func NewPetRepository(db *gorm.DB) *PetRepository {
	return &PetRepository{db: db}
}

func (r *PetRepository) Create(pet *model.Pet) error {
	return r.db.Create(pet).Error
}

func (r *PetRepository) Update(pet *model.Pet) error {
	return r.db.Save(pet).Error
}

func (r *PetRepository) GetByID(id uint) (*model.Pet, error) {
	var pet model.Pet
	if err := r.db.First(&pet, id).Error; err != nil {
		return nil, err
	}
	return &pet, nil
}

func (r *PetRepository) Delete(id uint) error {
	return r.db.Delete(&model.Pet{}, id).Error
}

func (r *PetRepository) List(userID uint, petType int, keyword string, page, pageSize int) ([]model.Pet, int64, error) {
	var pets []model.Pet
	var total int64
	query := r.db.Model(&model.Pet{}).Where("status = 1")

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if petType > 0 {
		query = query.Where("pet_type = ?", petType)
	}
	if keyword != "" {
		query = query.Where("name LIKE ? OR breed_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("id DESC").Find(&pets).Error; err != nil {
		return nil, 0, err
	}

	return pets, total, nil
}

// CountByUser 统计用户有效宠物数
func (r *PetRepository) CountByUser(userID uint) (int64, error) {
	var count int64
	return count, r.db.Model(&model.Pet{}).Where("user_id = ? AND status = 1", userID).Count(&count).Error
}

func (r *PetRepository) GetUserPets(userID uint) ([]model.Pet, error) {
	var pets []model.Pet
	if err := r.db.Where("user_id = ? AND status = 1", userID).Order("id DESC").Find(&pets).Error; err != nil {
		return nil, err
	}
	return pets, nil
}

func (r *PetRepository) IncrementViewCount(id uint) error {
	return r.db.Model(&model.Pet{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *PetRepository) ListBreeds(petType int) ([]model.PetBreed, error) {
	var breeds []model.PetBreed
	query := r.db.Model(&model.PetBreed{}).Where("status = 1")
	if petType > 0 {
		query = query.Where("pet_type = ?", petType)
	}
	if err := query.Order("sort_order DESC, id ASC").Find(&breeds).Error; err != nil {
		return nil, err
	}
	return breeds, nil
}

func (r *PetRepository) GetBreedByID(id uint) (*model.PetBreed, error) {
	var breed model.PetBreed
	if err := r.db.First(&breed, id).Error; err != nil {
		return nil, err
	}
	return &breed, nil
}

func (r *PetRepository) AddAlbumPhoto(photo *model.PetAlbum) error {
	return r.db.Create(photo).Error
}

func (r *PetRepository) GetPetAlbum(petID uint) ([]model.PetAlbum, error) {
	var photos []model.PetAlbum
	if err := r.db.Where("pet_id = ?", petID).Order("sort_order DESC, id DESC").Find(&photos).Error; err != nil {
		return nil, err
	}
	return photos, nil
}

func (r *PetRepository) GetAlbumPhotoByID(id, petID uint) (*model.PetAlbum, error) {
	var photo model.PetAlbum
	if err := r.db.Where("id = ? AND pet_id = ?", id, petID).First(&photo).Error; err != nil {
		return nil, err
	}
	return &photo, nil
}

func (r *PetRepository) MaxAlbumSortOrder(petID uint) (int, error) {
	var maxOrder *int
	err := r.db.Model(&model.PetAlbum{}).Where("pet_id = ?", petID).
		Select("MAX(sort_order)").Scan(&maxOrder).Error
	if err != nil {
		return 0, err
	}
	if maxOrder == nil {
		return 0, nil
	}
	return *maxOrder, nil
}

// UpdateAlbumSortOrders 按 ids 顺序写入 sort_order（首位最高）
func (r *PetRepository) UpdateAlbumSortOrders(petID uint, ids []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		n := len(ids)
		for i, id := range ids {
			sortOrder := n - i
			res := tx.Model(&model.PetAlbum{}).
				Where("id = ? AND pet_id = ?", id, petID).
				Update("sort_order", sortOrder)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return gorm.ErrRecordNotFound
			}
		}
		return nil
	})
}

func (r *PetRepository) DeleteAlbumPhoto(id, petID uint) error {
	return r.db.Where("id = ? AND pet_id = ?", id, petID).Delete(&model.PetAlbum{}).Error
}

func (r *PetRepository) CreateGrowthRecord(record *model.PetGrowthRecord) error {
	return r.db.Create(record).Error
}

func (r *PetRepository) UpdateGrowthRecord(record *model.PetGrowthRecord) error {
	return r.db.Save(record).Error
}

func (r *PetRepository) GetGrowthRecordByID(id uint) (*model.PetGrowthRecord, error) {
	var record model.PetGrowthRecord
	if err := r.db.First(&record, id).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *PetRepository) DeleteGrowthRecord(id, petID uint) error {
	return r.db.Where("id = ? AND pet_id = ?", id, petID).Delete(&model.PetGrowthRecord{}).Error
}

func (r *PetRepository) GetGrowthRecords(petID uint, page, pageSize int) ([]model.PetGrowthRecord, int64, error) {
	var records []model.PetGrowthRecord
	var total int64
	query := r.db.Model(&model.PetGrowthRecord{}).Where("pet_id = ?", petID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("record_date DESC, id DESC").Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

func (r *PetRepository) CreateLostPet(lostPet *model.LostPet) error {
	return r.db.Create(lostPet).Error
}

func (r *PetRepository) UpdateLostPet(lostPet *model.LostPet) error {
	return r.db.Save(lostPet).Error
}

func (r *PetRepository) GetLostPetByID(id uint) (*model.LostPet, error) {
	var lostPet model.LostPet
	if err := r.db.First(&lostPet, id).Error; err != nil {
		return nil, err
	}
	return &lostPet, nil
}

func (r *PetRepository) ListLostPets(petType int, lostLocation string, status int, page, pageSize int) ([]model.LostPet, int64, error) {
	var lostPets []model.LostPet
	var total int64
	query := r.db.Model(&model.LostPet{})

	if petType > 0 {
		query = query.Where("pet_type = ?", petType)
	}
	if lostLocation != "" {
		query = query.Where("lost_location LIKE ?", "%"+lostLocation+"%")
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("id DESC").Find(&lostPets).Error; err != nil {
		return nil, 0, err
	}

	return lostPets, total, nil
}

func (r *PetRepository) IncrementLostPetViewCount(id uint) error {
	return r.db.Model(&model.LostPet{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *PetRepository) InitBreeds() error {
	var count int64
	r.db.Model(&model.PetBreed{}).Count(&count)
	if count > 0 {
		return nil
	}

	breeds := []model.PetBreed{
		{PetType: 1, Name: "中华田园犬", EnglishName: "Chinese Rural Dog", SortOrder: 100},
		{PetType: 1, Name: "金毛寻回犬", EnglishName: "Golden Retriever", SortOrder: 99},
		{PetType: 1, Name: "拉布拉多", EnglishName: "Labrador Retriever", SortOrder: 98},
		{PetType: 1, Name: "哈士奇", EnglishName: "Siberian Husky", SortOrder: 97},
		{PetType: 1, Name: "柯基", EnglishName: "Welsh Corgi", SortOrder: 96},
		{PetType: 1, Name: "泰迪/贵宾", EnglishName: "Poodle", SortOrder: 95},
		{PetType: 1, Name: "柴犬", EnglishName: "Shiba Inu", SortOrder: 94},
		{PetType: 1, Name: "萨摩耶", EnglishName: "Samoyed", SortOrder: 93},
		{PetType: 1, Name: "边牧", EnglishName: "Border Collie", SortOrder: 92},
		{PetType: 1, Name: "阿拉斯加", EnglishName: "Alaskan Malamute", SortOrder: 91},

		{PetType: 2, Name: "中华田园猫", EnglishName: "Chinese Rural Cat", SortOrder: 100},
		{PetType: 2, Name: "英国短毛猫", EnglishName: "British Shorthair", SortOrder: 99},
		{PetType: 2, Name: "美国短毛猫", EnglishName: "American Shorthair", SortOrder: 98},
		{PetType: 2, Name: "布偶猫", EnglishName: "Ragdoll", SortOrder: 97},
		{PetType: 2, Name: "暹罗猫", EnglishName: "Siamese", SortOrder: 96},
		{PetType: 2, Name: "波斯猫", EnglishName: "Persian", SortOrder: 95},
		{PetType: 2, Name: "缅因猫", EnglishName: "Maine Coon", SortOrder: 94},
		{PetType: 2, Name: "橘猫", EnglishName: "Orange Tabby", SortOrder: 93},
		{PetType: 2, Name: "狸花猫", EnglishName: "Dragon Li", SortOrder: 92},
		{PetType: 2, Name: "加菲猫", EnglishName: "Exotic Shorthair", SortOrder: 91},

		{PetType: 3, Name: "仓鼠", EnglishName: "Hamster", SortOrder: 100},
		{PetType: 3, Name: "兔子", EnglishName: "Rabbit", SortOrder: 99},
		{PetType: 3, Name: "龙猫", EnglishName: "Chinchilla", SortOrder: 98},
		{PetType: 3, Name: "荷兰猪", EnglishName: "Guinea Pig", SortOrder: 97},
		{PetType: 3, Name: "刺猬", EnglishName: "Hedgehog", SortOrder: 96},

		{PetType: 4, Name: "鹦鹉", EnglishName: "Parrot", SortOrder: 100},
		{PetType: 4, Name: "虎皮鹦鹉", EnglishName: "Budgerigar", SortOrder: 99},
		{PetType: 4, Name: "玄凤鹦鹉", EnglishName: "Cockatiel", SortOrder: 98},

		{PetType: 5, Name: "金鱼", EnglishName: "Goldfish", SortOrder: 100},
		{PetType: 5, Name: "锦鲤", EnglishName: "Koi", SortOrder: 99},
		{PetType: 5, Name: "乌龟", EnglishName: "Turtle", SortOrder: 98},
	}

	return r.db.Create(&breeds).Error
}
