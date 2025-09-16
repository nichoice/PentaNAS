package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	disk "pnas/internal/utils"
	filelist "pnas/scripts/file_tools"
	"strings"
	"time"
)

func main() {
	// 🚀 模拟测试：你可以换成真实大目录，如 /home/user/downloads
	path := "/Users/nic/Documents/workspace/pnas/internal/utils" // 创建一个包含大量文件的测试目录！

	// 创建测试数据（可选：生成10万个文件用于测试）
	//	createTestFiles(path, 100_000)

	opts := filelist.Options{
		Path:       path,
		ShowHidden: false,      // 显示 .xxx 文件
		FilterType: "",         // 不过滤，全部返回
		SearchTerm: "",         // 只找 .txt 文件
		SortBy:     "modified", // 按修改时间排序
		Order:      "desc",     // 最新在前
		Page:       1,
		PageSize:   50,
		Recursive:  true, // 递归扫描子目录
	}

	fmt.Printf("🔍 正在扫描目录: %s (含 %d 文件)\n", path, 100_000)
	fmt.Printf("⚙️ 选项: %+v\n\n", opts)

	start := time.Now()

	files, total, err := filelist.List(opts)
	if err != nil {
		log.Fatal(err)
	}

	duration := time.Since(start)

	fmt.Printf("✅ 成功！耗时: %v\n", duration)
	fmt.Printf("📊 总计: %d 个文件 | 本页: %d 个\n", total, len(files))
	fmt.Println("\n📄 前 10 个结果:")
	fmt.Printf("%-30s %-8s %-12s %-20s %s\n",
		"NAME", "TYPE", "SIZE", "MODIFIED", "PERM")
	fmt.Println(strings.Repeat("-", 90))

	formatFile(files)

	// 测试第2页
	fmt.Println("\n--- 第2页 ---")
	opts.Page = 2
	files2, _, _ := filelist.List(opts)
	fmt.Printf("第2页: %d 个文件\n", len(files2))
	formatFile(files2)

	disk.ScanDisks()

}

// createTestFiles 创建大量测试文件（仅用于演示）
func createTestFiles(dir string, count int) {
	os.MkdirAll(dir, 0755)
	for i := 0; i < count; i++ {
		filename := fmt.Sprintf("file_%06d.txt", i)
		if i%10 == 0 {
			filename = fmt.Sprintf(".hidden_%06d.log", i)
		}
		fullPath := filepath.Join(dir, filename)
		err := os.WriteFile(fullPath, []byte(fmt.Sprintf("content-%d", i)), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "写入失败: %s\n", fullPath)
		}
	}
}

func formatFile(files []filelist.File) {
	for i, f := range files {
		if i >= 10 {
			break
		}
		modTime := f.Modified.Format("2006-01-02 15:04")
		sizeStr := fmt.Sprintf("%d", f.Size)
		if f.Mode == "dir" {
			sizeStr = "<DIR>"
		}
		fmt.Printf("%-30s %-8s %-12s %-20s %s\n",
			f.Name, f.Mode, sizeStr, modTime, f.Perm)
	}
}
