package tests

import (
	"strings"
	"testing"

	"pet-community/model"
	"pet-community/repository"
	"pet-community/service"
	"pet-community/tests/testutil"

	"gorm.io/gorm"
)

func newMessageServiceForTest(db *gorm.DB) *service.MessageService {
	return service.NewMessageService(
		repository.NewConversationRepository(db),
		repository.NewMessageRepository(db),
		repository.NewUserRepository(db),
		repository.NewBlockRepository(db),
	)
}

func TestMessageServiceSendMediaMessage(t *testing.T) {
	db := testutil.SetupTestDB(t)
	srv := newMessageServiceForTest(db)

	t.Run("image message", func(t *testing.T) {
		msg, err := srv.SendMessageWithMedia(
			1, 2, "", model.MessageTypeImage, service.MessageImageURLPrefix()+"a.jpg", 0,
		)
		if err != nil {
			t.Fatalf("send image failed: %v", err)
		}
		if msg.MediaURL != service.MessageImageURLPrefix()+"a.jpg" {
			t.Errorf("unexpected media url: %s", msg.MediaURL)
		}
		if msg.Content != "[图片]" {
			t.Errorf("expected placeholder content, got %q", msg.Content)
		}
		if msg.Duration != 0 {
			t.Errorf("image duration should be 0, got %d", msg.Duration)
		}
	})

	t.Run("voice message", func(t *testing.T) {
		msg, err := srv.SendMessageWithMedia(
			2, 1, "", model.MessageTypeVoice, service.MessageVoiceURLPrefix()+"a.webm", 5,
		)
		if err != nil {
			t.Fatalf("send voice failed: %v", err)
		}
		if msg.Duration != 5 {
			t.Errorf("expected duration 5, got %d", msg.Duration)
		}
		if msg.Content != "[语音]" {
			t.Errorf("expected placeholder content, got %q", msg.Content)
		}
	})

	t.Run("voice duration out of range", func(t *testing.T) {
		for _, duration := range []int{0, 61} {
			if _, err := srv.SendMessageWithMedia(
				1, 2, "", model.MessageTypeVoice, service.MessageVoiceURLPrefix()+"a.webm", duration,
			); err == nil {
				t.Errorf("expected error for duration %d", duration)
			}
		}
	})

	t.Run("media url required", func(t *testing.T) {
		if _, err := srv.SendMessageWithMedia(1, 2, "", model.MessageTypeImage, "", 0); err == nil {
			t.Error("expected error for empty image url")
		}
	})

	t.Run("reject path traversal media url", func(t *testing.T) {
		badURLs := []string{
			service.MessageImageURLPrefix() + "../avatar/x.jpg",
			service.MessageImageURLPrefix() + "..%2Favatar/x.jpg",
			"/api/v1/uploads/messages/image/../../avatar/x.jpg",
			service.MessageImageURLPrefix() + "a/b.jpg",
			service.MessageImageURLPrefix() + "a.jpg?x=1",
			"https://evil.example/a.jpg",
		}
		for _, url := range badURLs {
			if _, err := srv.SendMessageWithMedia(1, 2, "", model.MessageTypeImage, url, 0); err == nil {
				t.Errorf("expected error for media url %q", url)
			}
		}
	})

	t.Run("reject oversized media content", func(t *testing.T) {
		long := strings.Repeat("字", service.MaxMessageContentRunes+1)
		if _, err := srv.SendMessageWithMedia(
			1, 2, long, model.MessageTypeImage, service.MessageImageURLPrefix()+"a.jpg", 0,
		); err == nil {
			t.Error("expected error for content over 500 runes")
		}
	})

	t.Run("unsupported type", func(t *testing.T) {
		if _, err := srv.SendMessageWithMedia(1, 2, "hi", "file", "", 0); err == nil {
			t.Error("expected error for unsupported type")
		}
	})
}

func TestValidateMessageMediaURL(t *testing.T) {
	if err := service.ValidateMessageMediaURL(service.MessageImageURLPrefix()+"ok.jpg", service.MessageImageCategory); err != nil {
		t.Fatalf("valid url rejected: %v", err)
	}
	if err := service.ValidateMessageMediaURL(service.MessageImageURLPrefix()+"../x.jpg", service.MessageImageCategory); err == nil {
		t.Fatal("traversal url should be rejected")
	}

	t.Run("url outside configured prefix rejected", func(t *testing.T) {
		stale := "/uploads/messages/image/old.jpg"
		if err := service.ValidateMessageMediaURL(stale, service.MessageImageCategory); err == nil {
			t.Fatal("stale prefix url should be rejected")
		}
	})
}

func TestMessageServiceSearchMessages(t *testing.T) {
	db := testutil.SetupTestDB(t)
	srv := newMessageServiceForTest(db)

	conv := &model.Conversation{User1ID: 1, User2ID: 2}
	if err := db.Create(conv).Error; err != nil {
		t.Fatalf("create conversation failed: %v", err)
	}

	records := []model.Message{
		{ConversationID: conv.ID, FromID: 1, ToID: 2, Content: "今天猫咪很乖", Type: model.MessageTypeText},
		{ConversationID: conv.ID, FromID: 2, ToID: 1, Content: "猫咪吃饭了吗", Type: model.MessageTypeText},
		{ConversationID: conv.ID, FromID: 1, ToID: 2, Content: "猫咪已撤回", Type: model.MessageTypeText, IsRecalled: 1},
		{ConversationID: conv.ID, FromID: 1, ToID: 2, Content: "[图片]", Type: model.MessageTypeImage, MediaURL: "/uploads/messages/image/x.jpg"},
		{ConversationID: conv.ID, FromID: 2, ToID: 1, Content: "100% 好评", Type: model.MessageTypeText},
	}
	for i := range records {
		if err := db.Create(&records[i]).Error; err != nil {
			t.Fatalf("seed message failed: %v", err)
		}
	}

	t.Run("filters recalled and non-text", func(t *testing.T) {
		list, total, err := srv.SearchMessages(conv.ID, 1, "猫咪", 1, 20)
		if err != nil {
			t.Fatalf("search failed: %v", err)
		}
		if total != 2 {
			t.Errorf("expected total 2, got %d", total)
		}
		if len(list) != 2 {
			t.Errorf("expected 2 rows, got %d", len(list))
		}
	})

	t.Run("pagination", func(t *testing.T) {
		list, total, err := srv.SearchMessages(conv.ID, 1, "猫咪", 1, 1)
		if err != nil {
			t.Fatalf("search failed: %v", err)
		}
		if total != 2 {
			t.Errorf("expected total 2, got %d", total)
		}
		if len(list) != 1 {
			t.Errorf("expected 1 row, got %d", len(list))
		}
	})

	t.Run("wildcard is escaped", func(t *testing.T) {
		_, total, err := srv.SearchMessages(conv.ID, 1, "%", 1, 20)
		if err != nil {
			t.Fatalf("search failed: %v", err)
		}
		if total != 1 {
			t.Errorf("expected only the literal %% message, got total %d", total)
		}
	})
}

func TestMessageServiceConversationMemberGuard(t *testing.T) {
	db := testutil.SetupTestDB(t)
	srv := newMessageServiceForTest(db)

	conv := &model.Conversation{User1ID: 1, User2ID: 2}
	if err := db.Create(conv).Error; err != nil {
		t.Fatalf("create conversation failed: %v", err)
	}
	db.Create(&model.Message{ConversationID: conv.ID, FromID: 1, ToID: 2, Content: "hi", Type: model.MessageTypeText})

	t.Run("outsider cannot read messages", func(t *testing.T) {
		if _, _, err := srv.GetMessages(conv.ID, 3, 1, 20); err == nil {
			t.Error("expected permission error")
		}
	})

	t.Run("outsider cannot search", func(t *testing.T) {
		if _, _, err := srv.SearchMessages(conv.ID, 3, "hi", 1, 20); err == nil {
			t.Error("expected permission error")
		}
	})

	t.Run("outsider cannot mark read", func(t *testing.T) {
		if err := srv.MarkAsRead(conv.ID, 3); err == nil {
			t.Error("expected permission error")
		}
	})

	t.Run("member can read", func(t *testing.T) {
		list, total, err := srv.GetMessages(conv.ID, 2, 1, 20)
		if err != nil {
			t.Fatalf("member read failed: %v", err)
		}
		if total != 1 || len(list) != 1 {
			t.Errorf("expected 1 message, got total=%d len=%d", total, len(list))
		}
	})
}
