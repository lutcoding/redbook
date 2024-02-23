package cache

// CodeLocalCache 单实例生产环境下
// 考虑并发问题
type CodeLocalCache struct {
}

func NewCodeLocalCache() *CodeLocalCache {
	return &CodeLocalCache{}
}
