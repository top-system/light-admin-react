package errors

import "net/http"

var (
	TenantNotFound     = New("tenant not found")
	TenantDisabled     = New("tenant is disabled")
	TenantCodeRequired = New("tenant code is required")
	TenantCodeExists   = New("tenant code already exists")

	MemberRecordNotFound = New("member record not found")
	MemberAlreadyExists  = New("member already exists")
	MemberInvalidLogin   = New("invalid username or password")
	MemberIsDisabled     = New("member is disabled")

	MemberTokenInvalid = New("member token is invalid")
	MemberTokenExpired = New("member token is expired")
)

func init() {
	RegisterHTTPStatus(TenantNotFound, http.StatusNotFound)
	RegisterHTTPStatus(TenantDisabled, http.StatusForbidden)
	RegisterHTTPStatus(TenantCodeRequired, http.StatusBadRequest)
	RegisterHTTPStatus(TenantCodeExists, http.StatusConflict)

	RegisterHTTPStatus(MemberRecordNotFound, http.StatusNotFound)
	RegisterHTTPStatus(MemberAlreadyExists, http.StatusConflict)
	RegisterHTTPStatus(MemberInvalidLogin, http.StatusUnauthorized)
	RegisterHTTPStatus(MemberIsDisabled, http.StatusForbidden)

	RegisterHTTPStatus(MemberTokenInvalid, http.StatusUnauthorized)
	RegisterHTTPStatus(MemberTokenExpired, http.StatusUnauthorized)
}
