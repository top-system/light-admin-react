package service

import (
	"errors"
	"fmt"
	"time"

	apperrors "github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/models/member"
)

type MemberAuthService struct {
	issuer    string
	codec     *lib.TokenCodec
	expired   int
	tokenType string
}

func NewMemberAuthService(config lib.Config) MemberAuthService {
	issuer := config.Name
	signingKey := []byte(fmt.Sprintf("Jwt:member:%s", issuer)) // 独立密钥
	expired := 7200
	if config.Auth != nil && config.Auth.TokenExpired > 0 {
		expired = config.Auth.TokenExpired
	}

	return MemberAuthService{
		issuer:    issuer,
		tokenType: "Bearer",
		expired:   expired,
		codec:     lib.NewHS512TokenCodec(signingKey),
	}
}

func (a MemberAuthService) GenerateToken(m *member.Member) (*dto.LoginResponse, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(a.expired) * time.Second)
	claims := &dto.MemberClaims{
		ID:       m.ID,
		TenantID: m.TenantID,
		Username: m.Username,
		RegisteredClaims: lib.RegisteredClaims{
			Issuer:    a.issuer,
			ExpiresAt: lib.NewNumericDate(expiresAt),
			IssuedAt:  lib.NewNumericDate(now),
			NotBefore: lib.NewNumericDate(now),
		},
	}

	accessToken, err := a.codec.Sign(claims)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken: accessToken,
		TokenType:   a.tokenType,
		ExpiresIn:   a.expired,
	}, nil
}

func (a MemberAuthService) ParseToken(tokenString string) (*dto.MemberClaims, error) {
	claims := &dto.MemberClaims{}
	if err := a.codec.Parse(tokenString, claims); err != nil {
		if errors.Is(err, lib.ErrTokenExpired) {
			return nil, apperrors.MemberTokenExpired
		}
		return nil, apperrors.MemberTokenInvalid
	}
	return claims, nil
}
