package cache

import "fmt"

const (
	codeTTLSeconds     = 5 * 60
	sendMinuteLimit    = 1
	sendHourLimit      = 10
	loginFailLimit     = 5
	loginLockSeconds   = 30 * 60
	sessionTVTTL       = 24 * 60 * 60
	sessionDeviceTTL   = 7 * 24 * 60 * 60

	rankingTTLSeconds    = 5 * 60
	hotKeywordTTLSeconds = 2 * 60
	activeAdsTTLSeconds  = 2 * 60
)

func keyVerifyCode(account, codeType string) string {
	return fmt.Sprintf("vc:code:%s:%s", codeType, account)
}

func keySendMinute(account, ip string) string {
	return fmt.Sprintf("vc:rl:1m:%s:%s", account, ip)
}

func keySendHour(account, ip string) string {
	return fmt.Sprintf("vc:rl:1h:%s:%s", account, ip)
}

func keyLoginFail(scope string, id uint) string {
	return fmt.Sprintf("login:fail:%s:%d", scope, id)
}

func keyLoginLock(scope string, id uint) string {
	return fmt.Sprintf("login:lock:%s:%d", scope, id)
}

func keyTokenVersion(userID uint) string {
	return fmt.Sprintf("auth:tv:%d", userID)
}

func keyDevice(userID uint, deviceID string) string {
	return fmt.Sprintf("auth:dev:%d:%s", userID, deviceID)
}

func keyRateIPRoute(ip, route string) string {
	return fmt.Sprintf("rl:ip_route:%s:%s", ip, route)
}

func keyRateIPGlobal(ip string) string {
	return fmt.Sprintf("rl:ip:%s", ip)
}

func keyRateUserRoute(userID uint, route string) string {
	return fmt.Sprintf("rl:user_route:%d:%s", userID, route)
}

func keyRanking(period string) string {
	return fmt.Sprintf("rank:all:%s", period)
}

func keyHotKeywords(limit int) string {
	return fmt.Sprintf("hot:keywords:top:%d", limit)
}

const keyHotKeywordZSet = "hot:keywords:z"

func keyActiveAds(location string) string {
	if location == "" {
		location = "_all_"
	}
	return fmt.Sprintf("ads:active:%s", location)
}

const keyActiveAdsPattern = "ads:active:*"
