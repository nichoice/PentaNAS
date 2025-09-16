package utils

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

// 封装命令执行结果
type CmdResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
	Error    error
}

// 封装命令执行选项
type ExecOptions struct {
	Timeout time.Duration
	Dir     string
	Env     []string
	Input   string
	Stream  bool
}

func ExecCommand(opts ExecOptions, command string, args ...string) *CmdResult {
	result := &CmdResult{}

	start := time.Now()
	defer func() {
		result.Duration = time.Since(start)
	}()

	// 创建带有超时的上下文
	var ctx context.Context
	var cancel context.CancelFunc

	if opts.Timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), opts.Timeout)
	} else {
		ctx = context.Background()
	}
	defer cancel()

	// 创建命令
	cmd := exec.CommandContext(ctx, command, args...)

	if opts.Dir != "" {
		cmd.Dir = opts.Dir
	}
	if opts.Env != nil {
		cmd.Env = append(os.Environ(), opts.Env...)
	}

	// 设置输入输出
	var stdoutBuf, stderrBuf strings.Builder
	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	if opts.Input != "" {
		stdinPipe, _ := cmd.StdinPipe()
		go func() {
			defer stdinPipe.Close()
			io.WriteString(stdinPipe, opts.Input)
		}()
	}

	if err := cmd.Start(); err != nil {
		result.Error = fmt.Errorf("failed to start command: %w", err)
		return result
	}

	if opts.Stream {
		// 流式输出到控制台同时保存到缓冲区
		multiStdout := io.MultiWriter(&stdoutBuf, os.Stdout)
		multiStderr := io.MultiWriter(&stderrBuf, os.Stderr)

		go func() {
			scanner := bufio.NewScanner(stdoutPipe)
			for scanner.Scan() {
				multiStdout.Write([]byte(scanner.Text() + "\n"))
			}
		}()

		go func() {
			scanner := bufio.NewScanner(stderrPipe)
			for scanner.Scan() {
				multiStderr.Write([]byte(scanner.Text() + "\n"))
			}
		}()
	} else {
		// 仅保存到缓冲区
		go func() {
			io.Copy(&stdoutBuf, stdoutPipe)
		}()

		go func() {
			io.Copy(&stderrBuf, stderrPipe)
		}()
	}

	err := cmd.Wait()
	result.Stdout = stdoutBuf.String()
	result.Stderr = stderrBuf.String()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
			result.Error = fmt.Errorf("command exited with code %d", result.ExitCode)
		} else if ctx.Err() == context.DeadlineExceeded {
			result.Error = fmt.Errorf("command timed out: %w", err)
		} else if ctx.Err() == context.Canceled {
			result.Error = fmt.Errorf("command canceled: %w", err)
		} else {
			result.Error = fmt.Errorf("command failed: %w", err)
		}
	}
	return result
}

/**
 * 执行简单命令
 * result := ExecCommand(ExecOptions{}, "ls", "-la")
	if result.Error != nil {
		fmt.Printf("错误: %v\n", result.Error)
	} else {
		fmt.Printf("输出: \n%s\n", result.Stdout)
		fmt.Printf("执行时间: %v\n", result.Duration)
	}
 *
 *
 *  带超时的命令
	result2 := ExecCommand(ExecOptions{Timeout: 3 * time.Second}, "sleep", "5")
	if result2.Error != nil {
		fmt.Printf("预期中的错误: %v\n", result2.Error)
	} else {
		fmt.Printf("输出: %s\n", result2.Stdout)
	}
	fmt.Printf("执行时间: %v\n", result2.Duration)
 *
 *
 * 流式输出
 * result3 := ExecCommand(ExecOptions{Stream: true}, "bash", "-c", "for i in {1..3}; do echo \"输出 $i\"; sleep 0.5; done")
	if result3.Error != nil {
		fmt.Printf("错误: %v\n", result3.Error)
	}
	fmt.Printf("执行时间: %v\n", result3.Duration)

	带环境变量和工作目录
	result4 := ExecCommand(ExecOptions{
			Dir: "/tmp",
			Env: []string{"CUSTOM_VAR=hello"},
		}, "bash", "-c", "echo $CUSTOM_VAR && pwd")
		if result4.Error != nil {
			fmt.Printf("错误: %v\n", result4.Error)
		} else {
			fmt.Printf("输出: %s\n", result4.Stdout)
		}

	带输入的命令
	result5 := ExecCommand(ExecOptions{
			Input: "Alice\nBob\nCharlie",
		}, "grep", "li")
		if result5.Error != nil {
			fmt.Printf("错误: %v\n", result5.Error)
		} else {
			fmt.Printf("输出: %s\n", result5.Stdout)
		}
*/
