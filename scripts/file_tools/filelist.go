package scripts

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// List 返回符合选项的文件列表（带分页、排序、过滤）
func List(opts Options) ([]File, int, error) {
	if err := opts.Validate(); err != nil {
		return nil, 0, err
	}

	var (
		filesChan = make(chan File, 1000) // 缓冲通道，避免阻塞
		wg        sync.WaitGroup
		// ctx, cancel = context.WithCancel(context.Background())
		ctx, cancel = context.WithCancel(context.Background())
	)
	defer cancel()

	// 启动协程：根据是否递归选择遍历方式
	if opts.Recursive {
		wg.Add(1)
		go walkDir(ctx, opts.Path, opts.ShowHidden, opts.SearchTerm, opts.FilterType, filesChan, &wg)
	} else {
		wg.Add(1)
		go listDir(ctx, opts.Path, opts.ShowHidden, opts.SearchTerm, opts.FilterType, filesChan, &wg)
	}

	// 关闭通道：当所有 goroutine 完成后关闭
	go func() {
		wg.Wait()
		close(filesChan)
	}()

	// 收集所有文件
	var allFiles []File
	for f := range filesChan {
		allFiles = append(allFiles, f)
	}

	// 排序
	sortFiles(allFiles, opts.SortBy, opts.Order)

	// 分页
	total := len(allFiles)
	start := (opts.Page - 1) * opts.PageSize
	end := start + opts.PageSize
	if start >= total {
		return []File{}, total, nil // 无数据
	}
	if end > total {
		end = total
	}

	paged := allFiles[start:end]

	return paged, total, nil
}

// listDir 非递归遍历单个目录
func listDir(ctx context.Context, path string, showHidden bool, searchTerm string, filterType string, out chan<- File, wg *sync.WaitGroup) {
	defer wg.Done()

	entries, err := os.ReadDir(path)
	if err != nil {
		select {
		case <-ctx.Done():
			return
		default:
			fmt.Fprintf(os.Stderr, "无法读取目录 %s: %v\n", path, err)
		}
		return
	}

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return
		default:
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		// 过滤隐藏文件
		if !showHidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		// 过滤类型
		mode := getFileMode(info)
		if filterType != "" && mode != filterType {
			continue
		}

		// 搜索关键词
		if searchTerm != "" && !strings.Contains(entry.Name(), searchTerm) {
			continue
		}

		file := File{
			Name:     entry.Name(),
			Size:     info.Size(),
			Mode:     mode,
			Modified: info.ModTime(),
			Perm:     getPermissions(info.Mode()),
		}

		if info.IsDir() {
			file.Size = 0
		}

		out <- file
	}
}

// walkDir 递归遍历目录树（并发安全）
func walkDir(ctx context.Context, root string, showHidden bool, searchTerm string, filterType string, out chan<- File, wg *sync.WaitGroup) {
	defer wg.Done()

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			select {
			case <-ctx.Done():
				return filepath.SkipDir
			default:
				fmt.Fprintf(os.Stderr, "跳过: %s (%v)\n", path, err)
				return nil
			}
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		// 计算相对路径
		relPath, _ := filepath.Rel(root, path)
		if relPath == "." {
			return nil
		}

		// 过滤隐藏文件
		if !showHidden && strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		// 过滤类型
		mode := getFileMode(info)
		if filterType != "" && mode != filterType {
			return nil
		}

		// 搜索关键词
		if searchTerm != "" && !strings.Contains(d.Name(), searchTerm) {
			return nil
		}

		file := File{
			Name:     d.Name(),
			Size:     info.Size(),
			Mode:     mode,
			Modified: info.ModTime(),
			Perm:     getPermissions(info.Mode()),
			FullPath: relPath,
		}

		if info.IsDir() {
			file.Size = 0
		}

		select {
		case <-ctx.Done():
			return filepath.SkipDir
		case out <- file:
			// 发送成功
		}

		return nil
	})

	if err != nil {
		select {
		case <-ctx.Done():
		default:
			fmt.Fprintf(os.Stderr, "WalkDir 错误: %v\n", err)
		}
	}
}

// sortFiles 根据选项对文件列表排序
func sortFiles(files []File, sortBy, order string) {
	sort.Slice(files, func(i, j int) bool {
		var less bool

		switch sortBy {
		case "name":
			less = files[i].Name < files[j].Name
		case "size":
			less = files[i].Size < files[j].Size
		case "modified":
			less = files[i].Modified.Before(files[j].Modified)
		default:
			less = files[i].Name < files[j].Name
		}

		if order == "desc" {
			return !less
		}
		return less
	})
}

// getFileMode 返回 "file" 或 "dir"
func getFileMode(info os.FileInfo) string {
	if info.IsDir() {
		return "dir"
	}
	return "file"
}

// getPermissions 返回 Unix 权限字符串
func getPermissions(mode os.FileMode) string {
	var perm string

	// User
	if mode&0400 != 0 {
		perm += "r"
	} else {
		perm += "-"
	}
	if mode&0200 != 0 {
		perm += "w"
	} else {
		perm += "-"
	}
	if mode&0100 != 0 {
		perm += "x"
	} else {
		perm += "-"
	}

	// Group
	if mode&0040 != 0 {
		perm += "r"
	} else {
		perm += "-"
	}
	if mode&0020 != 0 {
		perm += "w"
	} else {
		perm += "-"
	}
	if mode&0010 != 0 {
		perm += "x"
	} else {
		perm += "-"
	}

	// Others
	if mode&0004 != 0 {
		perm += "r"
	} else {
		perm += "-"
	}
	if mode&0002 != 0 {
		perm += "w"
	} else {
		perm += "-"
	}
	if mode&0001 != 0 {
		perm += "x"
	} else {
		perm += "-"
	}

	return perm
}
