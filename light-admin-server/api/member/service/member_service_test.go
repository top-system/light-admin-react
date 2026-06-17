package service

import (
	stderrors "errors"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/top-system/light-admin/api/member/repository"
	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/models/member"
)

// newTestService 构造基于内存 sqlite 的真实 MemberRepository + MemberService。
// 被测路径（Register/Verify）不会解引用 logger，故传入零值 lib.Logger{}。
func newTestService(t *testing.T) (MemberService, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	assert.NoError(t, db.AutoMigrate(&member.Member{}))

	repo := repository.NewMemberRepository(lib.Database{ORM: db}, lib.Logger{})
	return NewMemberService(lib.Logger{}, repo), db
}

func TestRegister_Then_Verify_Succeeds(t *testing.T) {
	// Arrange
	svc, _ := newTestService(t)
	form := &dto.MemberRegister{Username: "alice", Password: "secret123", Nickname: "Alice"}

	// Act
	registered, err := svc.Register("tA", form)
	assert.NoError(t, err)
	assert.NotNil(t, registered)

	got, err := svc.Verify("tA", "alice", "secret123")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "alice", got.Username)
	assert.Equal(t, "tA", got.TenantID)
}

func TestRegister_DuplicateUsername_ReturnsAlreadyExists(t *testing.T) {
	// Arrange
	svc, _ := newTestService(t)
	form := &dto.MemberRegister{Username: "alice", Password: "secret123"}
	_, err := svc.Register("tA", form)
	assert.NoError(t, err)

	// Act
	_, err = svc.Register("tA", &dto.MemberRegister{Username: "alice", Password: "another"})

	// Assert
	assert.True(t, stderrors.Is(err, errors.MemberAlreadyExists))
}

func TestVerify_WrongPassword_ReturnsInvalidLogin(t *testing.T) {
	// Arrange
	svc, _ := newTestService(t)
	_, err := svc.Register("tA", &dto.MemberRegister{Username: "alice", Password: "secret123"})
	assert.NoError(t, err)

	// Act
	_, err = svc.Verify("tA", "alice", "wrongpass")

	// Assert
	assert.True(t, stderrors.Is(err, errors.MemberInvalidLogin))
}

func TestVerify_NonexistentUser_ReturnsInvalidLogin(t *testing.T) {
	// Arrange
	svc, _ := newTestService(t)

	// Act — 用户不存在也必须返回 InvalidLogin（防枚举），不得泄露 NotFound
	_, err := svc.Verify("tA", "ghost", "whatever")

	// Assert
	assert.True(t, stderrors.Is(err, errors.MemberInvalidLogin))
}

func TestVerify_DisabledMember_ReturnsDisabled(t *testing.T) {
	// Arrange
	svc, db := newTestService(t)
	registered, err := svc.Register("tA", &dto.MemberRegister{Username: "alice", Password: "secret123"})
	assert.NoError(t, err)

	// 直接通过 gorm 将该会员置为禁用
	err = db.Model(&member.Member{}).Where("id = ?", registered.ID).Update("status", 0).Error
	assert.NoError(t, err)

	// Act
	_, err = svc.Verify("tA", "alice", "secret123")

	// Assert
	assert.True(t, stderrors.Is(err, errors.MemberIsDisabled))
}
