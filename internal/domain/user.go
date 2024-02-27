package domain

type User struct {
	Id         int64
	Email      string
	Password   string
	Phone      string
	WechatInfo Wechat
}

func (u User) Func() {
}
