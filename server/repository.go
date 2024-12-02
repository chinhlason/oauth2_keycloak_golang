package server

import (
	"gorm.io/gorm"
	"time"
)

type IRepository interface {
	Save(id, username, email, password, fullname, role string, createdAt, updatedAt time.Time) error
	Get(id string) (*User, error)
	Update(data UpdateUserDTO) error
	ChangePassword(id, password string) error
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) IRepository {
	return &Repository{db}
}

func (r *Repository) Save(id, username, email, password, fullname, role string, createdAt, updatedAt time.Time) error {
	user := User{
		ID:        id,
		Username:  username,
		Email:     email,
		FullName:  fullname,
		Password:  password,
		Role:      role,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
	return r.db.Create(&user).Error
}

func (r *Repository) Get(id string) (*User, error) {
	var user User
	err := r.db.Where("id = ?", id).First(&user).Error
	return &user, err
}

func (r *Repository) Update(data UpdateUserDTO) error {
	return r.db.Model(&User{}).Where("id = ?", data.ID).Updates(data).Error
}

func (r *Repository) ChangePassword(id, password string) error {
	return r.db.Model(&User{}).Where("id = ?", id).Update("password", password).Error
}
