package auth

import (
	"errors"

	"github.com/golang/glog"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var _ glog.Level
var _ grpc.Address

var (
	ErrUserNameExist = errors.New("The user name already exists!")
)

type AuthModule struct {
	db *gorm.DB
}

func NewAuthModule(addr string) (*AuthModule, error) {
	db, err := startDBConnection(addr)
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
	if err := tx.Where("username = ?", userName).First(&user).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			tx.Rollback()
			return err
		}
	}
	if user.Username == userName {
		return ErrUserNameExist
	}
	user_to_add := User{Username: userName, FullName: "", PasswordHash: []byte(password), IsDisabled: false}
	hash_pass, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return err
	}
	user_to_add.PasswordHash = hash_pass
	if err := tx.Create(&user_to_add).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func AuthFunc(ctx context.Context) (context.Context, error) {
	return ctx, nil
}
