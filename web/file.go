package web

import (
	"github.com/hashicorp/golang-lru/v2"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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

type StaticResourceHandler struct {
	pathname string
	dir      string
	// 根据文件名后缀进行匹配文件类型
	extContentTypeMap map[string]string

	cache *lru.Cache[string, []byte]

	// 大文件不缓存
	maxSize int
}

type StaticResourceHandlerOption func(handler *StaticResourceHandler)

func NewStaticResourceHandler(pathname string, dir string, opts ...StaticResourceHandlerOption) (*StaticResourceHandler, error) {
	// TODO: 考虑缓存为对象, 能存储更多的信息
	c, err := lru.New[string, []byte](1000)
	if err != nil {
		return nil, err
	}
	r := &StaticResourceHandler{
		pathname: pathname,
		dir:      dir,
		cache:    c,
		extContentTypeMap: map[string]string{
			"jpg":  "image/jpg",
			"png":  "image/png",
			"jpeg": "image/jpeg",
			"pdf":  "image/pdf",
		},
		maxSize: 10 * 1024 * 1024, // 10MB
	}
	for _, opt := range opts {
		opt(r)
	}
	return r, nil
}

func StaticWithExtensionContentTypeMap(extContentTypeMap map[string]string) StaticResourceHandlerOption {
	return func(handler *StaticResourceHandler) {
		handler.extContentTypeMap = extContentTypeMap
	}
}

// func StaticWithCache(cache *lru.Cache) StaticResourceHandlerOption {
// 	return func (handler *StaticResourceHandler) {
// 		handler.cache = cache
// 	}
// }

func StaticWithMaxFileSize(maxSize int) StaticResourceHandlerOption {
	return func(handler *StaticResourceHandler) {
		handler.maxSize = maxSize
	}
}

func (r *StaticResourceHandler) Handle(ctx *Context) {
	// 1.拿到目标文件名
	// 2.定位到目标文件, 并且读取出来
	// 3.返回给前端

	file, err := ctx.PathValue(r.pathname)
	if err != nil {
		ctx.RespStatusCode = http.StatusBadRequest
		ctx.RespData = []byte("cannot found target file")
		return
	}

	dst := filepath.Join(r.dir, file)
	data, ok := r.cache.Get(file)
	if !ok {
		// 无缓存, 则重新读取文件, 再添加进入缓存中
		data, err := os.ReadFile(dst)
		if err != nil {
			ctx.RespStatusCode = http.StatusInternalServerError
			ctx.RespData = []byte(http.StatusText(http.StatusNotFound))
			return
		}
		if len(data) <= r.maxSize {
			r.cache.Add(file, data)
		}
	}

	ext := filepath.Ext(dst)[1:]
	contentType, ok := r.extContentTypeMap[ext]
	if !ok {
		ctx.RespStatusCode = http.StatusBadRequest
		ctx.RespData = []byte("cannot support " + ext + " media type")
		return
	}
	header := ctx.Resp.Header()
	header.Set("Content-Type", contentType)
	header.Set("Content-Length", strconv.Itoa(len(data)))

	ctx.RespStatusCode = http.StatusOK
	ctx.RespData = data
}
