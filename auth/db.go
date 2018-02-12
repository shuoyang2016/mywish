package auth

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"time"
	"github.com/golang/glog"
)

type User struct {
	gorm.Model
	Username string
	FullName string
	PasswordHash []byte  // bcrypt password
	IsDisabled bool
}

type UserSession struct {
	gorm.Model
	SessionKey string
	UserID uint  // int not null, -- Could have a hard "references User"
	LoginTime time.Time
	LastSeenTime time.Time
}

func startDBConnection() (*gorm.DB, error) {
	db, err := gorm.Open("mysql", "mywishtest:mywishtest@/mywishtest?parseTime=true")
	if err != nil {
		glog.Info("failed to connect database")
		return nil, err
	}
	db.LogMode(true)
	// Migrate the schema
	db.AutoMigrate(&User{})
	db.AutoMigrate(&UserSession{})
	return db, nil
}