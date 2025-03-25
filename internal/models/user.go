package models
import (
	"gorm.io/gorm"
	"time"
)
type UserRole string

const (
    Admin          UserRole = "ADMIN"
    ProjectManager UserRole = "PROJECT_MANAGER"
    TeamMember     UserRole = "TEAM_MEMBER"
)
type User struct {
    ID        int            `gorm:"primaryKey;autoIncrement" json:"id"`
    Username  string         `gorm:"unique;not null" json:"username"`
    Email     string         `gorm:"unique;not null" json:"email"`
    Password  string         `gorm:"not null" json:"password"`
    Role      UserRole       `gorm:"type:user_role;not null" json:"role"`
    FirstName string         `json:"first_name"`
    LastName  string         `json:"last_name"`
    CreatedAt time.Time      `json:"created_at"`      
    UpdatedAt time.Time      `json:"updated_at"`     
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
func (u *User)BeforeCreate(tx *gorm.DB) error{
	u.ID = 0
	return nil
}