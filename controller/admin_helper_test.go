package controller

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// TestParseTimeRangeDateGranularity 守住「后台日期筛选真的生效」。
//
// 后台日志页的日期控件是 granularity='day'，提交的是 YYYY-MM-DD；
// 而历史实现只认 "2006-01-02 15:04:05"，解析失败被 `_ =` 吞掉后返回零值，
// 于是「按时间筛选」静默变成「不限时间」，页面显示全量数据。
func TestParseTimeRangeDateGranularity(t *testing.T) {
	start, end, err := ParseTimeRange("2026-11-05", "2026-11-07")
	if err != nil {
		t.Fatalf("日期格式应被接受，实际报错: %v", err)
	}

	wantStart := time.Date(2026, 11, 5, 0, 0, 0, 0, time.Local)
	if !start.Equal(wantStart) {
		t.Errorf("开始时间 = %v，期望 %v", start, wantStart)
	}

	// 结束日期代表「截止到这一天结束」，仓储层是 created_at <= endTime，
	// 不补齐到 23:59:59 会漏掉最后一天的全部记录
	wantEnd := time.Date(2026, 11, 7, 23, 59, 59, 0, time.Local)
	if !end.Equal(wantEnd) {
		t.Errorf("结束时间 = %v，期望 %v", end, wantEnd)
	}
}

// TestParseTimeRangeDateTimeGranularity 保证带时分秒的值不被截断。
func TestParseTimeRangeDateTimeGranularity(t *testing.T) {
	start, end, err := ParseTimeRange("2026-11-05 08:30:00", "2026-11-05 18:00:00")
	if err != nil {
		t.Fatalf("日期时间格式应被接受，实际报错: %v", err)
	}

	wantStart := time.Date(2026, 11, 5, 8, 30, 0, 0, time.Local)
	if !start.Equal(wantStart) {
		t.Errorf("开始时间 = %v，期望 %v", start, wantStart)
	}

	// 精确到秒的值不应被补齐到当天末尾
	wantEnd := time.Date(2026, 11, 5, 18, 0, 0, 0, time.Local)
	if !end.Equal(wantEnd) {
		t.Errorf("结束时间 = %v，期望 %v（不应补齐）", end, wantEnd)
	}
}

// TestParseTimeRangeInvalidReturnsError 非法时间必须报错，而不是静默当作「不限时间」。
func TestParseTimeRangeInvalidReturnsError(t *testing.T) {
	cases := []struct{ start, end string }{
		{"2026/11/05", ""},
		{"", "not-a-date"},
		{"2026-13-45", ""},
	}

	for _, c := range cases {
		if _, _, err := ParseTimeRange(c.start, c.end); err == nil {
			t.Errorf("ParseTimeRange(%q, %q) 应返回错误，实际通过", c.start, c.end)
		}
	}
}

// TestParseTimeRangeEmptyMeansNoFilter 空值仍表示「不筛选」，不能被当成错误。
func TestParseTimeRangeEmptyMeansNoFilter(t *testing.T) {
	start, end, err := ParseTimeRange("", "")
	if err != nil {
		t.Fatalf("空参数不应报错: %v", err)
	}
	if !start.IsZero() || !end.IsZero() {
		t.Errorf("空参数应返回零值时间，实际 start=%v end=%v", start, end)
	}
}

// TestParseIntQueryDefault 非法状态参数必须报错。
//
// status=abc 曾因 `_ =` 变成 0，而 0 是「待审核」的有效值 —— 用户传错参数
// 看到的是被悄悄换成「待审核」的列表，比报错更难排查。
func TestParseIntQueryDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)

	newCtx := func(raw string) *gin.Context {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/?status="+raw, nil)
		return c
	}

	if v, err := ParseIntQueryDefault(newCtx(""), "status", -1); err != nil || v != -1 {
		t.Errorf("空值应回默认值 -1，实际 v=%d err=%v", v, err)
	}
	if v, err := ParseIntQueryDefault(newCtx("2"), "status", -1); err != nil || v != 2 {
		t.Errorf("合法值应解析为 2，实际 v=%d err=%v", v, err)
	}
	if _, err := ParseIntQueryDefault(newCtx("abc"), "status", -1); err == nil {
		t.Error("非法值应返回错误，实际通过")
	}
}
