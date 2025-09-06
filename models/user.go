package models

import (
	"database/sql/driver"
	"time"

	"gorm.io/gorm"
)

type UserRole string

const (
	UserRoleEndUser   UserRole = "终端用户"
	UserRoleAreaMgr   UserRole = "区域管理员"
	UserRoleCityAdmin UserRole = "总管理员"
)

func (r UserRole) String() string {
	return string(r)
}

func (r *UserRole) Value() (driver.Value, error) {
	return string(*r), nil
}

func (r *UserRole) Scan(value interface{}) error {
	*r = UserRole(value.(string))
	return nil
}

type User struct {
	ID       int    `gorm:"column:id;primaryKey;autoIncrement;comment:用户ID"`
	Username string `gorm:"column:username;not null;comment:用户名"`
	Phone    string `gorm:"column:phone;not null;comment:手机号"`

	Role UserRole `gorm:"column:role;not null;comment:角色"`

	CreatedAt time.Time `gorm:"column:created_at;not null;comment:创建时间"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;comment:更新时间"`
}

func (User) TableName() string {
	return GetTableNames("user")
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
	u.UpdatedAt = time.Now()
	return nil
}
