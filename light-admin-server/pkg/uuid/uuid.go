package uuid

import (
	"strings"

	"github.com/google/uuid"
)

// UUID Define alias
type UUID = uuid.UUID

// Nil is the zero UUID.
var Nil = uuid.Nil

// NewUUID Create uuid
func NewUUID() (UUID, error) {
	return uuid.NewRandom()
}

// NewV4 returns a random (version 4) UUID. Kept as the single entry point so
// callers don't import a uuid library directly (see infra-abstraction-plan §5).
func NewV4() (UUID, error) {
	return uuid.NewRandom()
}

// Must returns u, panicking if err is non-nil. Mirrors the underlying library's
// Must so `Must(NewV4())` reads naturally.
func Must(u UUID, err error) UUID {
	return uuid.Must(u, err)
}

// FromStringOrNil parses s, returning Nil if it is not a valid UUID.
func FromStringOrNil(s string) UUID {
	u, err := uuid.Parse(s)
	if err != nil {
		return Nil
	}
	return u
}

// MustUUID Create uuid(Throw panic if something goes wrong)
func MustUUID() UUID {
	v, err := NewUUID()
	if err != nil {
		panic(err)
	}
	return v
}

// MustString Create uuid
func MustString() string {
	return MustUUID().String()
}

// NewID 生成 32 位无连字符的十六进制 UUID 字符串，用作实体主键
func NewID() string {
	return strings.ReplaceAll(MustUUID().String(), "-", "")
}
