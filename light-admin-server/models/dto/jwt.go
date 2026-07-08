package dto

import (
	"github.com/top-system/light-admin/lib"
)

type JwtClaims struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	lib.RegisteredClaims
}
