package constants

// version
const Version = "1.0"

// humanized time format
const TimeFormat = "2006-01-02 15:04:05"

// captcha store key prefix
const CaptchaKeyPrefix = "captcha"
const CaptchaExpireTimes = 90

// echo
const CurrentUser = "current-user"
const RoutesCacheKey = "routes"

// 多租户 / 会员 上下文 key
const CurrentTenantID = "current-tenant-id" // 解析出的租户 ID
const CurrentMember = "current-member"      // 会员 JWT claims

// RedisDB
const RedisMainDB = 0
const RedisTaskDB = 1