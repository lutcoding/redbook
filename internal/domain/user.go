package domain

type User struct {
	Id       int64
	Email    string
	Password string
	Phone    string
}

func (u User) Func() {
}
