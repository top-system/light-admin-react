package scopes

import "gorm.io/gorm"

// Tenant 返回一个按 tenant_id 过滤的 GORM Scope。
// tenantID 为空时强制匹配空字符串（查不到任何数据），防止越权读取全表。
func Tenant(tenantID string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("tenant_id = ?", tenantID)
	}
}
