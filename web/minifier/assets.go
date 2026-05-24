package minifier

import (
	"bytes"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"

	"github.com/evanw/esbuild/pkg/api"
)

type MinifiedAssetsFS struct {
	base  fs.FS
	cache sync.Map // map[string][]byte
}

func NewMinifiedAssetsFS(base fs.FS) *MinifiedAssetsFS {
	return &MinifiedAssetsFS{base: base}
}

type minifiedAssetsFile struct {
	data   []byte
	reader *bytes.Reader
	info   fs.FileInfo
}

func (f *MinifiedAssetsFS) Open(name string) (fs.File, error) {
	file, err := f.base.Open("assets/" + name)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return file, nil
	}

	ext := strings.ToLower(filepath.Ext(name))
	if ext != ".js" && ext != ".css" && ext != ".mjs" {
		return file, nil
	}

	if cached, ok := f.cache.Load(name); ok {
		data := cached.([]byte)
		return newMinifiedAssetsFile(data, info), nil
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	loader := api.LoaderJS
	if ext == ".css" {
		loader = api.LoaderCSS
	}

	result := api.Transform(string(content), api.TransformOptions{
		Loader:            loader,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		MinifyWhitespace:  true,
		MangleQuoted:      api.MangleQuotedTrue,
		Drop:              api.DropConsole | api.DropDebugger,
		LegalComments:     api.LegalCommentsNone,
		Sourcefile:        name,
	})

	if len(result.Errors) > 0 {
		// If minification fails, serve the original asset content.
		return newMinifiedAssetsFile(content, info), nil
	}

	data := result.Code
	f.cache.Store(name, data)
	return newMinifiedAssetsFile(data, info), nil
}

func newMinifiedAssetsFile(data []byte, info fs.FileInfo) *minifiedAssetsFile {
	return &minifiedAssetsFile{
		data:   data,
		reader: bytes.NewReader(data),
		info:   &minifiedAssetsFileInfo{FileInfo: info, size: int64(len(data))},
	}
}

func (f *minifiedAssetsFile) Read(p []byte) (int, error) {
	return f.reader.Read(p)
}

func (f *minifiedAssetsFile) Close() error {
	return nil
}

func (f *minifiedAssetsFile) Seek(offset int64, whence int) (int64, error) {
	return f.reader.Seek(offset, whence)
}

func (f *minifiedAssetsFile) Stat() (fs.FileInfo, error) {
	return f.info, nil
}

func (f *minifiedAssetsFile) ReadDir(count int) ([]fs.DirEntry, error) {
	return nil, fs.ErrInvalid
}

type minifiedAssetsFileInfo struct {
	fs.FileInfo
	size int64
}

func (f *minifiedAssetsFileInfo) Size() int64 {
	return f.size
}
