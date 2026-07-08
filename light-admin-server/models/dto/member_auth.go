package dto

import (
	"github.com/top-system/light-admin/lib"
)

// MemberRegister 会员注册请求
type MemberRegister struct {
	Username string `json:"username" validate:"required,min=3,max=64"`
	Password string `json:"password" validate:"required,min=6,max=64"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Mobile   string `json:"mobile"`
}

// MemberLogin 会员登录请求
type MemberLogin struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// MemberClaims 会员 JWT claims（独立于后台 JwtClaims）
type MemberClaims struct {
	ID       string `json:"id"`
	TenantID string `json:"tenantId"`
	Username string `json:"username"`
	lib.RegisteredClaims
}
