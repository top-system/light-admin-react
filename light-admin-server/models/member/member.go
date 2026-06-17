package member

import "github.com/top-system/light-admin/models/dto"

// Member 会员模型（C 端用户，按 tenant_id 行级隔离）
// Status: 1-正常 0-禁用 ; Gender: 0-保密 1-男 2-女
type Member struct {
	ID            string       `gorm:"primaryKey;type:char(32)" json:"id"`
	TenantID      string       `gorm:"column:tenant_id;type:char(32);index:idx_member_tenant;uniqueIndex:uniq_tenant_username,priority:1;uniqueIndex:uniq_tenant_email,priority:1" json:"tenantId"`
	Username      string       `gorm:"column:username;size:64;uniqueIndex:uniq_tenant_username,priority:2" json:"username"`
	Email         string       `gorm:"column:email;size:128;uniqueIndex:uniq_tenant_email,priority:2" json:"email"`
	Mobile        string       `gorm:"column:mobile;size:20" json:"mobile"`
	Password      string       `gorm:"column:password;size:100" json:"-"` // json:"-" 是主要防护(永不序列化)；CleanSecure 仅作内存层二次清零
	Nickname      string       `gorm:"column:nickname;size:64" json:"nickname"`
	Avatar        string       `gorm:"column:avatar;size:255" json:"avatar"`
	Gender        int          `gorm:"column:gender;default:0" json:"gender"`
	Status        int          `gorm:"column:status;default:1" json:"status"`
	LastLoginTime dto.DateTime `gorm:"column:last_login_time" json:"lastLoginTime"`
	LastLoginIP   string       `gorm:"column:last_login_ip;size:64" json:"lastLoginIp"`
	CreateTime    dto.DateTime `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime    dto.DateTime `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
	IsDeleted     int          `gorm:"column:is_deleted;default:0" json:"isDeleted"`
}

func (Member) TableName() string { return "t_member" }

type Members []*Member

// MemberProfileForm 会员资料表单（会员本人更新）
type MemberProfileForm struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Gender   int    `json:"gender"`
	Mobile   string `json:"mobile"`
	Email    string `json:"email"`
}

// MemberQueryParam 后台会员查询参数
type MemberQueryParam struct {
	dto.PaginationParam
	dto.OrderParam
	TenantID string `query:"tenantId"`
	Keywords string `query:"keywords"`
	Status   *int   `query:"status"`
}

// MemberQueryResult 后台会员查询结果
type MemberQueryResult struct {
	List       Members         `json:"list"`
	Pagination *dto.Pagination `json:"pagination"`
}

func (a *Member) CleanSecure() *Member { a.Password = ""; return a }
