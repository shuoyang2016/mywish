package auth

import (
	"errors"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNameExist = errors.New("The user name already exists!")
)

type AuthModule struct {
	db *gorm.DB
}

func NewAuthModule() (*AuthModule, error) {
	db, err := startDBConnection()
	if err != nil {
		return nil, err
	}
	auth_module := &AuthModule{db: db}
	return auth_module, nil
}

func (s *AuthModule) CheckOrCreateUser(userName string, password string) error {
	db := s.db
	tx := db.Begin()
	var user User
	tx.Where(&User{Username: "test"}).First(&user)
	if user.Username == "test" {
		return ErrUserNameExist
	}
	user_to_add := User{Username: userName, FullName: "", PasswordHash: []byte(""), IsDisabled:false}
	hash_pass, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return err
	}
	user_to_add.PasswordHash = hash_pass
	tx.Create(&user_to_add)
	tx.Commit()
	return nil
}