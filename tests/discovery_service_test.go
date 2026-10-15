package tests

import (
	"testing"
	"time"

	"gorm.io/gorm"

	"pet-community/model"
	"pet-community/repository"
	"pet-community/service"
	"pet-community/tests/testutil"
	"pet-community/types"
)

func setupDiscoveryService(db *gorm.DB) *service.DiscoveryService {
	postRepo := repository.NewPostRepository(db)
	userRepo := repository.NewUserRepository(db)
	topicRepo := repository.NewTopicRepository(db)
	likeRepo := repository.NewLikeRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	petRepo := repository.NewPetRepository(db)
	searchHistoryRepo := repository.NewSearchHistoryRepository(db)
	userInterestRepo := repository.NewUserInterestRepository(db)
	topicCategoryRepo := repository.NewTopicCategoryRepository(db)
	topicParticipantRepo := repository.NewTopicParticipantRepository(db)
	hotKeywordRepo := repository.NewHotKeywordRepository(db)
	followRepo := repository.NewFollowRepository(db)
	return service.NewDiscoveryService(
		postRepo, userRepo, topicRepo, likeRepo, favoriteRepo, petRepo,
		searchHistoryRepo, userInterestRepo, topicCategoryRepo, topicParticipantRepo,
		hotKeywordRepo, followRepo,
	)
}

func TestDiscoveryServiceSearch(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.SetGlobalDB(db)
	srv := setupDiscoveryService(db)

	db.Create(&model.User{Nickname: "Alice", Phone: "alice_phone", Email: "alice@test.com", Status: 1})
	db.Create(&model.User{Nickname: "Bob", Phone: "bob_phone", Email: "bob@test.com", Status: 1})
	db.Create(&model.Post{UserID: 1, Content: "hello world", Status: 1, Visibility: 0})
	db.Create(&model.Topic{Name: "#Hello", Description: "hello topic", Status: 1, PostsCount: 5})

	t.Run("empty keyword", func(t *testing.T) {
		_, err := srv.Search("", 1)
		if err == nil {
			t.Error("expected error for empty keyword")
		}
	})

	t.Run("success", func(t *testing.T) {
		result, err := srv.Search("hello", 1)
		if err != nil {
			t.Fatalf("search failed: %v", err)
		}
		if len(result.Users) == 0 && len(result.Posts) == 0 && len(result.Topics) == 0 {
			t.Error("expected at least one search result")
		}
	})
}

func TestDiscoveryServiceRecommendPosts(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.SetGlobalDB(db)
	srv := setupDiscoveryService(db)

	db.Create(&model.User{ID: 1, Nickname: "Alice", Phone: "alice_phone", Email: "alice@test.com", Status: 1})
	db.Create(&model.Post{UserID: 1, Content: "hot post", Status: 1, Visibility: 0, LikesCount: 10})
	db.Create(&model.Post{UserID: 1, Content: "normal post", Status: 1, Visibility: 0, LikesCount: 1})
	db.Create(&model.Post{UserID: 1, Content: "draft post", Status: 0, Visibility: 0})

	t.Run("list public", func(t *testing.T) {
		resp, err := srv.GetRecommendPosts(0, 1, 10, "")
		if err != nil {
			t.Fatalf("get recommend posts failed: %v", err)
		}
		if resp.Total != 2 {
			t.Errorf("expected total 2, got %d", resp.Total)
		}
		if len(resp.List) != 2 {
			t.Errorf("expected 2 posts, got %d", len(resp.List))
		}
	})

	t.Run("filter image", func(t *testing.T) {
		db.Create(&model.Post{UserID: 1, Content: "image post", Status: 1, Visibility: 0, Images: "a.jpg"})
		resp, err := srv.GetRecommendPosts(0, 1, 10, "image")
		if err != nil {
			t.Fatalf("get image posts failed: %v", err)
		}
		if resp.Total != 1 {
			t.Errorf("expected total 1 image post, got %d", resp.Total)
		}
	})
}

func TestDiscoveryServiceHotRanking(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.SetGlobalDB(db)
	srv := setupDiscoveryService(db)

	db.Create(&model.User{ID: 1, Nickname: "Alice", Phone: "alice_phone", Email: "alice@test.com", Status: 1})
	db.Create(&model.Post{UserID: 1, Content: "trending", Status: 1, Visibility: 0, LikesCount: 5, CommentsCount: 3, SharesCount: 2})
	db.Create(&model.Topic{Name: "#Hot", Description: "hot", Status: 1, PostsCount: 10})

	t.Run("daily ranking", func(t *testing.T) {
		resp, err := srv.GetHotRanking("day")
		if err != nil {
			t.Fatalf("get hot ranking failed: %v", err)
		}
		if len(resp.Posts) == 0 {
			t.Error("expected at least one post in ranking")
		}
		if len(resp.Users) == 0 {
			t.Error("expected at least one user in ranking")
		}
		if len(resp.Topics) == 0 {
			t.Error("expected at least one topic in ranking")
		}
	})
}

func TestDiscoveryServiceTagCloud(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.SetGlobalDB(db)
	srv := setupDiscoveryService(db)

	db.Create(&model.Topic{Name: "#A", Status: 1, PostsCount: 100})
	db.Create(&model.Topic{Name: "#B", Status: 1, PostsCount: 50})
	db.Create(&model.Topic{Name: "#C", Status: 1, PostsCount: 10})

	resp, err := srv.GetTagCloud()
	if err != nil {
		t.Fatalf("get tag cloud failed: %v", err)
	}
	if len(resp.Tags) != 3 {
		t.Errorf("expected 3 tags, got %d", len(resp.Tags))
	}
}

func TestDiscoveryServiceTopicCategories(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.SetGlobalDB(db)
	srv := setupDiscoveryService(db)

	db.Create(&model.TopicCategory{Name: "Daily", Description: "daily", Status: 1, SortOrder: 1})
	db.Create(&model.TopicCategory{Name: "Hidden", Description: "hidden", Status: 0, SortOrder: 2})

	cats, err := srv.GetTopicCategories()
	if err != nil {
		t.Fatalf("get topic categories failed: %v", err)
	}
	if len(cats) != 2 {
		t.Errorf("expected 2 categories (default Status is 1), got %d", len(cats))
	}
}

func TestDiscoveryServiceTopicDetail(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.SetGlobalDB(db)
	srv := setupDiscoveryService(db)

	topic := &model.Topic{Name: "#Topic", Description: "desc", Status: 1, PostsCount: 5}
	db.Create(topic)
	db.Create(&model.User{ID: 1, Nickname: "Alice", Phone: "alice_phone", Email: "alice@test.com", Status: 1})
	db.Create(&model.TopicParticipant{TopicID: topic.ID, UserID: 1})

	t.Run("success", func(t *testing.T) {
		info, err := srv.GetTopicDetail(topic.ID, 1)
		if err != nil {
			t.Fatalf("get topic detail failed: %v", err)
		}
		if info.ID != topic.ID {
			t.Errorf("expected topic ID %d, got %d", topic.ID, info.ID)
		}
		if !info.IsParticipated {
			t.Error("expected user participated")
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := srv.GetTopicDetail(9999, 1)
		if err == nil {
			t.Error("expected error for non-existent topic")
		}
	})
}

func TestDiscoveryServiceTopicPosts(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.SetGlobalDB(db)
	srv := setupDiscoveryService(db)

	topicID := uint(1)
	db.Create(&model.User{ID: 1, Nickname: "Alice", Phone: "alice_phone", Email: "alice@test.com", Status: 1})
	db.Create(&model.Post{UserID: 1, Content: "topic post", Status: 1, Visibility: 0, TopicIDs: "1"})

	resp, err := srv.GetTopicPosts(topicID, 0, 1, 10)
	if err != nil {
		t.Fatalf("get topic posts failed: %v", err)
	}
	if resp.Total != 1 {
		t.Errorf("expected total 1, got %d", resp.Total)
	}
}

func TestDiscoveryServiceParticipateTopic(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.SetGlobalDB(db)
	srv := setupDiscoveryService(db)

	topic := &model.Topic{Name: "#Join", Description: "join", Status: 1}
	db.Create(topic)

	t.Run("participate", func(t *testing.T) {
		err := srv.ParticipateTopic(topic.ID, 1)
		if err != nil {
			t.Fatalf("participate failed: %v", err)
		}
	})

	t.Run("cancel", func(t *testing.T) {
		err := srv.ParticipateTopic(topic.ID, 1)
		if err != nil {
			t.Fatalf("cancel participation failed: %v", err)
		}
	})
}

func TestDiscoveryServiceCreateTopic(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.SetGlobalDB(db)
	srv := setupDiscoveryService(db)

	db.Create(&model.User{ID: 1, Nickname: "Alice", Phone: "alice_phone", Email: "alice@test.com", Status: 1})
	db.Create(&model.UserLevel{UserID: 1, Level: 5, Points: 800, TotalPoints: 800})

	t.Run("success", func(t *testing.T) {
		info, err := srv.CreateTopic(1, &types.TopicCreateRequest{
			Name:        "#NewTopic",
			Description: "new topic",
		})
		if err != nil {
			t.Fatalf("create topic failed: %v", err)
		}
		if info.ID == 0 {
			t.Error("topic ID should not be zero")
		}
		if info.Status != 0 {
			t.Errorf("expected pending status 0, got %d", info.Status)
		}
	})

	t.Run("duplicate", func(t *testing.T) {
		_, err := srv.CreateTopic(1, &types.TopicCreateRequest{
			Name:        "#NewTopic",
			Description: "duplicate",
		})
		if err == nil {
			t.Error("expected error for duplicate topic")
		}
	})

	t.Run("level too low", func(t *testing.T) {
		db.Create(&model.User{ID: 2, Nickname: "Bob", Phone: "bob_phone", Email: "bob@test.com", Status: 1})
		db.Create(&model.UserLevel{UserID: 2, Level: 1, Points: 0, TotalPoints: 0})
		_, err := srv.CreateTopic(2, &types.TopicCreateRequest{Name: "#LowLevel"})
		if err == nil {
			t.Error("expected Lv5 gate error")
		}
	})
}

func TestDiscoveryServiceTrackUserInterest(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.SetGlobalDB(db)
	srv := setupDiscoveryService(db)

	db.Create(&model.User{ID: 1, Nickname: "Alice", Phone: "alice_phone", Email: "alice@test.com", Status: 1})
	topic := &model.Topic{Name: "#Interest", Status: 1}
	db.Create(topic)
	db.Create(&model.Post{UserID: 1, Content: "post", Status: 1, Visibility: 0, TopicIDs: "1"})

	srv.TrackUserInterest(1, "view_post", 1)

	interests, err := repository.NewUserInterestRepository(db).ListByUser(1)
	if err != nil {
		t.Fatalf("list interests failed: %v", err)
	}
	if len(interests) != 1 {
		t.Errorf("expected 1 interest, got %d", len(interests))
	}
}

func TestDiscoveryServiceSearchHistory(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.SetGlobalDB(db)
	srv := setupDiscoveryService(db)

	db.Create(&model.SearchHistory{UserID: 1, Keyword: "cat", CreatedAt: time.Now()})
	db.Create(&model.SearchHistory{UserID: 1, Keyword: "dog", CreatedAt: time.Now().Add(-time.Hour)})

	t.Run("list", func(t *testing.T) {
		histories, err := srv.GetSearchHistory(1, 10)
		if err != nil {
			t.Fatalf("get search history failed: %v", err)
		}
		if len(histories) != 2 {
			t.Errorf("expected 2 histories, got %d", len(histories))
		}
	})

	t.Run("clear", func(t *testing.T) {
		err := srv.ClearSearchHistory(1)
		if err != nil {
			t.Fatalf("clear search history failed: %v", err)
		}
		histories, _ := srv.GetSearchHistory(1, 10)
		if len(histories) != 0 {
			t.Errorf("expected 0 histories after clear, got %d", len(histories))
		}
	})
}

func TestDiscoveryServiceHotKeywords(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.SetGlobalDB(db)
	srv := setupDiscoveryService(db)

	db.Create(&model.HotKeyword{Keyword: "cat", SearchCount: 10, Rank: 1})
	db.Create(&model.HotKeyword{Keyword: "dog", SearchCount: 5, Rank: 2})

	keywords, err := srv.GetHotKeywords(10)
	if err != nil {
		t.Fatalf("get hot keywords failed: %v", err)
	}
	if len(keywords) != 2 {
		t.Errorf("expected 2 keywords, got %d", len(keywords))
	}
}
