package tests

import (
	"testing"

	"pet-community/controller"
	"pet-community/router"

	"github.com/gin-gonic/gin"
)

// TestActivityRoutesRegisterWithoutConflict 只验证路由树能注册成功。
// gin 在「同一路径段既有静态子节点又有通配子节点」时会 panic，
// 活动侧同时存在 /activities/:id/sign-ins 与 /activities/:id/sign-ins/stats，
// 用一个不执行业务的 nil controller 就能在单测阶段暴露这类冲突。
func TestActivityRoutesRegisterWithoutConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	group := gin.New().Group("/api/v1")

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("activity routes registration panicked: %v", r)
		}
	}()

	router.SetupActivityRoutes(group, (*controller.ActivityController)(nil))
}
