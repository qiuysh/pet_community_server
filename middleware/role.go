package middleware

import (
	"net/http"

	"pet-community/types"
	"pet-community/utils"

	"github.com/gin-gonic/gin"
)

const (
	RoleSuperAdmin   = "super_admin"
	RoleContentAdmin = "content_admin"
	RoleDataAdmin    = "data_admin"
	RoleService      = "service"
)

var rolePermissions = map[string][]string{
	RoleSuperAdmin: {
		"admin:users:manage",
		"admin:settings:manage",
		"admin:logs:view",
		"admin:content:audit",
		"admin:data:view",
		"admin:service:manage",
	},
	RoleContentAdmin: {
		"admin:content:audit",
		"admin:logs:view",
	},
	RoleDataAdmin: {
		"admin:data:view",
		"admin:logs:view",
	},
	RoleService: {
		"admin:service:manage",
		"admin:logs:view",
	},
}

// NormalizeRole 遗留 role=admin 视同 super_admin
func NormalizeRole(role string) string {
	if role == "admin" {
		return RoleSuperAdmin
	}
	return role
}

func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		role, exists := ctx.Get("role")
		if !exists {
			utils.RespondError(ctx, types.CodeUnauthorized, "未授权")
			ctx.Abort()
			return
		}

		userRole, ok := role.(string)
		if !ok {
			utils.RespondError(ctx, types.CodeUnauthorized, "角色信息无效")
			ctx.Abort()
			return
		}
		userRole = NormalizeRole(userRole)

		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				ctx.Next()
				return
			}
		}

		utils.RespondError(ctx, types.CodeForbidden, "权限不足")
		ctx.Abort()
	}
}

func RequirePermission(permission string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		role, exists := ctx.Get("role")
		if !exists {
			ctx.JSON(http.StatusUnauthorized, types.ErrorResponse(types.CodeUnauthorized, "未授权"))
			ctx.Abort()
			return
		}

		userRole, ok := role.(string)
		if !ok {
			ctx.JSON(http.StatusForbidden, types.ErrorResponse(types.CodeForbidden, "角色信息无效"))
			ctx.Abort()
			return
		}
		userRole = NormalizeRole(userRole)

		permissions, roleExists := rolePermissions[userRole]
		if !roleExists {
			ctx.JSON(http.StatusForbidden, types.ErrorResponse(types.CodeForbidden, "角色不存在"))
			ctx.Abort()
			return
		}

		for _, p := range permissions {
			if p == permission {
				ctx.Next()
				return
			}
		}

		ctx.JSON(http.StatusForbidden, types.ErrorResponse(types.CodePermissionDenied, "权限不足"))
		ctx.Abort()
	}
}

func HasPermission(role string, permission string) bool {
	permissions, exists := rolePermissions[NormalizeRole(role)]
	if !exists {
		return false
	}
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

func GetAllRoles() []string {
	return []string{RoleSuperAdmin, RoleContentAdmin, RoleDataAdmin, RoleService}
}

func GetRoleName(role string) string {
	roleNames := map[string]string{
		RoleSuperAdmin:   "超级管理员",
		RoleContentAdmin: "内容管理员",
		RoleDataAdmin:    "数据管理员",
		RoleService:      "客服管理员",
	}
	if name, ok := roleNames[role]; ok {
		return name
	}
	return role
}
