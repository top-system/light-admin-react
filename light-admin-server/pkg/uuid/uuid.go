package uuid

import (
	"strings"

	"github.com/google/uuid"
)

// UUID Define alias
type UUID = uuid.UUID

// NewUUID Create uuid
func NewUUID() (UUID, error) {
	return uuid.NewRandom()
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
