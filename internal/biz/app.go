package biz

type AppStore interface{}

type AppBiz struct {
	store AppStore
}

func NewAppBiz(store *AppStore) *AppBiz {
	return &AppBiz{}
}
