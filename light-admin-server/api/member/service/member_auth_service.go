package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	apperrors "github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/models/member"
)

type MemberAuthService struct {
	issuer        string
	signingMethod jwt.SigningMethod
	signingKey    []byte
	keyfunc       jwt.Keyfunc
	expired       int
	tokenType     string
}

func NewMemberAuthService(config lib.Config) MemberAuthService {
	issuer := config.Name
	signingKey := []byte(fmt.Sprintf("Jwt:member:%s", issuer)) // 独立密钥
	expired := 7200
	if config.Auth != nil && config.Auth.TokenExpired > 0 {
		expired = config.Auth.TokenExpired
	}

	return MemberAuthService{
		issuer:        issuer,
		tokenType:     "Bearer",
		expired:       expired,
		signingMethod: jwt.SigningMethodHS512,
		signingKey:    signingKey,
		keyfunc: func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, apperrors.MemberTokenInvalid
			}
			return signingKey, nil
		},
	}
}

func (a MemberAuthService) GenerateToken(m *member.Member) (*dto.LoginResponse, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(a.expired) * time.Second)
	claims := &dto.MemberClaims{
		ID:       m.ID,
		TenantID: m.TenantID,
		Username: m.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    a.issuer,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(a.signingMethod, claims)
	accessToken, err := token.SignedString(a.signingKey)
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
	token, err := jwt.ParseWithClaims(tokenString, &dto.MemberClaims{}, a.keyfunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, apperrors.MemberTokenExpired
		}
		return nil, apperrors.MemberTokenInvalid
	}
	if claims, ok := token.Claims.(*dto.MemberClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, apperrors.MemberTokenInvalid
}
