package tests

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// routePathRe 匹配前端路由表里的静态 path（front/src/router/index.tsx）。
var routePathRe = regexp.MustCompile(`path:\s*'([^']+)'`)

// gotoRe 匹配 e2e 用例里页面的直接跳转。
var gotoRe = regexp.MustCompile(`\.goto\(\s*'([^']+)'`)

// frontRouterDir 返回前端源码根目录（Go 测试的工作目录是 server/tests）。
func frontRouterDir() string {
	return filepath.Join("..", "..", "front")
}

// TestE2EGotoPathsExistInRouter 守住「e2e 用例跳转的页面必须真实注册」。
//
// 背景：后台用 `path: '*'` 兜底到 AdminNotFound，所以把用例写到不存在的
// /admin/adopts 也能正常渲染出一页 404，断言很容易误通过或被 test.skip 掩盖。
// 这里对 goto 目标做静态比对，任何未注册路径直接失败。
func TestE2EGotoPathsExistInRouter(t *testing.T) {
	root := frontRouterDir()

	raw, err := os.ReadFile(filepath.Join(root, "src", "router", "index.tsx"))
	if err != nil {
		t.Fatalf("读取前端路由表失败: %v", err)
	}

	var routes []string
	for _, m := range routePathRe.FindAllStringSubmatch(string(raw), -1) {
		p := m[1]
		// 通配路由（'*'、'/docs/*'）会命中任意路径，参与匹配等于放行一切
		if strings.Contains(p, "*") {
			continue
		}
		routes = append(routes, p)
	}
	if len(routes) == 0 {
		t.Fatal("未从前端路由表解析到任何 path，解析规则可能失效")
	}

	specFiles, err := filepath.Glob(filepath.Join(root, "e2e", "*.spec.ts"))
	if err != nil || len(specFiles) == 0 {
		t.Fatalf("未找到 e2e 用例文件: %v", err)
	}

	var failures []string
	for _, file := range specFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("读取 %s 失败: %v", file, err)
		}
		for _, m := range gotoRe.FindAllStringSubmatch(string(content), -1) {
			target := m[1]
			if idx := strings.IndexByte(target, '?'); idx >= 0 {
				target = target[:idx]
			}
			if strings.HasPrefix(target, "http") {
				continue
			}

			matched := false
			for _, route := range routes {
				if pathMatchesRoute(route, target) {
					matched = true
					break
				}
			}
			if !matched {
				failures = append(failures, filepath.Base(file)+" -> "+target)
			}
		}
	}

	if len(failures) > 0 {
		sort.Strings(failures)
		t.Errorf("e2e 跳转到未注册的前端路由（%d 处）：\n  %s", len(failures), strings.Join(failures, "\n  "))
	}
}
