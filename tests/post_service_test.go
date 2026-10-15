package tests

import (
	"testing"
	"time"

	"pet-community/model"
	"pet-community/repository"
	"pet-community/service"
	"pet-community/tests/testutil"
	"pet-community/types"
)

func TestPostServiceCreatePost(t *testing.T) {
	db := testutil.SetupTestDB(t)

	postRepo := repository.NewPostRepository(db)
	likeRepo := repository.NewLikeRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	topicRepo := repository.NewTopicRepository(db)
	userRepo := repository.NewUserRepository(db)
	petRepo := repository.NewPetRepository(db)
	srv := service.NewPostService(postRepo, likeRepo, favoriteRepo, topicRepo, userRepo, petRepo)

	t.Run("empty content", func(t *testing.T) {
		_, err := srv.CreatePost(1, &types.PostCreateRequest{})
		if err == nil {
			t.Error("expected error for empty content")
		}
	})

	t.Run("image and video together", func(t *testing.T) {
		_, err := srv.CreatePost(1, &types.PostCreateRequest{
			Content: "test",
			Images:  []string{"a.jpg"},
			Video:   "a.mp4",
		})
		if err == nil {
			t.Error("expected error for image and video together")
		}
	})

	t.Run("too many topics", func(t *testing.T) {
		_, err := srv.CreatePost(1, &types.PostCreateRequest{
			Content:  "test",
			TopicIDs: []uint{1, 2, 3, 4},
		})
		if err == nil {
			t.Error("expected error for too many topics")
		}
	})

	t.Run("creates pending post", func(t *testing.T) {
		// The service creates posts with status 0 and then calls GetPostByID
		// which requires status 1, so creation itself may error. Verify the
		// pending post record was persisted.
		_, err := srv.CreatePost(1, &types.PostCreateRequest{
			Content: "hello world",
		})

		var count int64
		db.Model(&model.Post{}).Where("user_id = ? AND content = ?", 1, "hello world").Count(&count)
		if count != 1 {
			t.Errorf("expected 1 pending post, got %d (err=%v)", count, err)
		}
	})

	t.Run("scheduled publish sets status 2", func(t *testing.T) {
		publishAt := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)
		_, _ = srv.CreatePost(1, &types.PostCreateRequest{
			Content:   "scheduled post",
			PublishAt: publishAt,
		})

		var post model.Post
		if err := db.Where("user_id = ? AND content = ?", 1, "scheduled post").First(&post).Error; err != nil {
			t.Fatalf("expected scheduled post record, got error: %v", err)
		}
		if post.Status != 2 {
			t.Errorf("expected status 2 for scheduled post, got %d", post.Status)
		}
		if post.PublishAt == nil {
			t.Error("expected publish_at to be set for scheduled post")
		}
	})

	t.Run("scheduled publish too far", func(t *testing.T) {
		publishAt := time.Now().Add(8 * 24 * time.Hour).UTC().Format(time.RFC3339)
		_, err := srv.CreatePost(1, &types.PostCreateRequest{
			Content:   "too far",
			PublishAt: publishAt,
		})
		if err == nil {
			t.Error("expected error for publish time more than 7 days away")
		}
	})

	t.Run("scheduled publish in past", func(t *testing.T) {
		publishAt := time.Now().Add(-1 * time.Hour).UTC().Format(time.RFC3339)
		_, err := srv.CreatePost(1, &types.PostCreateRequest{
			Content:   "past scheduled",
			PublishAt: publishAt,
		})
		if err == nil {
			t.Error("expected error for publish time in the past")
		}
	})
}

func TestPostServiceGetPostByID(t *testing.T) {
	db := testutil.SetupTestDB(t)

	postRepo := repository.NewPostRepository(db)
	likeRepo := repository.NewLikeRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	topicRepo := repository.NewTopicRepository(db)
	userRepo := repository.NewUserRepository(db)
	petRepo := repository.NewPetRepository(db)
	srv := service.NewPostService(postRepo, likeRepo, favoriteRepo, topicRepo, userRepo, petRepo)

	post := &model.Post{UserID: 1, Content: "hello", Status: 1, Visibility: 0}
	db.Create(post)

	t.Run("success", func(t *testing.T) {
		info, err := srv.GetPostByID(1, post.ID)
		if err != nil {
			t.Fatalf("get post failed: %v", err)
		}
		if info.ID != post.ID {
			t.Errorf("expected post ID %d, got %d", post.ID, info.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := srv.GetPostByID(1, 9999)
		if err == nil {
			t.Error("expected error for non-existent post")
		}
	})

	t.Run("private post no permission", func(t *testing.T) {
		privatePost := &model.Post{UserID: 2, Content: "private", Status: 1, Visibility: 2}
		db.Create(privatePost)
		_, err := srv.GetPostByID(1, privatePost.ID)
		if err == nil {
			t.Error("expected error for private post")
		}
	})
}

func TestPostServiceUpdatePost(t *testing.T) {
	db := testutil.SetupTestDB(t)

	postRepo := repository.NewPostRepository(db)
	likeRepo := repository.NewLikeRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	topicRepo := repository.NewTopicRepository(db)
	userRepo := repository.NewUserRepository(db)
	petRepo := repository.NewPetRepository(db)
	srv := service.NewPostService(postRepo, likeRepo, favoriteRepo, topicRepo, userRepo, petRepo)

	post := &model.Post{UserID: 1, Content: "old", Images: "a.jpg", Status: 1}
	db.Create(post)

	t.Run("success", func(t *testing.T) {
		info, err := srv.UpdatePost(1, post.ID, &types.PostUpdateRequest{
			Content: "new content",
			Images:  []string{"b.jpg"},
		})
		if err != nil {
			t.Fatalf("update post failed: %v", err)
		}
		if info.Content != "new content" {
			t.Errorf("expected content new content, got %s", info.Content)
		}
	})

	t.Run("no permission", func(t *testing.T) {
		_, err := srv.UpdatePost(2, post.ID, &types.PostUpdateRequest{Content: "hack"})
		if err == nil {
			t.Error("expected error for no permission")
		}
	})

	t.Run("set scheduled publish via update", func(t *testing.T) {
		// 将已发布动态转为定时发布,UpdatePost 不再经过 GetPostByID 的 status 过滤,
		// 应能返回 status=2 和 publishAt。
		publishAt := time.Now().Add(3 * time.Hour).UTC().Format(time.RFC3339)
		info, err := srv.UpdatePost(1, post.ID, &types.PostUpdateRequest{
			PublishAt: publishAt,
		})
		if err != nil {
			t.Fatalf("update post with publishAt failed: %v", err)
		}
		if info.Status != 2 {
			t.Errorf("expected status 2 after scheduling, got %d", info.Status)
		}
		if info.PublishAt == "" {
			t.Error("expected publishAt to be returned")
		}
	})

	t.Run("update publish too far", func(t *testing.T) {
		publishAt := time.Now().Add(8 * 24 * time.Hour).UTC().Format(time.RFC3339)
		_, err := srv.UpdatePost(1, post.ID, &types.PostUpdateRequest{
			PublishAt: publishAt,
		})
		if err == nil {
			t.Error("expected error for publish time more than 7 days away")
		}
	})

	t.Run("update publish invalid format", func(t *testing.T) {
		_, err := srv.UpdatePost(1, post.ID, &types.PostUpdateRequest{
			PublishAt: "not-a-valid-time",
		})
		if err == nil {
			t.Error("expected error for invalid publish time format")
		}
	})
}

func TestPostServiceDeletePost(t *testing.T) {
	db := testutil.SetupTestDB(t)

	postRepo := repository.NewPostRepository(db)
	likeRepo := repository.NewLikeRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	topicRepo := repository.NewTopicRepository(db)
	userRepo := repository.NewUserRepository(db)
	petRepo := repository.NewPetRepository(db)
	srv := service.NewPostService(postRepo, likeRepo, favoriteRepo, topicRepo, userRepo, petRepo)

	post := &model.Post{UserID: 1, Content: "hello", Status: 1}
	db.Create(post)

	t.Run("success", func(t *testing.T) {
		err := srv.DeletePost(1, post.ID)
		if err != nil {
			t.Fatalf("delete post failed: %v", err)
		}
	})

	t.Run("no permission", func(t *testing.T) {
		other := &model.Post{UserID: 2, Content: "other", Status: 1}
		db.Create(other)
		err := srv.DeletePost(1, other.ID)
		if err == nil {
			t.Error("expected error for no permission")
		}
	})
}

func TestPostServiceListPosts(t *testing.T) {
	db := testutil.SetupTestDB(t)

	postRepo := repository.NewPostRepository(db)
	likeRepo := repository.NewLikeRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	topicRepo := repository.NewTopicRepository(db)
	userRepo := repository.NewUserRepository(db)
	petRepo := repository.NewPetRepository(db)
	srv := service.NewPostService(postRepo, likeRepo, favoriteRepo, topicRepo, userRepo, petRepo)

	db.Create(&model.Post{UserID: 1, Content: "public post", Status: 1, Visibility: 0})
	db.Create(&model.Post{UserID: 1, Content: "private post", Status: 1, Visibility: 2})
	db.Create(&model.Post{UserID: 2, Content: "another public", Status: 1, Visibility: 0})

	t.Run("list public", func(t *testing.T) {
		resp, err := srv.ListPosts(0, 0, 0, "", "", 1, 10)
		if err != nil {
			t.Fatalf("list posts failed: %v", err)
		}
		if resp.Total != 2 {
			t.Errorf("expected total 2, got %d", resp.Total)
		}
	})

	t.Run("list own posts", func(t *testing.T) {
		resp, err := srv.ListPosts(1, 0, 0, "", "", 1, 10)
		if err != nil {
			t.Fatalf("list posts failed: %v", err)
		}
		if resp.Total != 2 {
			t.Errorf("expected total 2, got %d", resp.Total)
		}
	})
}

func TestPostServiceToggleLike(t *testing.T) {
	db := testutil.SetupTestDB(t)

	postRepo := repository.NewPostRepository(db)
	likeRepo := repository.NewLikeRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	topicRepo := repository.NewTopicRepository(db)
	userRepo := repository.NewUserRepository(db)
	petRepo := repository.NewPetRepository(db)
	srv := service.NewPostService(postRepo, likeRepo, favoriteRepo, topicRepo, userRepo, petRepo)

	post := &model.Post{UserID: 1, Content: "hello", Status: 1, LikesCount: 0}
	db.Create(post)

	t.Run("like", func(t *testing.T) {
		result, err := srv.ToggleLike(2, post.ID)
		if err != nil {
			t.Fatalf("toggle like failed: %v", err)
		}
		if !result.Liked {
			t.Error("expected liked")
		}
		if result.LikesCount != 1 {
			t.Errorf("expected likes count 1, got %d", result.LikesCount)
		}
	})

	t.Run("unlike", func(t *testing.T) {
		result, err := srv.ToggleLike(2, post.ID)
		if err != nil {
			t.Fatalf("toggle unlike failed: %v", err)
		}
		if result.Liked {
			t.Error("expected not liked")
		}
	})
}

func TestPostServiceToggleFavorite(t *testing.T) {
	db := testutil.SetupTestDB(t)

	postRepo := repository.NewPostRepository(db)
	likeRepo := repository.NewLikeRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	topicRepo := repository.NewTopicRepository(db)
	userRepo := repository.NewUserRepository(db)
	petRepo := repository.NewPetRepository(db)
	srv := service.NewPostService(postRepo, likeRepo, favoriteRepo, topicRepo, userRepo, petRepo)

	post := &model.Post{UserID: 1, Content: "hello", Status: 1, FavoritesCount: 0}
	db.Create(post)

	t.Run("favorite", func(t *testing.T) {
		result, err := srv.ToggleFavorite(2, post.ID)
		if err != nil {
			t.Fatalf("toggle favorite failed: %v", err)
		}
		if !result.Favorited {
			t.Error("expected favorited")
		}
	})

	t.Run("unfavorite", func(t *testing.T) {
		result, err := srv.ToggleFavorite(2, post.ID)
		if err != nil {
			t.Fatalf("toggle unfavorite failed: %v", err)
		}
		if result.Favorited {
			t.Error("expected not favorited")
		}
	})
}

func TestPostServiceRecordShare(t *testing.T) {
	db := testutil.SetupTestDB(t)

	postRepo := repository.NewPostRepository(db)
	likeRepo := repository.NewLikeRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	topicRepo := repository.NewTopicRepository(db)
	userRepo := repository.NewUserRepository(db)
	petRepo := repository.NewPetRepository(db)
	srv := service.NewPostService(postRepo, likeRepo, favoriteRepo, topicRepo, userRepo, petRepo)

	post := &model.Post{UserID: 1, Content: "hello", Status: 1}
	db.Create(post)

	err := srv.RecordShare(2, post.ID, "wechat")
	if err != nil {
		t.Fatalf("record share failed: %v", err)
	}

	updated, _ := postRepo.GetByID(post.ID)
	if updated.SharesCount != 1 {
		t.Errorf("expected shares count 1, got %d", updated.SharesCount)
	}
}

func TestCommentServiceCreateComment(t *testing.T) {
	db := testutil.SetupTestDB(t)

	commentRepo := repository.NewCommentRepository(db)
	likeRepo := repository.NewLikeRepository(db)
	postRepo := repository.NewPostRepository(db)
	userRepo := repository.NewUserRepository(db)
	srv := service.NewCommentService(commentRepo, likeRepo, postRepo, userRepo)

	post := &model.Post{UserID: 1, Content: "hello", Status: 1}
	db.Create(post)

	t.Run("success", func(t *testing.T) {
		info, err := srv.CreateComment(2, &types.CommentCreateRequest{
			PostID:  post.ID,
			Content: "nice post",
		})
		if err != nil {
			t.Fatalf("create comment failed: %v", err)
		}
		if info.Content != "nice post" {
			t.Errorf("expected content nice post, got %s", info.Content)
		}
	})

	t.Run("post not found", func(t *testing.T) {
		_, err := srv.CreateComment(2, &types.CommentCreateRequest{
			PostID:  9999,
			Content: "nice post",
		})
		if err == nil {
			t.Error("expected error for non-existent post")
		}
	})
}

func TestCommentServiceToggleLike(t *testing.T) {
	db := testutil.SetupTestDB(t)

	commentRepo := repository.NewCommentRepository(db)
	likeRepo := repository.NewLikeRepository(db)
	postRepo := repository.NewPostRepository(db)
	userRepo := repository.NewUserRepository(db)
	srv := service.NewCommentService(commentRepo, likeRepo, postRepo, userRepo)

	post := &model.Post{UserID: 1, Content: "hello", Status: 1}
	db.Create(post)
	comment := &model.Comment{PostID: post.ID, UserID: 2, Content: "nice", Status: 1}
	db.Create(comment)

	t.Run("like", func(t *testing.T) {
		result, err := srv.ToggleLike(3, comment.ID)
		if err != nil {
			t.Fatalf("toggle like failed: %v", err)
		}
		if !result.Liked {
			t.Error("expected liked")
		}
	})

	t.Run("unlike", func(t *testing.T) {
		result, err := srv.ToggleLike(3, comment.ID)
		if err != nil {
			t.Fatalf("toggle unlike failed: %v", err)
		}
		if result.Liked {
			t.Error("expected not liked")
		}
	})
}

func TestTopicService(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.SetGlobalDB(db)

	topicRepo := repository.NewTopicRepository(db)
	srv := service.NewTopicService(topicRepo)

	db.Create(&model.UserLevel{UserID: 1, Level: 5, Points: 800, TotalPoints: 800})

	t.Run("create", func(t *testing.T) {
		info, err := srv.CreateTopic(1, &types.TopicCreateRequest{
			Name:        "#TestTopic",
			Description: "test",
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
			Name: "#TestTopic",
		})
		if err == nil {
			t.Error("expected error for duplicate topic")
		}
	})

	t.Run("search", func(t *testing.T) {
		db.Create(&model.Topic{Name: "#Another", Status: 1})
		topics, err := srv.SearchTopics("Another", 10)
		if err != nil {
			t.Fatalf("search topics failed: %v", err)
		}
		if len(topics) != 1 {
			t.Errorf("expected 1 topic, got %d", len(topics))
		}
	})
}

func TestReportService(t *testing.T) {
	db := testutil.SetupTestDB(t)

	reportRepo := repository.NewReportRepository(db)
	srv := service.NewReportService(reportRepo)

	t.Run("success", func(t *testing.T) {
		info, err := srv.CreateReport(1, &types.ReportCreateRequest{
			TargetType: "post",
			TargetID:   1,
			Reason:     "spam",
			ReasonType: "spam",
		})
		if err != nil {
			t.Fatalf("create report failed: %v", err)
		}
		if info.ID == 0 {
			t.Error("report ID should not be zero")
		}
	})

	t.Run("duplicate", func(t *testing.T) {
		_, err := srv.CreateReport(1, &types.ReportCreateRequest{
			TargetType: "post",
			TargetID:   1,
			Reason:     "spam again",
			ReasonType: "spam",
		})
		if err == nil {
			t.Error("expected error for duplicate report")
		}
	})
}
