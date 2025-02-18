package models

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	Role     string `json:"role"` 
}

type Role struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
