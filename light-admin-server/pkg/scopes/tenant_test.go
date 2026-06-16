package scopes

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type memberRow struct {
	ID       string `gorm:"primaryKey"`
	TenantID string
	Name     string
}

func (memberRow) TableName() string { return "members" }

func setupDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	assert.NoError(t, db.AutoMigrate(&memberRow{}))
	db.Create(&memberRow{ID: "1", TenantID: "tA", Name: "alice"})
	db.Create(&memberRow{ID: "2", TenantID: "tB", Name: "bob"})
	return db
}

func TestTenantScopeFiltersByTenant(t *testing.T) {
	db := setupDB(t)

	var rows []memberRow
	err := db.Model(&memberRow{}).Scopes(Tenant("tA")).Find(&rows).Error

	assert.NoError(t, err)
	assert.Len(t, rows, 1)
	assert.Equal(t, "alice", rows[0].Name)
}

func TestTenantScopeEmptyTenantReturnsNone(t *testing.T) {
	db := setupDB(t)

	var rows []memberRow
	err := db.Model(&memberRow{}).Scopes(Tenant("")).Find(&rows).Error

	assert.NoError(t, err)
	assert.Len(t, rows, 0) // 空租户 ID 必须查不到任何数据，防止越权
}
