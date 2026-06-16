package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/top-system/light-admin/models/member"
	"github.com/top-system/light-admin/pkg/hash"
)

func TestVerifyPasswordMatch(t *testing.T) {
	hashed, _ := hash.BcryptHash("secret123")
	m := &member.Member{Password: hashed, Status: 1}

	assert.True(t, hash.BcryptCheck("secret123", m.Password))
	assert.False(t, hash.BcryptCheck("wrong", m.Password))
}
