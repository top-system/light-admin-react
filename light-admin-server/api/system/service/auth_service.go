package service

import (
	"errors"
	"fmt"
	"time"

	apperrors "github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/models/system"
)

type AuthService struct {
	codec     *lib.TokenCodec
	cache     lib.Cache
	expired   int
	tokenType string
}

func NewAuthService(cache lib.Cache, config lib.Config) AuthService {
	signingKey := fmt.Sprintf("Jwt:%s", config.Name)

	return AuthService{
		cache:     cache,
		codec:     lib.NewHS512TokenCodec([]byte(signingKey)),
		expired:   config.Auth.TokenExpired,
		tokenType: "Bearer",
	}
}

func wrapperAuthKey(key string) string {
	return fmt.Sprintf("auth:%s", key)
}

func (a AuthService) GenerateToken(user *system.User) (*dto.LoginResponse, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(a.expired) * time.Second)
	claims := &dto.JwtClaims{
		ID:       user.ID,
		Username: user.Username,
		RegisteredClaims: lib.RegisteredClaims{
			ExpiresAt: lib.NewNumericDate(expiresAt),
			IssuedAt:  lib.NewNumericDate(now),
			NotBefore: lib.NewNumericDate(now),
		},
	}

	err := a.cache.Set(wrapperAuthKey(claims.Username), 1, time.Until(expiresAt))
	if err != nil {
		return nil, err
	}

	accessToken, err := a.codec.Sign(claims)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: "", // 暂未实现刷新令牌
		TokenType:    a.tokenType,
		ExpiresIn:    a.expired,
	}, nil
}

func (a AuthService) ParseToken(tokenString string) (*dto.JwtClaims, error) {
	claims := &dto.JwtClaims{}
	if err := a.codec.Parse(tokenString, claims); err != nil {
		switch {
		case errors.Is(err, lib.ErrTokenMalformed):
			return nil, apperrors.AuthTokenMalformed
		case errors.Is(err, lib.ErrTokenExpired):
			return nil, apperrors.AuthTokenExpired
		case errors.Is(err, lib.ErrTokenNotValidYet):
			return nil, apperrors.AuthTokenNotValidYet
		default:
			return nil, apperrors.AuthTokenInvalid
		}
	}
	return claims, nil
}

func (a AuthService) DestroyToken(username string) error {
	_, err := a.cache.Delete(wrapperAuthKey(username))
	return err
}
