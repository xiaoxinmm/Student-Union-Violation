package model

import "time"

type User struct {
	ID           uint      `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	DisplayName  string    `json:"display_name"`
	Role         string    `json:"role"` // "admin" or "staff"
	CreatedAt    time.Time `json:"created_at"`
}

type Violation struct {
	ID          uint      `json:"id"`
	Dorm        string    `json:"dorm"`
	StudentName string    `json:"student_name"`
	ClassName   string    `json:"class_name"`
	Period      string    `json:"period"`
	Reason      string    `json:"reason"`
	Department  string    `json:"department"`
	Inspector   string    `json:"inspector"`
	PhotoPath   string    `json:"photo_path"`
	CreatedBy   uint      `json:"created_by"`
	CreatorName string    `json:"creator_name"` // joined field
	CreatedAt   time.Time `json:"created_at"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ViolationRequest struct {
	Dorm        string `form:"dorm" binding:"required,max=20"`
	StudentName string `form:"student_name" binding:"required,max=50"`
	ClassName   string `form:"class_name" binding:"required,max=50"`
	Period      string `form:"period" binding:"required,max=20"`
	Reason      string `form:"reason" binding:"required,max=2000"`
	Department  string `form:"department" binding:"required,max=30"`
	Inspector   string `form:"inspector" binding:"required,max=100"`
}

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}
