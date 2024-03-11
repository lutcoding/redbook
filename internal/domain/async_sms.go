package domain

type AsyncSms struct {
	Id       int64
	RetryMax int64
	AsyncSmsConfig
}

type AsyncSmsConfig struct {
	TplId   string
	Args    []string
	Numbers []string
}
