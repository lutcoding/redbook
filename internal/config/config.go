package config

type Config struct {
	DB     DB     `yaml:"db"`
	Redis  Redis  `yaml:"redis"`
	Wechat Wechat `yaml:"wechat"`
	Ding   Ding   `yaml:"ding"`
}

type DB struct {
	Mysql Mysql `yaml:"mysql"`
	Mongo Mongo `yaml:"mongo"`
}

type Mysql struct {
	DSN string `yaml:"dsn"`
}

type Redis struct {
	Addr string `yaml:"addr"`
	Pwd  string `yaml:"pwd"`
}

type Wechat struct {
	AppID     string `yaml:"appID"`
	AppSecret string `yaml:"appSecret"`
}

type Ding struct {
	AppKey    string `yaml:"appKey"`
	AppSecret string `yaml:"appSecret"`
}

type Mongo struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	URI      string `yaml:"uri"`
}
