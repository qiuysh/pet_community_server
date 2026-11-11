package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"pet-community/controller"
	"pet-community/repository"
	"pet-community/service"
	"pet-community/tests/testutil"
	"pet-community/types"
)

// 回归：详情页并行拉取回顾；未发布时必须 200 + null，不可 HTTP 404
//（前端 axios 对 404 会 toast「请求的资源不存在」）。
func TestGetReview_MissingReturnsNullSuccess(t *testing.T) {
	db := testutil.SetupTestDB(t)
	activity := createActiveActivity(db)

	reviewRepo := repository.NewActivityReviewRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	userRepo := repository.NewUserRepository(db)
	reviewService := service.NewActivityReviewService(reviewRepo, activityRepo, userRepo)
	ctrl := controller.NewActivityController(nil, nil, nil, nil, nil, nil, reviewService)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/activities/:id/review", ctrl.GetReview)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/activities/%d/review", activity.ID), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp types.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("expected business code 0, got %d message=%s", resp.Code, resp.Message)
	}
	if resp.Data != nil {
		t.Fatalf("expected null data for missing review, got %#v", resp.Data)
	}
}

func TestGetReview_ServiceMissingIsNil(t *testing.T) {
	db := testutil.SetupTestDB(t)
	activity := createActiveActivity(db)

	reviewRepo := repository.NewActivityReviewRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	userRepo := repository.NewUserRepository(db)
	srv := service.NewActivityReviewService(reviewRepo, activityRepo, userRepo)

	got, err := srv.GetReview(activity.ID)
	if err != nil {
		t.Fatalf("GetReview: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil review, got %#v", got)
	}
}
