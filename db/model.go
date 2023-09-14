package db

import (
	"gorm.io/gorm"
	"strings"
	"time"
)

type Channel struct {
	gorm.Model
	Name string `json:"name" gorm:"type:varchar(48) not null;unique"`

	Subscribers []*User `gorm:"many2many:subscriptions;constraint:OnDelete:CASCADE;"`

	CreatorID int64
	Creator   *User `gorm:"belongsTo:User"`
}

type User struct {
	// All these data are received from Telegram
	ID        int64  `json:"id" gorm:"primarykey"`               // This must be a signed int, because telegram assign negative id to groups
	FirstName string `json:"first_name" gorm:"type:varchar(64)"` // https://limits.tginfo.me/en
	LastName  string `json:"last_name" gorm:"type:varchar(64)"`
	Username  string `json:"username" gorm:"type:varchar(32)"` //https://core.telegram.org/method/account.checkUsername
	PhotoUrl  string `json:"photo_url" gorm:"type:varchar(128)"`

	Active *bool `json:"active" gorm:"default:false"`
	Admin  *bool `json:"admin" gorm:"default:false"`

	Subscriptions []*Channel `gorm:"many2many:subscriptions;constraint:OnDelete:CASCADE;"`
}

func (u *User) IsActive() bool {
	return u.Active != nil && *u.Active
}

func (u *User) IsAdmin() bool {
	return u.Admin != nil && *u.Admin
}

// Greet returns a proper name compiled from the FirstName, LastName and Username fields
func (u *User) Greet() string {
	name := u.FirstName // first name must be always present
	if u.LastName != "" {
		name += " " + u.LastName
	}

	if strings.TrimSpace(name) == "" {
		return "@" + u.Username
	} else {
		return name
	}

}

type Token struct {
	gorm.Model

	Name string `gorm:"type:varchar(48) not null; unique"`

	LastUsed *time.Time `gorm:"null"`

	TokenHash []byte `json:"-" gorm:"not null; unique"`

	AllowedChannels []*Channel `gorm:"many2many:token_channels;constraint:OnDelete:CASCADE;"`

	// quick and dirty
	CapNotify   bool `gorm:"not null;default:false"`
	CapQuestion bool `gorm:"not null;default:false"`
}
