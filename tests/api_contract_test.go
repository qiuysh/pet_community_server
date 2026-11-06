package tests

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"pet-community/controller"
	"pet-community/router"

	"github.com/gin-gonic/gin"
)

// methodCallRe 定位前端 api 层的请求调用起点（get/post/put/del/patch）。
// 泛型参数用 [^(]* 限住，避免跨过参数列表误配下一个调用。
var methodCallRe = regexp.MustCompile(`\b(get|post|put|del|patch)\s*(?:<[^(]*>)?\s*\(`)

// templateParamRe 把路径里的 ${...} 归一化为占位符，便于与 gin 的 :param 对齐。
var templateParamRe = regexp.MustCompile(`\$\{[^}]*\}`)

// httpMethodOf 前端封装名 → HTTP 方法。
var httpMethodOf = map[string]string{
	"get":   "GET",
	"post":  "POST",
	"put":   "PUT",
	"del":   "DELETE",
	"patch": "PATCH",
}

// apiCall 表示前端发起的一次请求（路径已归一化）。
type apiCall struct {
	file   string
	method string
	path   string
}

// collectBackendRoutes 通过唯一入口 router.RegisterRoutes 拿到后端真实路由表。
//
// 传零值 Handlers：gin 在注册时只保存 handler 的方法值，不会调用它，
// 所以不需要数据库、配置或 service 实例。
func collectBackendRoutes(t *testing.T) []string {
	t.Helper()

	gin.SetMode(gin.TestMode)
	engine := gin.New()

	// nil controller 取方法值在 Go 里是合法的（只有调用时才需要非 nil receiver），
	// 但注册过程中若被意外调用会 panic，这里转成明确的失败信息。
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("注册路由时 panic，Handlers 结构可能被误用: %v", r)
		}
	}()

	router.RegisterRoutes(engine, router.Handlers{
		Admin: router.AdminControllers{
			Auth:         (*controller.AdminAuthController)(nil),
			Users:        (*controller.AdminUsersController)(nil),
			FrontUsers:   (*controller.AdminFrontUsersController)(nil),
			Settings:     (*controller.AdminSettingsController)(nil),
			Logs:         (*controller.AdminLogsController)(nil),
			Review:       (*controller.AdminReviewController)(nil),
			Report:       (*controller.AdminReportController)(nil),
			Stats:        (*controller.AdminStatsController)(nil),
			Dashboard:    (*controller.AdminDashboardController)(nil),
			Notification: (*controller.AdminNotificationController)(nil),
			Announcement: (*controller.AdminAnnouncementController)(nil),
			Ad:           (*controller.AdminAdController)(nil),
		},
		User:                 (*controller.UserController)(nil),
		Pet:                  (*controller.PetController)(nil),
		Post:                 (*controller.PostController)(nil),
		Notification:         (*controller.NotificationController)(nil),
		Follow:               (*controller.FollowController)(nil),
		Block:                (*controller.BlockController)(nil),
		Message:              (*controller.MessageController)(nil),
		Points:               (*controller.PointsController)(nil),
		NotificationSettings: (*controller.NotificationSettingsController)(nil),
		Discovery:            (*controller.DiscoveryController)(nil),
		Activity:             (*controller.ActivityController)(nil),
		Adopt:                (*controller.AdoptController)(nil),
		FrontAd:              (*controller.FrontAdController)(nil),
		Site:                 (*controller.SiteController)(nil),
		Upload:               (*controller.UploadController)(nil),
	})

	routes := engine.Routes()
	if len(routes) == 0 {
		t.Fatal("未注册到任何路由，RegisterRoutes 可能被改坏")
	}

	out := make([]string, 0, len(routes))
	for _, r := range routes {
		out = append(out, r.Method+" "+r.Path)
	}
	return out
}

// parseStringLiteral 从 s[offset:] 起读一个 JS 字符串字面量。
// 返回字面量内容与是否成功（调用参数不是字面量时返回 false）。
func parseStringLiteral(s string, offset int) (string, bool) {
	i := offset
	for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') {
		i++
	}
	if i >= len(s) {
		return "", false
	}
	quote := s[i]
	if quote != '\'' && quote != '"' && quote != '`' {
		return "", false
	}
	start := i + 1
	for j := start; j < len(s); j++ {
		if s[j] == '\\' {
			j++
			continue
		}
		if s[j] == quote {
			return s[start:j], true
		}
	}
	return "", false
}

// collectFrontendCalls 解析 front/src/api/*.ts，提取「方法 + 归一化路径」。
func collectFrontendCalls(t *testing.T) []apiCall {
	t.Helper()

	dir := filepath.Join("..", "..", "front", "src", "api")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("读取前端 api 目录 %s 失败: %v", dir, err)
	}

	var calls []apiCall
	for _, e := range entries {
		name := e.Name()
		// axios.ts 是封装本身（get(url, config)），不发起具体路径；
		// 测试文件里的路径属于夹具，不参与契约校验。
		if e.IsDir() || !strings.HasSuffix(name, ".ts") || name == "axios.ts" || strings.HasSuffix(name, ".test.ts") {
			continue
		}

		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("读取 %s 失败: %v", name, err)
		}
		src := string(raw)

		for _, m := range methodCallRe.FindAllStringSubmatchIndex(src, -1) {
			fn := src[m[2]:m[3]]
			literal, ok := parseStringLiteral(src, m[1])
			if !ok || !strings.HasPrefix(literal, "/api/v1/") {
				continue // 参数不是 /api/v1 字面量（动态拼接或非 REST 调用），交给人工审查
			}

			path := literal
			if idx := strings.IndexByte(path, '?'); idx >= 0 {
				path = path[:idx] // query 不参与路由匹配
			}
			path = templateParamRe.ReplaceAllString(path, ":param")

			calls = append(calls, apiCall{file: name, method: httpMethodOf[fn], path: path})
		}
	}

	if len(calls) == 0 {
		t.Fatal("未从前端 api 层解析到任何请求，解析规则可能失效")
	}
	return calls
}

// splitSegments 按 / 切分路径，忽略空段。
func splitSegments(p string) []string {
	parts := strings.Split(p, "/")
	out := make([]string, 0, len(parts))
	for _, s := range parts {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// pathMatchesRoute 判断前端请求路径是否命中某条注册路由的路径模板。
// gin 的通配段（:id、*filepath）视为任意值。
func pathMatchesRoute(routePath, reqPath string) bool {
	r := splitSegments(routePath)
	q := splitSegments(reqPath)

	for i, seg := range r {
		if strings.HasPrefix(seg, "*") {
			return true // catch-all 之后不再比较
		}
		if i >= len(q) {
			return false
		}
		if strings.HasPrefix(seg, ":") {
			continue
		}
		if seg != q[i] {
			return false
		}
	}
	return len(r) == len(q)
}

// TestFrontendAPICallsMatchBackendRoutes 守住「前端 api 封装必须打到已注册路由」。
//
// 背景：api/*.ts 曾长期保留后端从未注册的封装（如话题绑定 openID、活动统计），
// 调用即 404；这类问题在类型检查和单测里都看不出来，只有用户点下去才暴露。
// 这里把路由表与 api 层做静态比对，任一调用找不到对应路由（或方法不一致）即失败。
func TestFrontendAPICallsMatchBackendRoutes(t *testing.T) {
	routes := collectBackendRoutes(t)
	calls := collectFrontendCalls(t)

	// 后端路由按路径索引，便于给出「路径存在但方法不对」的提示
	byPath := make(map[string][]string, len(routes))
	for _, route := range routes {
		method, path, ok := strings.Cut(route, " ")
		if !ok {
			continue
		}
		byPath[path] = append(byPath[path], method)
	}

	var failures []string
	for _, call := range calls {
		var matchedPath string
		allowed := []string(nil)
		for routePath, methods := range byPath {
			if pathMatchesRoute(routePath, call.path) {
				for _, m := range methods {
					if m == call.method {
						matchedPath = routePath
						break
					}
				}
				if matchedPath != "" {
					break
				}
				if allowed == nil {
					allowed = methods
					matchedPath = ""
				}
			}
		}
		if matchedPath != "" {
			continue
		}
		if len(allowed) > 0 {
			sort.Strings(allowed)
			failures = append(failures, call.file+" "+call.method+" "+call.path+
				" —— 路径存在但方法不匹配，后端支持: "+strings.Join(allowed, ", "))
			continue
		}
		failures = append(failures, call.file+" "+call.method+" "+call.path+" —— 后端无此路由")
	}

	if len(failures) > 0 {
		sort.Strings(failures)
		t.Errorf("前端 api 封装与后端路由表不一致（%d 处）：\n  %s", len(failures), strings.Join(failures, "\n  "))
	}
}
