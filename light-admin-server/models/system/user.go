package system

import (
	"github.com/top-system/light-admin/models/dto"
)

// User 用户模型
// Status: 1-正常 0-禁用
// Gender: 1-男 2-女 0-保密
type User struct {
	ID         string       `gorm:"primaryKey;type:char(32)" json:"id"`
	Username   string       `gorm:"column:username;size:64;index:idx_username" json:"username"`
	Nickname   string       `gorm:"column:nickname;size:64" json:"nickname"`
	Gender     int          `gorm:"column:gender;default:1" json:"gender"`
	Password   string       `gorm:"column:password;size:100" json:"password,omitempty"`
	DeptID     string       `gorm:"column:dept_id;index:idx_dept_id" json:"deptId"`
	Avatar     string       `gorm:"column:avatar;size:255" json:"avatar"`
	Mobile     string       `gorm:"column:mobile;size:20" json:"mobile"`
	Status     int          `gorm:"column:status;default:1;index:idx_user_status" json:"status"`
	Email      string       `gorm:"column:email;size:128" json:"email"`
	CreateTime dto.DateTime `gorm:"column:create_time;autoCreateTime;index:idx_user_create_time" json:"createTime"`
	CreateBy   string       `gorm:"column:create_by" json:"createBy"`
	UpdateTime dto.DateTime `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
	UpdateBy   string       `gorm:"column:update_by" json:"updateBy"`
	IsDeleted  int          `gorm:"column:is_deleted;default:0;index:idx_user_is_deleted" json:"isDeleted"`
	OpenID     string       `gorm:"column:openid;size:28" json:"openid,omitempty"`

	// 非数据库字段
	RoleIds  []string `gorm:"-" json:"roleIds,omitempty"`
	DeptName string   `gorm:"-" json:"deptName,omitempty"`
}

// TableName 指定表名
func (User) TableName() string {
	return "t_user"
}

type Users []*User

type UserInfo struct {
	ID       string `json:"userId"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Roles    Roles  `json:"roles"`
}

type UserQueryParam struct {
	dto.PaginationParam
	dto.OrderParam

	QueryPassword  bool
	Username       string   `query:"username"`
	Nickname       string   `query:"nickname"`
	QueryValue     string   `query:"query_value"`
	Keywords       string   `query:"keywords"`
	Status         *int     `query:"status"`
	DeptID         string   `query:"deptId"`
	RoleIDs        []string `query:"-"`
	CreateTimeFrom string   `query:"createTime[0]"`
	CreateTimeTo   string   `query:"createTime[1]"`
}

type UserQueryResult struct {
	List       Users           `json:"list"`
	Pagination *dto.Pagination `json:"pagination"`
}

func (a *User) CleanSecure() *User {
	a.Password = ""
	return a
}

func (a Users) ToIDs() []string {
	ids := make([]string, len(a))
	for i, item := range a {
		ids[i] = item.ID
	}
	return ids
}

// UserForm 用户表单
type UserForm struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Nickname string   `json:"nickname"`
	Mobile   string   `json:"mobile"`
	Gender   int      `json:"gender"`
	Avatar   string   `json:"avatar"`
	Email    string   `json:"email"`
	Status   int      `json:"status"`
	DeptId   string   `json:"deptId"`
	RoleIds  []string `json:"roleIds"`
}

// UserOption 用户下拉选项
type UserOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// ProfileForm 用户资料表单（用户自己更新资料）
type ProfileForm struct {
	Nickname string `json:"nickname"`
	Gender   int    `json:"gender"`
	Avatar   string `json:"avatar"`
	Mobile   string `json:"mobile"`
	Email    string `json:"email"`
}

// ChangePasswordForm 用户自助修改密码表单
type ChangePasswordForm struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

// ToOptions 转换为下拉选项列表
func (a Users) ToOptions() []*UserOption {
	options := make([]*UserOption, len(a))
	for i, user := range a {
		options[i] = &UserOption{
			Value: user.ID,
			Label: user.Nickname,
		}
	}
	return options
}
