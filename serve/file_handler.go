package serve

import (
	routing "github.com/qiangxue/fasthttp-routing"
)

type FileHandler struct {
	App *App
}

func NewFileHandler(app *App) *FileHandler {
	return &FileHandler{App: app}
}

func (handler *FileHandler) handlePut(ctx *routing.Context) error {
	value := ctx.Request.Body()
	hash, err := handler.App.Hasher.Calc(value)
	if err != nil {
		return err
	}
	strHash := string(hash)
	err = handler.App.Storage.Store(strHash, value)
	if err != nil {
		return err
	}
	err = handler.App.Cache.Set(strHash, value)
	if err != nil {
		return err
	}
	ctx.SetBody(hash)
	return nil
}

func (handler *FileHandler) handleGet(ctx *routing.Context) error {
	hash := ctx.Param("hash")
	cached, err := handler.App.Cache.Get(hash)
	if err != nil {
		return err
	}
	if cached != nil {
		ctx.SetBody(cached)
		return nil
	}
	contents, err := handler.App.Storage.Get(hash)
	if err != nil {
		return err
	}
	err = handler.App.Cache.Set(hash, contents)
	if err != nil {
		return err
	}
	ctx.SetBody(contents)
	return nil
}