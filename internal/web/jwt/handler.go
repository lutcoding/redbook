package jwt

type Handler struct {
	AccessKey  []byte
	RefreshKey []byte
}

func NewHandler() *Handler {
	return &Handler{
		AccessKey:  []byte("NqdHZfporsLtXRTPhc01IZJXDnFsaTHsmsMWixjPEgQJyiZxsXKcsmkg1XvAWXIp"),
		RefreshKey: []byte("NqdHZfporsLtXRTPhc01IZJXDnFsaTHsmsMWixjPEgQJyiZxsXKcsmkg1XvAWXIx"),
	}
}
