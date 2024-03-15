package article

type Article struct {
	Id      int64  `gorm:"primaryKey, autoIncrement" bson:"id,omitempty"`
	Tittle  string `gorm:"type=varchar(1024)" bson:"tittle,omitempty"`
	Content string `gorm:"type=BLOB" bson:"content,omitempty"`

	Status   uint8 `bson:"status,omitempty"`
	AuthorId int64 `gorm:"index" bson:"author_id,omitempty"`

	CreateTime int64 `bson:"create_time,omitempty"`
	UpdateTime int64 `bson:"update_time,omitempty"`
}

type PublishArticle struct {
	Id      int64  `gorm:"primaryKey, autoIncrement" bson:"id,omitempty"`
	Tittle  string `gorm:"type=varchar(1024)" bson:"tittle,omitempty"`
	Content string `gorm:"type=BLOB" bson:"content,omitempty"`

	Status   uint8 `bson:"status,omitempty"`
	AuthorId int64 `gorm:"index" bson:"author_id,omitempty"`

	CreateTime int64 `bson:"create_time,omitempty"`
	UpdateTime int64 `bson:"update_time,omitempty"`
}
