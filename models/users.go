package models

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Name     string `json:"full_name"`
	Role     string `json:"user_role"`
	Email    string `json:"email"`
	Picture  string `json:"picture"`
}
