package lib

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// This file is the single sanctioned entry point for golang-jwt (see
// infra-abstraction-plan §5). golang-jwt has changed organisations twice in its
// history, so both auth services and the claim structs go through this facade
// instead of importing the library directly — swapping it out means touching one
// file, and the depguard boundary keeps it that way.

// Re-exports let claim structs embed the registered-claims payload and build
// numeric dates without importing golang-jwt themselves.
type (
	// Claims is any JWT payload that can be signed and parsed.
	Claims = jwt.Claims
	// RegisteredClaims is the standard set of JWT registered claims to embed.
	RegisteredClaims = jwt.RegisteredClaims
	// NumericDate is a JWT timestamp.
	NumericDate = jwt.NumericDate
)

// NewNumericDate wraps a time.Time as a JWT NumericDate.
func NewNumericDate(t time.Time) *NumericDate { return jwt.NewNumericDate(t) }

// Token parse failures, mapped from golang-jwt's sentinel errors so callers can
// switch on them with errors.Is without importing golang-jwt.
var (
	ErrTokenMalformed   = errors.New("token is malformed")
	ErrTokenExpired     = errors.New("token is expired")
	ErrTokenNotValidYet = errors.New("token is not valid yet")
	ErrTokenInvalid     = errors.New("token is invalid")
)

// TokenCodec signs and parses JWTs with a fixed HMAC signing method and key.
type TokenCodec struct {
	method  jwt.SigningMethod
	key     []byte
	keyfunc jwt.Keyfunc
}

// NewHS512TokenCodec returns a codec that signs with HS512 using secret.
func NewHS512TokenCodec(secret []byte) *TokenCodec {
	c := &TokenCodec{method: jwt.SigningMethodHS512, key: secret}
	c.keyfunc = func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return secret, nil
	}
	return c
}

// Sign serialises claims into a signed token string.
func (c *TokenCodec) Sign(claims Claims) (string, error) {
	return jwt.NewWithClaims(c.method, claims).SignedString(c.key)
}

// Parse validates tokenString and populates claims (which must be a pointer). On
// failure it returns one of the Err* sentinels above.
func (c *TokenCodec) Parse(tokenString string, claims Claims) error {
	token, err := jwt.ParseWithClaims(tokenString, claims, c.keyfunc)
	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenMalformed):
			return ErrTokenMalformed
		case errors.Is(err, jwt.ErrTokenExpired):
			return ErrTokenExpired
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			return ErrTokenNotValidYet
		default:
			return ErrTokenInvalid
		}
	}
	if !token.Valid {
		return ErrTokenInvalid
	}
	return nil
}
