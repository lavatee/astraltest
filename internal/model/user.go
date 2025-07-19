package model

type User struct {
	ID       int    `db:"user_id"`
	Login    string `db:"login"`
	Password string `db:"password_hash"`
}

type RegisterRequest struct {
	Token string `json:"token" binding:"required"`
	Login string `json:"login" binding:"required,min=8,alphanum"`
	Pswd  string `json:"pswd" binding:"required,min=8,containsany=!@#$%^&*,containsany=ABCDEFGHIJKLMNOPQRSTUVWXYZ,containsany=abcdefghijklmnopqrstuvwxyz,containsany=0123456789"`
}

type AuthRequest struct {
	Login string `json:"login" binding:"required"`
	Pswd  string `json:"pswd" binding:"required"`
}
