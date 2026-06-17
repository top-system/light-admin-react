package service

import (
	stderrors "errors"

	"github.com/top-system/light-admin/api/member/repository"
	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/models/member"
	"github.com/top-system/light-admin/pkg/hash"
)

type MemberService struct {
	logger     lib.Logger
	memberRepo repository.MemberRepository
}

func NewMemberService(logger lib.Logger, memberRepo repository.MemberRepository) MemberService {
	return MemberService{logger: logger, memberRepo: memberRepo}
}

// Register 在指定租户内注册会员
func (a MemberService) Register(tenantID string, form *dto.MemberRegister) (*member.Member, error) {
	exists, err := a.memberRepo.ExistsUsername(tenantID, form.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.MemberAlreadyExists
	}

	hashed, err := hash.BcryptHash(form.Password)
	if err != nil {
		return nil, err
	}

	m := &member.Member{
		TenantID: tenantID, // 服务端强制注入，不信任客户端
		Username: form.Username,
		Password: hashed,
		Nickname: form.Nickname,
		Email:    form.Email,
		Mobile:   form.Mobile,
		Status:   1,
	}
	if err := a.memberRepo.Create(m); err != nil {
		return nil, err
	}
	return m.CleanSecure(), nil
}

// Verify 校验会员账号密码（登录用）
func (a MemberService) Verify(tenantID, username, password string) (*member.Member, error) {
	m, err := a.memberRepo.GetByUsername(tenantID, username)
	if err != nil {
		if stderrors.Is(err, errors.MemberRecordNotFound) {
			return nil, errors.MemberInvalidLogin // 不区分"不存在/密码错"，防枚举
		}
		return nil, err // 真实 DB 错误透传（→ 500），不伪装成 401
	}
	if !hash.BcryptCheck(password, m.Password) {
		return nil, errors.MemberInvalidLogin
	}
	if m.Status != 1 {
		return nil, errors.MemberIsDisabled
	}
	return m, nil
}

func (a MemberService) RecordLogin(tenantID, id, ip string) {
	if err := a.memberRepo.UpdateLoginInfo(tenantID, id, ip); err != nil {
		a.logger.Zap.Warnf("failed to record login for member %s: %v", id, err)
	}
}

func (a MemberService) GetProfile(tenantID, id string) (*member.Member, error) {
	m, err := a.memberRepo.GetByID(tenantID, id)
	if err != nil {
		return nil, err
	}
	return m.CleanSecure(), nil
}

func (a MemberService) UpdateProfile(tenantID, id string, form *member.MemberProfileForm) error {
	if _, err := a.memberRepo.GetByID(tenantID, id); err != nil {
		return err
	}
	return a.memberRepo.UpdateProfile(tenantID, id, form)
}

// ---- 后台管理 ----

func (a MemberService) Query(param *member.MemberQueryParam) (*member.MemberQueryResult, error) {
	return a.memberRepo.Query(param)
}

func (a MemberService) SetStatus(tenantID, id string, status int) error {
	return a.memberRepo.UpdateStatus(tenantID, id, status)
}

func (a MemberService) ResetPassword(tenantID, id, newPassword string) error {
	hashed, err := hash.BcryptHash(newPassword)
	if err != nil {
		return err
	}
	return a.memberRepo.UpdatePassword(tenantID, id, hashed)
}
