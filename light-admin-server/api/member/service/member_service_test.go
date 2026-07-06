package service

import (
	"context"
	stderrors "errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"

	"github.com/top-system/light-admin/api/member/repository"
	"github.com/top-system/light-admin/db/sqlc"
	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
)

// memberStoreQuerier is a tiny in-memory stand-in for the member subset of
// sqlc.Querier. It keeps the Register/Verify/SetStatus paths honest without a
// real database, replacing the previous in-memory sqlite GORM fixture.
type memberStoreQuerier struct {
	sqlc.Querier
	rows []sqlc.TMember
}

func (m *memberStoreQuerier) CountMembersByUsername(_ context.Context, arg sqlc.CountMembersByUsernameParams) (int64, error) {
	var n int64
	for _, r := range m.rows {
		if r.TenantID == arg.TenantID && r.Username == arg.Username && r.IsDeleted == 0 {
			n++
		}
	}
	return n, nil
}

func (m *memberStoreQuerier) CreateMember(_ context.Context, arg sqlc.CreateMemberParams) error {
	m.rows = append(m.rows, sqlc.TMember{
		ID:        arg.ID,
		TenantID:  arg.TenantID,
		Username:  arg.Username,
		Email:     arg.Email,
		Mobile:    arg.Mobile,
		Password:  arg.Password,
		Nickname:  arg.Nickname,
		Avatar:    arg.Avatar,
		Gender:    arg.Gender,
		Status:    arg.Status,
		IsDeleted: arg.IsDeleted,
	})
	return nil
}

func (m *memberStoreQuerier) GetMemberByUsername(_ context.Context, arg sqlc.GetMemberByUsernameParams) (sqlc.TMember, error) {
	for _, r := range m.rows {
		if r.TenantID == arg.TenantID && r.Username == arg.Username && r.IsDeleted == 0 {
			return r, nil
		}
	}
	return sqlc.TMember{}, pgx.ErrNoRows
}

func (m *memberStoreQuerier) UpdateMemberStatus(_ context.Context, arg sqlc.UpdateMemberStatusParams) error {
	for i := range m.rows {
		if m.rows[i].TenantID == arg.TenantID && m.rows[i].ID == arg.ID {
			m.rows[i].Status = arg.Status
		}
	}
	return nil
}

// newTestService 构造基于内存 mock Querier 的真实 MemberRepository + MemberService。
func newTestService(t *testing.T) (MemberService, *memberStoreQuerier) {
	t.Helper()
	store := &memberStoreQuerier{}
	repo := repository.NewMemberRepositoryWithQuerier(store, lib.Logger{})
	return NewMemberService(lib.Logger{}, repo), store
}

func TestRegister_Then_Verify_Succeeds(t *testing.T) {
	svc, _ := newTestService(t)
	form := &dto.MemberRegister{Username: "alice", Password: "secret123", Nickname: "Alice"}

	registered, err := svc.Register("tA", form)
	assert.NoError(t, err)
	assert.NotNil(t, registered)

	got, err := svc.Verify("tA", "alice", "secret123")
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "alice", got.Username)
	assert.Equal(t, "tA", got.TenantID)
}

func TestRegister_DuplicateUsername_ReturnsAlreadyExists(t *testing.T) {
	svc, _ := newTestService(t)
	form := &dto.MemberRegister{Username: "alice", Password: "secret123"}
	_, err := svc.Register("tA", form)
	assert.NoError(t, err)

	_, err = svc.Register("tA", &dto.MemberRegister{Username: "alice", Password: "another"})
	assert.True(t, stderrors.Is(err, errors.MemberAlreadyExists))
}

func TestVerify_WrongPassword_ReturnsInvalidLogin(t *testing.T) {
	svc, _ := newTestService(t)
	_, err := svc.Register("tA", &dto.MemberRegister{Username: "alice", Password: "secret123"})
	assert.NoError(t, err)

	_, err = svc.Verify("tA", "alice", "wrongpass")
	assert.True(t, stderrors.Is(err, errors.MemberInvalidLogin))
}

func TestVerify_NonexistentUser_ReturnsInvalidLogin(t *testing.T) {
	svc, _ := newTestService(t)

	// 用户不存在也必须返回 InvalidLogin（防枚举），不得泄露 NotFound
	_, err := svc.Verify("tA", "ghost", "whatever")
	assert.True(t, stderrors.Is(err, errors.MemberInvalidLogin))
}

func TestVerify_DisabledMember_ReturnsDisabled(t *testing.T) {
	svc, _ := newTestService(t)
	registered, err := svc.Register("tA", &dto.MemberRegister{Username: "alice", Password: "secret123"})
	assert.NoError(t, err)

	// 将该会员置为禁用
	assert.NoError(t, svc.SetStatus("tA", registered.ID, 0))

	_, err = svc.Verify("tA", "alice", "secret123")
	assert.True(t, stderrors.Is(err, errors.MemberIsDisabled))
}
