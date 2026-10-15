package controller

import "strconv"

// parseInt 解析字符串为 int,失败返回 error。
// 从原 admin_controller.go 的各处内联 strconv.Atoi 集中。
func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}
