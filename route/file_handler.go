package route

import (
	"fmt"
	"strings"
	"time"

	"github.com/heynemann/hyper-cas/utils"
	"github.com/valyala/fasthttp"
)

type FileHandler struct {
	App *App
}

func NewFileHandler(app *App) *FileHandler {
	return &FileHandler{App: app}
}

func (handler *FileHandler) handleGet(ctx *fasthttp.RequestCtx) {
	distroHash, path, err := handler.App.GetDistroAndPath(string(ctx.Host()), string(ctx.Path()), ctx.Request.Header.Peek)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	label, err := handler.App.Cache.Get(distroHash)
	if label != nil {
		distroHash = string(label)
	} else {
		label, _ := handler.App.Storage.GetLabel(distroHash)
		if label != "" {
			handler.App.Cache.Set(distroHash, []byte(label))
			distroHash = label
		}
	}

	distro, err := handler.App.Cache.GetDistro(distroHash)
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString("Error getting distro from cache.")
		return
	}
	if distro == nil {
		distroFiles, err := handler.App.Storage.GetDistro(distroHash)
		if err != nil {
			ctx.SetStatusCode(500)
			ctx.SetBodyString("Error getting distro.")
			return
		}

		distro, err = handler.buildDistro(distroFiles)
		if err != nil {
			ctx.SetStatusCode(500)
			ctx.SetBodyString("Error building distro.")
			return
		}

		err = handler.App.Cache.SetDistro(distroHash, distro)
		if err != nil {
			ctx.SetStatusCode(500)
			ctx.SetBodyString("Error storing distro in cache.")
			return
		}
	}

	// Serve index.html if ends in '/' or if root of distribution
	if path == "" || strings.HasSuffix(path, "/") {
		path = fmt.Sprintf("%s/index.html", path)
	}
	path = strings.TrimLeft(path, "/")
	if val, ok := distro[path]; ok {
		contents, err := handler.getFile(val)
		if err != nil {
			ctx.SetStatusCode(500)
			ctx.SetBodyString(fmt.Sprintf("%v", err))
			return
		}
		err = writeContents(ctx, contents)
		if err != nil {
			ctx.SetStatusCode(500)
			ctx.SetBodyString(fmt.Sprintf("%v", err))
		}
	} else {
		ctx.SetStatusCode(404)
	}
}

func (h *FileHandler) buildDistro(paths []string) (map[string]string, error) {
	r := map[string]string{}
	for _, p := range paths {
		parts := strings.Split(p, ":")
		path := parts[0]
		hash := parts[1]
		r[path] = hash
	}
	return r, nil
}

func (h *FileHandler) getFile(hash string) ([]byte, error) {
	f, err := h.App.Cache.Get(hash)
	if err != nil {
		return nil, err
	}
	if f != nil {
		return f, nil
	}
	f, err = h.App.Storage.Get(hash)
	if err != nil {
		return nil, err
	}
	err = h.App.Cache.Set(hash, f)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func writeContents(ctx *fasthttp.RequestCtx, contents []byte) error {
	gzipEnabled := strings.Contains(string(ctx.Request.Header.Peek("Accept-Encoding")), "gzip")
	ctx.Response.Header.Add("content-type", "text/html; charset=utf-8")
	ctx.Response.Header.Add("date", time.Now().Format("RFC1123"))
	ctx.Response.Header.Set("server", "hyper-cas")
	if gzipEnabled {
		ctx.Response.Header.Add("content-encoding", "gzip")
		ctx.SetBody(contents)
	} else {
		res, err := utils.Unzip(contents)
		if err != nil {
			return err
		}
		ctx.SetBody(res)
	}

	return nil
}