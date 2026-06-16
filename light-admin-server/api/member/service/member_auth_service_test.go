package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/member"
)

func newTestAuth() MemberAuthService {
	cfg := lib.Config{Name: "test", Auth: &lib.AuthConfig{TokenExpired: 7200}}
	return NewMemberAuthService(cfg)
}

func TestMemberTokenRoundTrip(t *testing.T) {
	svc := newTestAuth()
	m := &member.Member{ID: "m1", TenantID: "tA", Username: "alice"}

	resp, err := svc.GenerateToken(m)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)

	claims, err := svc.ParseToken(resp.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, "m1", claims.ID)
	assert.Equal(t, "tA", claims.TenantID)
	assert.Equal(t, "alice", claims.Username)
}

func TestMemberParseInvalidToken(t *testing.T) {
	svc := newTestAuth()
	_, err := svc.ParseToken("not-a-token")
	assert.Error(t, err)
}
