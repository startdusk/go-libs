package web

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// 目前来说, 上传下载的功能推荐大家使用OSS, 而不是自建服务器

type FileUploader struct {
	FileField string
	// 为什么交给用户处理这个文件路径
	// 因为存在文件重名的问题, 这要交给用户来处理
	DstPathFunc func(fileHeader *multipart.FileHeader) string
}

func (u FileUploader) Handle() HandleFunc {
	return func(ctx *Context) {
		// 上传文件逻辑

		// 1.读到文件内容
		// 2.计算出目标路径
		// 3.保存文件
		// 4.返回响应
		file, fileHeader, err := ctx.Req.FormFile(u.FileField)
		if err != nil {
			ctx.RespStatusCode = http.StatusInternalServerError
			ctx.RespData = []byte(http.StatusText(http.StatusNotFound))
			return
		}
		defer file.Close()

		dst := u.DstPathFunc(fileHeader)
		// 如果存在额外的路径, 先将路径创建出来
		if dir, _ := filepath.Split(dst); dir != "" {
			err := os.MkdirAll(dir, 0o666)
			if err != nil {
				ctx.RespStatusCode = http.StatusInternalServerError
				ctx.RespData = []byte(http.StatusText(http.StatusNotFound))
				return
			}
		}

		// O_WRONLY 写入数据
		// O_TRUNC 文件本身存在就清空文件内容
		// O_CREATE 创建一个新的文件
		dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0o666)
		if err != nil {
			ctx.RespStatusCode = http.StatusInternalServerError
			ctx.RespData = []byte(http.StatusText(http.StatusNotFound))
			return
		}
		defer dstFile.Close()

		// 最后一个参数为 buf 用来控制每次copy的大小, 如果传nil, 它将会为我们生成一个32k大小的buf来每次copy32k的数据
		// buf会影响性能, 你要考虑复用
		if _, err := io.CopyBuffer(dstFile, file, nil); err != nil {
			ctx.RespStatusCode = http.StatusInternalServerError
			ctx.RespData = []byte(http.StatusText(http.StatusNotFound))
			return
		}

		ctx.RespStatusCode = http.StatusOK
		ctx.RespData = []byte(http.StatusText(http.StatusOK))
	}
}

type FileDownloader struct {
	FileField string
	Dir       string
}

func (f *FileDownloader) Handle() HandleFunc {
	return func(ctx *Context) {
		req, err := ctx.QueryValue(f.FileField)
		if err != nil {
			ctx.RespStatusCode = http.StatusBadRequest
			ctx.RespData = []byte("cannot found target file")
			return
		}
		path := filepath.Join(f.Dir, filepath.Clean(req))
		filename := filepath.Base(path)
		// 安全校验, 防止相对路径引起攻击者下载了你的系统文件
		absPath, err := filepath.Abs(path)
		if err != nil {
			ctx.RespStatusCode = http.StatusBadRequest
			ctx.RespData = []byte("cannot found target file")
			return
		}
		if !strings.Contains(absPath, f.Dir) {
			ctx.RespStatusCode = http.StatusBadRequest
			ctx.RespData = []byte("cannot found target file")
			return
		}

		header := ctx.Resp.Header()
		// 指定为 attachment 就是保存在本地;filename就是设置文件的名字
		header.Set("Content-Disposition", "attachment;filename="+filename)
		header.Set("Content-Description", "File Transfer")
		// octet-stream 表示通用二进制文件
		header.Set("Content-Type", "application/octet-stream")
		// 这里设置为binary, 相当于直接传输
		header.Set("Content-Transfer-Encoding", "binary")
		header.Set("Expires", "0")
		// must-revalidate 消除缓存, 每次都从服务器获取
		header.Set("Cache-Control", "must-revalidate")
		header.Set("Pragma", "public")
		http.ServeFile(ctx.Resp, ctx.Req, path)
	}
}
