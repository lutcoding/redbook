package domain

type ArticleStatus uint8

const (
	// ArticleStatusUnKnown 一般定义常量, 最好不好把零值定义成有意义的值
	// 比如前端传过来, 不清楚是否具体定义了零值还是没传，默认值
	ArticleStatusUnKnown = iota
	ArticleStatusUnPublished
	ArticleStatusPublished
	ArticleStatusPrivate
)

func (s ArticleStatus) ToUint8() uint8 {
	return uint8(s)
}

type Article struct {
	Id       int64
	Tittle   string
	Content  string
	AuthorId int64
	ArticleStatus
}
