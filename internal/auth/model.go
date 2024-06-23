package auth

type User struct {
	UserID     int
	Username   string
	PublicKey  string
	Authorized bool
}

type Users []User
