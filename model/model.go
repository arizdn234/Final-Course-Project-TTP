package model

import "time"

type Category struct {
	ID   int    `gorm:"primaryKey" json:"id"`
	Name string `json:"name"`
}

type User struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	Fullname  string    `json:"fullname" gorm:"type:varchar(255);"`
	Email     string    `json:"email" gorm:"type:varchar(255);not null"`
	Password  string    `json:"password" gorm:"type:varchar(255);not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserLogin struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserRegister struct {
	Fullname string `json:"fullname" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type Task struct {
	ID         int    `gorm:"primaryKey" json:"id"`
	Title      string `json:"title"`
	Deadline   string `json:"deadline"`
	Priority   int    `json:"priority"`
	Status     string `json:"status"`
	CategoryID int    `json:"category_id"`
	UserID     int    `json:"user_id"`
}

type Session struct {
	ID     int       `gorm:"primaryKey" json:"id"`
	Token  string    `json:"token"`
	Email  string    `json:"email"`
	Expiry time.Time `json:"expiry"`
}

type TaskCategory struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Category string `json:"category"`
}

type UserTaskCategory struct {
	ID       int    `json:"id"`
	Fullname string `json:"fullname"`
	Email    string `json:"email"`
	Task     string `json:"task"`
	Deadline string `json:"deadline"`
	Priority int    `json:"priority"`
	Status   string `json:"status"`
	Category string `json:"category"`
}

type Credential struct {
	Host         string
	Username     string
	Password     string
	DatabaseName string
	Port         int
	Schema       string
}
