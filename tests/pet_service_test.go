package tests

import (
	"os"
	"strings"
	"testing"
	"time"

	"pet-community/model"
	"pet-community/repository"
	"pet-community/service"
	"pet-community/tests/testutil"
	"pet-community/types"
)

func TestPetServiceCreatePet(t *testing.T) {
	db := testutil.SetupTestDB(t)

	petRepo := repository.NewPetRepository(db)
	srv := service.NewPetService(petRepo)

	t.Run("success", func(t *testing.T) {
		req := &types.PetCreateRequest{
			Name:    "Buddy",
			PetType: 1,
			BreedID: 0,
			Gender:  1,
		}
		info, err := srv.CreatePet(1, req)
		if err != nil {
			t.Fatalf("create pet failed: %v", err)
		}
		if info.ID == 0 {
			t.Error("pet ID should not be zero")
		}
		if info.Name != "Buddy" {
			t.Errorf("expected name Buddy, got %s", info.Name)
		}
	})

	t.Run("with breed auto fill", func(t *testing.T) {
		breed := &model.PetBreed{PetType: 1, Name: "Golden Retriever", Status: 1}
		db.Create(breed)

		req := &types.PetCreateRequest{
			Name:    "Max",
			PetType: 1,
			BreedID: breed.ID,
			Gender:  1,
		}
		info, err := srv.CreatePet(1, req)
		if err != nil {
			t.Fatalf("create pet with breed failed: %v", err)
		}
		if info.BreedName != "Golden Retriever" {
			t.Errorf("expected breed name Golden Retriever, got %s", info.BreedName)
		}
	})
}

func TestPetServiceGetPetByID(t *testing.T) {
	db := testutil.SetupTestDB(t)

	petRepo := repository.NewPetRepository(db)
	srv := service.NewPetService(petRepo)

	pet := &model.Pet{UserID: 1, Name: "Buddy", PetType: 1, Status: 1}
	db.Create(pet)

	t.Run("success", func(t *testing.T) {
		info, err := srv.GetPetByID(pet.ID)
		if err != nil {
			t.Fatalf("get pet failed: %v", err)
		}
		if info.ID != pet.ID {
			t.Errorf("expected pet ID %d, got %d", pet.ID, info.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := srv.GetPetByID(9999)
		if err == nil {
			t.Error("expected error for non-existent pet")
		}
	})
}

func TestPetServiceUpdatePet(t *testing.T) {
	db := testutil.SetupTestDB(t)

	petRepo := repository.NewPetRepository(db)
	srv := service.NewPetService(petRepo)

	pet := &model.Pet{UserID: 1, Name: "Buddy", PetType: 1, Status: 1}
	db.Create(pet)

	t.Run("success", func(t *testing.T) {
		newName := "Buddy Updated"
		gender := 2
		sterilized := 1
		vaccinated := 1
		info, err := srv.UpdatePet(1, pet.ID, &types.PetUpdateRequest{
			Name:         newName,
			Gender:       gender,
			Weight:       12.5,
			Personality:  "friendly",
			Bio:          "lovely dog",
			IsSterilized: &sterilized,
			IsVaccinated: &vaccinated,
		})
		if err != nil {
			t.Fatalf("update pet failed: %v", err)
		}
		if info.Name != newName {
			t.Errorf("expected name %s, got %s", newName, info.Name)
		}
		if info.Gender != gender {
			t.Errorf("expected gender %d, got %d", gender, info.Gender)
		}
		if info.Weight != 12.5 {
			t.Errorf("expected weight 12.5, got %f", info.Weight)
		}
	})

	t.Run("no permission", func(t *testing.T) {
		_, err := srv.UpdatePet(2, pet.ID, &types.PetUpdateRequest{Name: "Hacker"})
		if err == nil {
			t.Error("expected error for no permission")
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := srv.UpdatePet(1, 9999, &types.PetUpdateRequest{Name: "Ghost"})
		if err == nil {
			t.Error("expected error for non-existent pet")
		}
	})
}

func TestPetServiceDeletePet(t *testing.T) {
	db := testutil.SetupTestDB(t)

	petRepo := repository.NewPetRepository(db)
	srv := service.NewPetService(petRepo)

	pet := &model.Pet{UserID: 1, Name: "Buddy", PetType: 1, Status: 1}
	db.Create(pet)

	t.Run("success", func(t *testing.T) {
		err := srv.DeletePet(1, pet.ID)
		if err != nil {
			t.Fatalf("delete pet failed: %v", err)
		}
		deleted, _ := petRepo.GetByID(pet.ID)
		if deleted.Status != 0 {
			t.Errorf("expected status 0, got %d", deleted.Status)
		}
	})

	t.Run("no permission", func(t *testing.T) {
		other := &model.Pet{UserID: 2, Name: "Other", PetType: 1, Status: 1}
		db.Create(other)
		err := srv.DeletePet(1, other.ID)
		if err == nil {
			t.Error("expected error for no permission")
		}
	})
}

func TestPetServiceGetUserPets(t *testing.T) {
	db := testutil.SetupTestDB(t)

	petRepo := repository.NewPetRepository(db)
	srv := service.NewPetService(petRepo)

	db.Create(&model.Pet{UserID: 1, Name: "Buddy", PetType: 1, Status: 1})
	db.Create(&model.Pet{UserID: 1, Name: "Max", PetType: 1, Status: 1})
	db.Create(&model.Pet{UserID: 2, Name: "Luna", PetType: 2, Status: 1})

	pets, err := srv.GetUserPets(1)
	if err != nil {
		t.Fatalf("get user pets failed: %v", err)
	}
	if len(pets) != 2 {
		t.Errorf("expected 2 pets, got %d", len(pets))
	}
}

func TestPetServiceListPets(t *testing.T) {
	db := testutil.SetupTestDB(t)

	petRepo := repository.NewPetRepository(db)
	srv := service.NewPetService(petRepo)

	db.Create(&model.Pet{UserID: 1, Name: "Buddy", PetType: 1, Status: 1})
	db.Create(&model.Pet{UserID: 1, Name: "Max", PetType: 1, Status: 1})
	db.Create(&model.Pet{UserID: 2, Name: "Luna", PetType: 2, Status: 1})

	t.Run("list all", func(t *testing.T) {
		resp, err := srv.ListPets(&types.PetListRequest{Page: 1, PageSize: 10})
		if err != nil {
			t.Fatalf("list pets failed: %v", err)
		}
		if resp.Total != 3 {
			t.Errorf("expected total 3, got %d", resp.Total)
		}
	})

	t.Run("filter by user", func(t *testing.T) {
		resp, err := srv.ListPets(&types.PetListRequest{UserID: 1, Page: 1, PageSize: 10})
		if err != nil {
			t.Fatalf("list pets failed: %v", err)
		}
		if resp.Total != 2 {
			t.Errorf("expected total 2, got %d", resp.Total)
		}
	})

	t.Run("filter by keyword", func(t *testing.T) {
		resp, err := srv.ListPets(&types.PetListRequest{Keyword: "Buddy", Page: 1, PageSize: 10})
		if err != nil {
			t.Fatalf("list pets failed: %v", err)
		}
		if resp.Total != 1 {
			t.Errorf("expected total 1, got %d", resp.Total)
		}
	})

	t.Run("filter by type", func(t *testing.T) {
		resp, err := srv.ListPets(&types.PetListRequest{PetType: 2, Page: 1, PageSize: 10})
		if err != nil {
			t.Fatalf("list pets failed: %v", err)
		}
		if resp.Total != 1 {
			t.Errorf("expected total 1, got %d", resp.Total)
		}
	})
}

func TestPetServiceGetBreeds(t *testing.T) {
	db := testutil.SetupTestDB(t)

	petRepo := repository.NewPetRepository(db)
	srv := service.NewPetService(petRepo)

	db.Create(&model.PetBreed{PetType: 1, Name: "Golden", Status: 1})
	db.Create(&model.PetBreed{PetType: 1, Name: "Labrador", Status: 1})
	db.Create(&model.PetBreed{PetType: 2, Name: "Persian", Status: 1})

	t.Run("filter by type", func(t *testing.T) {
		breeds, err := srv.GetBreeds(1)
		if err != nil {
			t.Fatalf("get breeds failed: %v", err)
		}
		if len(breeds) != 2 {
			t.Errorf("expected 2 breeds, got %d", len(breeds))
		}
	})

	t.Run("all types", func(t *testing.T) {
		breeds, err := srv.GetBreeds(0)
		if err != nil {
			t.Fatalf("get breeds failed: %v", err)
		}
		if len(breeds) != 3 {
			t.Errorf("expected 3 breeds, got %d", len(breeds))
		}
	})
}

func TestPetServiceUploadAlbumPhoto(t *testing.T) {
	db := testutil.SetupTestDB(t)
	srv := service.NewPetService(repository.NewPetRepository(db))

	t.Run("reject invalid ext", func(t *testing.T) {
		_, err := srv.UploadAlbumPhoto(strings.NewReader("fake"), "exe")
		if err == nil {
			t.Fatal("expected error for invalid extension")
		}
	})

	t.Run("accept png", func(t *testing.T) {
		url, err := srv.UploadAlbumPhoto(strings.NewReader("fakepngdata"), "png")
		if err != nil {
			t.Fatalf("upload failed: %v", err)
		}
		if !strings.HasPrefix(url, "/uploads/pet-album/") {
			t.Fatalf("unexpected url: %s", url)
		}
		_ = os.Remove("public" + url)
	})
}

func TestPetServiceUploadAlbumVideo(t *testing.T) {
	db := testutil.SetupTestDB(t)
	srv := service.NewPetService(repository.NewPetRepository(db))

	t.Run("reject non-mp4", func(t *testing.T) {
		_, err := srv.UploadAlbumVideo(strings.NewReader("fake"), "avi")
		if err == nil {
			t.Fatal("expected error for non-mp4")
		}
	})

	t.Run("accept mp4", func(t *testing.T) {
		url, err := srv.UploadAlbumVideo(strings.NewReader("fakevideodata"), "mp4")
		if err != nil {
			t.Fatalf("upload failed: %v", err)
		}
		if !strings.HasPrefix(url, "/uploads/pet-album/") {
			t.Fatalf("unexpected url: %s", url)
		}
		_ = os.Remove("public" + url)
	})
}

func TestPetServiceReorderAlbum(t *testing.T) {
	db := testutil.SetupTestDB(t)
	petRepo := repository.NewPetRepository(db)
	srv := service.NewPetService(petRepo)

	pet, err := srv.CreatePet(1, &types.PetCreateRequest{Name: "SortPet", PetType: 1})
	if err != nil {
		t.Fatalf("create pet failed: %v", err)
	}

	a, _ := srv.AddAlbumPhoto(1, pet.ID, &types.PetAlbumCreateRequest{ImageURL: "/a.jpg", Title: "A"})
	b, _ := srv.AddAlbumPhoto(1, pet.ID, &types.PetAlbumCreateRequest{ImageURL: "/b.jpg", Title: "B", MediaType: "image"})
	c, _ := srv.AddAlbumPhoto(1, pet.ID, &types.PetAlbumCreateRequest{ImageURL: "/c.mp4", Title: "C", MediaType: "video"})

	t.Run("reorder", func(t *testing.T) {
		if err := srv.ReorderAlbum(1, pet.ID, []uint{a.ID, c.ID, b.ID}); err != nil {
			t.Fatalf("reorder failed: %v", err)
		}
		list, err := srv.GetPetAlbum(pet.ID)
		if err != nil {
			t.Fatalf("get album failed: %v", err)
		}
		if len(list) != 3 || list[0].ID != a.ID || list[1].ID != c.ID || list[2].ID != b.ID {
			t.Fatalf("unexpected order: %+v", list)
		}
		if list[1].MediaType != "video" {
			t.Fatalf("expected video media type, got %s", list[1].MediaType)
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		err := srv.ReorderAlbum(2, pet.ID, []uint{a.ID, b.ID, c.ID})
		if err == nil {
			t.Fatal("expected permission error")
		}
	})
}

func TestPetServiceAlbum(t *testing.T) {
	db := testutil.SetupTestDB(t)

	petRepo := repository.NewPetRepository(db)
	srv := service.NewPetService(petRepo)

	pet := &model.Pet{UserID: 1, Name: "Buddy", PetType: 1, Status: 1}
	db.Create(pet)

	t.Run("add album photo", func(t *testing.T) {
		info, err := srv.AddAlbumPhoto(1, pet.ID, &types.PetAlbumCreateRequest{
			ImageURL: "http://example.com/1.jpg",
			Title:    "Cute",
		})
		if err != nil {
			t.Fatalf("add album photo failed: %v", err)
		}
		if info.ImageURL != "http://example.com/1.jpg" {
			t.Errorf("expected image url, got %s", info.ImageURL)
		}
	})

	t.Run("no permission", func(t *testing.T) {
		_, err := srv.AddAlbumPhoto(2, pet.ID, &types.PetAlbumCreateRequest{
			ImageURL: "http://example.com/2.jpg",
		})
		if err == nil {
			t.Error("expected error for no permission")
		}
	})

	t.Run("get album", func(t *testing.T) {
		photos, err := srv.GetPetAlbum(pet.ID)
		if err != nil {
			t.Fatalf("get pet album failed: %v", err)
		}
		if len(photos) != 1 {
			t.Errorf("expected 1 photo, got %d", len(photos))
		}
	})
}

func TestPetServiceDeleteAlbumPhoto(t *testing.T) {
	db := testutil.SetupTestDB(t)

	petRepo := repository.NewPetRepository(db)
	srv := service.NewPetService(petRepo)

	pet := &model.Pet{UserID: 1, Name: "Buddy", PetType: 1, Status: 1}
	db.Create(pet)

	photo := &model.PetAlbum{PetID: pet.ID, ImageURL: "http://example.com/1.jpg"}
	db.Create(photo)

	t.Run("success", func(t *testing.T) {
		err := srv.DeleteAlbumPhoto(1, pet.ID, photo.ID)
		if err != nil {
			t.Fatalf("delete album photo failed: %v", err)
		}
	})

	t.Run("no permission", func(t *testing.T) {
		err := srv.DeleteAlbumPhoto(2, pet.ID, photo.ID)
		if err == nil {
			t.Error("expected error for no permission")
		}
	})
}

func TestPetServiceGrowthRecord(t *testing.T) {
	db := testutil.SetupTestDB(t)

	petRepo := repository.NewPetRepository(db)
	srv := service.NewPetService(petRepo)

	pet := &model.Pet{UserID: 1, Name: "Buddy", PetType: 1, Status: 1}
	db.Create(pet)

	recordDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("create", func(t *testing.T) {
		info, err := srv.CreateGrowthRecord(1, pet.ID, &types.PetGrowthRecordCreateRequest{
			RecordDate: recordDate,
			Weight:     10.0,
			Height:     30.0,
			Title:      "First Record",
			Content:    "Good boy",
			Images:     []string{"http://example.com/1.jpg"},
		})
		if err != nil {
			t.Fatalf("create growth record failed: %v", err)
		}
		if info.Title != "First Record" {
			t.Errorf("expected title First Record, got %s", info.Title)
		}
		if len(info.Images) != 1 {
			t.Errorf("expected 1 image, got %d", len(info.Images))
		}
	})

	t.Run("no permission", func(t *testing.T) {
		_, err := srv.CreateGrowthRecord(2, pet.ID, &types.PetGrowthRecordCreateRequest{
			RecordDate: recordDate,
			Title:      "Hack",
		})
		if err == nil {
			t.Error("expected error for no permission")
		}
	})
}

func TestPetServiceUpdateGrowthRecord(t *testing.T) {
	db := testutil.SetupTestDB(t)

	petRepo := repository.NewPetRepository(db)
	srv := service.NewPetService(petRepo)

	pet := &model.Pet{UserID: 1, Name: "Buddy", PetType: 1, Status: 1}
	db.Create(pet)

	record := &model.PetGrowthRecord{
		PetID:      pet.ID,
		RecordDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Title:      "Old Title",
	}
	db.Create(record)

	t.Run("success", func(t *testing.T) {
		info, err := srv.UpdateGrowthRecord(1, pet.ID, record.ID, &types.PetGrowthRecordUpdateRequest{
			Title:   "New Title",
			Weight:  15.0,
			Content: "Updated",
			Images:  []string{"a.jpg", "b.jpg"},
		})
		if err != nil {
			t.Fatalf("update growth record failed: %v", err)
		}
		if info.Title != "New Title" {
			t.Errorf("expected title New Title, got %s", info.Title)
		}
		if len(info.Images) != 2 {
			t.Errorf("expected 2 images, got %d", len(info.Images))
		}
	})

	t.Run("record not match pet", func(t *testing.T) {
		_, err := srv.UpdateGrowthRecord(1, 9999, record.ID, &types.PetGrowthRecordUpdateRequest{
			Title: "Hack",
		})
		if err == nil {
			t.Error("expected error for record not match pet")
		}
	})
}

func TestPetServiceDeleteGrowthRecord(t *testing.T) {
	db := testutil.SetupTestDB(t)

	petRepo := repository.NewPetRepository(db)
	srv := service.NewPetService(petRepo)

	pet := &model.Pet{UserID: 1, Name: "Buddy", PetType: 1, Status: 1}
	db.Create(pet)

	record := &model.PetGrowthRecord{
		PetID:      pet.ID,
		RecordDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Title:      "Record",
	}
	db.Create(record)

	t.Run("success", func(t *testing.T) {
		err := srv.DeleteGrowthRecord(1, pet.ID, record.ID)
		if err != nil {
			t.Fatalf("delete growth record failed: %v", err)
		}
	})

	t.Run("no permission", func(t *testing.T) {
		err := srv.DeleteGrowthRecord(2, pet.ID, record.ID)
		if err == nil {
			t.Error("expected error for no permission")
		}
	})
}

func TestPetServiceGetGrowthRecords(t *testing.T) {
	db := testutil.SetupTestDB(t)

	petRepo := repository.NewPetRepository(db)
	srv := service.NewPetService(petRepo)

	pet := &model.Pet{UserID: 1, Name: "Buddy", PetType: 1, Status: 1}
	db.Create(pet)

	db.Create(&model.PetGrowthRecord{PetID: pet.ID, RecordDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Title: "A"})
	db.Create(&model.PetGrowthRecord{PetID: pet.ID, RecordDate: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), Title: "B"})

	records, total, err := srv.GetGrowthRecords(pet.ID, 1, 10)
	if err != nil {
		t.Fatalf("get growth records failed: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if len(records) != 2 {
		t.Errorf("expected 2 records, got %d", len(records))
	}
}

func TestPetServiceCreateLostPet(t *testing.T) {
	db := testutil.SetupTestDB(t)

	petRepo := repository.NewPetRepository(db)
	srv := service.NewPetService(petRepo)

	t.Run("success", func(t *testing.T) {
		info, err := srv.CreateLostPet(1, &types.LostPetCreateRequest{
			PetName:      "Buddy",
			PetType:      1,
			Description:  "Lost dog",
			LostLocation: "Beijing",
			LostTime:     time.Now(),
			Contact:      "13800138000",
			Images:       []string{"http://example.com/1.jpg"},
		})
		if err != nil {
			t.Fatalf("create lost pet failed: %v", err)
		}
		if info.ID == 0 {
			t.Error("lost pet ID should not be zero")
		}
		if len(info.Images) != 1 {
			t.Errorf("expected 1 image, got %d", len(info.Images))
		}
	})
}

func TestPetServiceGetLostPetByID(t *testing.T) {
	db := testutil.SetupTestDB(t)

	petRepo := repository.NewPetRepository(db)
	srv := service.NewPetService(petRepo)

	lostPet := &model.LostPet{
		UserID:       1,
		PetName:      "Buddy",
		PetType:      1,
		Description:  "Lost",
		LostLocation: "Beijing",
		LostTime:     time.Now(),
		Contact:      "13800138000",
		Status:       1,
	}
	db.Create(lostPet)

	t.Run("success", func(t *testing.T) {
		info, err := srv.GetLostPetByID(lostPet.ID, 1)
		if err != nil {
			t.Fatalf("get lost pet failed: %v", err)
		}
		if info.ID != lostPet.ID {
			t.Errorf("expected lost pet ID %d, got %d", lostPet.ID, info.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := srv.GetLostPetByID(9999, 0)
		if err == nil {
			t.Error("expected error for non-existent lost pet")
		}
	})
}

func TestPetServiceListLostPets(t *testing.T) {
	db := testutil.SetupTestDB(t)

	petRepo := repository.NewPetRepository(db)
	srv := service.NewPetService(petRepo)

	db.Create(&model.LostPet{UserID: 1, PetName: "A", PetType: 1, LostLocation: "Beijing", LostTime: time.Now(), Contact: "1", Status: 1})
	db.Create(&model.LostPet{UserID: 2, PetName: "B", PetType: 1, LostLocation: "Shanghai", LostTime: time.Now(), Contact: "2", Status: 1})
	db.Create(&model.LostPet{UserID: 3, PetName: "C", PetType: 2, LostLocation: "Beijing", LostTime: time.Now(), Contact: "3", Status: 2})

	t.Run("filter by status", func(t *testing.T) {
		resp, err := srv.ListLostPets(&types.LostPetListRequest{Status: 2, Page: 1, PageSize: 10}, 0)
		if err != nil {
			t.Fatalf("list lost pets failed: %v", err)
		}
		if resp.Total != 1 {
			t.Errorf("expected total 1, got %d", resp.Total)
		}
	})

	t.Run("filter by pet type", func(t *testing.T) {
		resp, err := srv.ListLostPets(&types.LostPetListRequest{PetType: 2, Status: -1, Page: 1, PageSize: 10}, 1)
		if err != nil {
			t.Fatalf("list lost pets failed: %v", err)
		}
		if resp.Total != 1 {
			t.Errorf("expected total 1, got %d", resp.Total)
		}
	})
}

func TestPetServiceUpdateLostPetStatus(t *testing.T) {
	db := testutil.SetupTestDB(t)

	petRepo := repository.NewPetRepository(db)
	srv := service.NewPetService(petRepo)

	lostPet := &model.LostPet{
		UserID:       1,
		PetName:      "Buddy",
		PetType:      1,
		Description:  "Lost",
		LostLocation: "Beijing",
		LostTime:     time.Now(),
		Contact:      "13800138000",
		Status:       1,
	}
	db.Create(lostPet)

	t.Run("success", func(t *testing.T) {
		err := srv.UpdateLostPetStatus(1, lostPet.ID, 2)
		if err != nil {
			t.Fatalf("update lost pet status failed: %v", err)
		}
		updated, _ := petRepo.GetLostPetByID(lostPet.ID)
		if updated.Status != 2 {
			t.Errorf("expected status 2, got %d", updated.Status)
		}
	})

	t.Run("no permission", func(t *testing.T) {
		err := srv.UpdateLostPetStatus(2, lostPet.ID, 3)
		if err == nil {
			t.Error("expected error for no permission")
		}
	})
}
