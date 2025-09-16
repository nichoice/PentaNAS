package scripts

import (
	"fmt"
	"time"
)

// File 表示一个文件或目录
type File struct {
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	Mode     string    `json:"mode"` // "file" 或 "dir"
	Modified time.Time `json:"modified"`
	Perm     string    `json:"perm"`
	FullPath string    `json:"full_path,omitempty"` // 用于递归时保留完整路径
}

// Options 控制文件列表行为
type Options struct {
	Path       string // 要扫描的根目录
	ShowHidden bool   // 是否显示隐藏文件（.开头）
	FilterType string // 过滤类型："file"、"dir"、""（不过滤）
	SearchTerm string // 按文件名模糊搜索（如 ".jpg"、"report"）
	SortBy     string // "name", "size", "modified"
	Order      string // "asc", "desc"
	Page       int    // 当前页码，从 1 开始
	PageSize   int    // 每页大小
	Recursive  bool   // 是否递归遍历子目录
}

// Validate 验证选项是否合法
func (o *Options) Validate() error {
	if o.Path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	if o.Page < 1 {
		o.Page = 1
	}
	if o.PageSize < 1 {
		o.PageSize = 100
	}
	if o.PageSize > 10000 {
		o.PageSize = 10000 // 防止内存爆炸
	}
	if o.SortBy != "name" && o.SortBy != "size" && o.SortBy != "modified" {
		o.SortBy = "name"
	}
	if o.Order != "asc" && o.Order != "desc" {
		o.Order = "asc"
	}
	if o.FilterType != "file" && o.FilterType != "dir" && o.FilterType != "" {
		o.FilterType = ""
	}
	return nil
}
