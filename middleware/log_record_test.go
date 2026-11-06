package middleware

import "testing"

// 行为日志的类型归档依赖真实路由前缀；路由改名而这里没跟上时，统计会静默归零。
func TestExtractActionTypeMatchesRealRoutes(t *testing.T) {
	cases := []struct {
		path   string
		method string
		want   string
	}{
		{"/api/v1/auth/login", "POST", "login"},
		{"/api/v1/auth/logout", "POST", "logout"},
		{"/api/v1/auth/register", "POST", "register"},
		{"/api/v1/auth/reset-password", "POST", "password_change"},
		{"/api/v1/user/password", "PUT", "password_change"},
		{"/api/v1/user/profile", "PUT", "profile_update"},
		{"/api/v1/posts/12/like", "POST", "like"},
		{"/api/v1/posts/12/favorite", "POST", "favorite"},
		{"/api/v1/posts/12/share", "POST", "share"},
		{"/api/v1/posts/12/comments", "POST", "comment_create"},
		{"/api/v1/posts/comments/9", "DELETE", "comment_delete"},
		{"/api/v1/posts/comments/9/like", "POST", "like"},
		{"/api/v1/posts", "GET", "post_view"},
		{"/api/v1/pets/3", "PUT", "pet_update"},
		{"/api/v1/users/7/follow", "POST", "follow"},
		{"/api/v1/users/7/follow", "DELETE", "unfollow"},
		{"/api/v1/social/follow/7", "POST", "follow"},
		{"/api/v1/activities/3/register", "POST", "activity_register"},
		{"/api/v1/activities/3/sign-in", "POST", "activity_sign_in"},
		{"/api/v1/activities/3/vote", "POST", "activity_vote"},
		{"/api/v1/admin/activities/3/publish", "POST", "activity_publish"},
		{"/api/v1/admin/activities", "POST", "activity_create"},
		{"/api/v1/adopts/applications", "POST", "adopt_create"},
		{"/api/v1/social/messages/send", "POST", "message_send"},
		{"/api/v1/social/points/level", "GET", "points_query"},
		{"/api/v1/unknown", "GET", "api_request"},
	}

	for _, c := range cases {
		if got := extractActionType(c.path, c.method); got != c.want {
			t.Errorf("extractActionType(%q, %q) = %q, want %q", c.path, c.method, got, c.want)
		}
	}
}

func TestExtractTargetType(t *testing.T) {
	cases := map[string]string{
		"/api/v1/posts/comments/9":  "comment",
		"/api/v1/posts/12":          "post",
		"/api/v1/pets/3":            "pet",
		"/api/v1/users/7/follow":    "follow",
		"/api/v1/favorites":         "favorite",
		"/api/v1/likes":             "like",
		"/api/v1/activities/3":      "activity",
		"/api/v1/adopts/2":          "adopt",
		"/api/v1/social/messages/1": "message",
		"/api/v1/social/points":     "points",
		"/api/v1/auth/login":        "user",
		"/api/v1/unknown":           "",
	}

	for path, want := range cases {
		if got := extractTargetType(path); got != want {
			t.Errorf("extractTargetType(%q) = %q, want %q", path, got, want)
		}
	}
}
