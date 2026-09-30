package errors

import "net/http"

// BusinessCode returns the stable response code for an error and HTTP status.
// Error identity takes priority over the generic HTTP status category.
func BusinessCode(err error, status int) string {
	if err != nil {
		for target, code := range domainCodes {
			if Is(err, target) {
				return code
			}
		}
	}
	switch status {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted, http.StatusNoContent:
		return "00000"
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return "A0400"
	case http.StatusUnauthorized:
		return "A0401"
	case http.StatusForbidden:
		return "A0403"
	case http.StatusNotFound:
		return "A0404"
	case http.StatusMethodNotAllowed:
		return "A0405"
	case http.StatusConflict:
		return "A0409"
	case http.StatusTooManyRequests:
		return "A0429"
	default:
		if status >= http.StatusInternalServerError {
			return "C0500"
		}
		if status >= http.StatusBadRequest {
			return "A0400"
		}
		return "00000"
	}
}

var domainCodes = map[error]string{
	UserAlreadyExists:           "B1001",
	UserInvalidPassword:         "B1002",
	UserIsDisable:               "B1003",
	UserOldPasswordWrong:        "B1004",
	UserPasswordSame:            "B1005",
	RoleAlreadyExists:           "B1101",
	RoleCodeAlreadyExists:       "B1102",
	RoleNotAllowDeleteWithUser:  "B1103",
	MenuAlreadyExists:           "B1201",
	MenuInvalidParent:           "B1202",
	MenuNotAllowDeleteWithChild: "B1203",
	ConfigKeyAlreadyExists:      "B1301",
	TenantCodeExists:            "B1401",
	TenantDisabled:              "B1402",
	MemberAlreadyExists:         "B1501",
	MemberInvalidLogin:          "B1502",
	MemberIsDisabled:            "B1503",
	CaptchaAnswerCodeNoMatch:    "B1601",
	AuthTokenExpired:            "A0411",
	MemberTokenExpired:          "A0412",
}
