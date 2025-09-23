package utils

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

// CommandResult 命令执行结果
type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Error    error
}

// CommandExecutor 命令执行器
type CommandExecutor struct {
	timeout time.Duration
}

// NewCommandExecutor 创建命令执行器
func NewCommandExecutor(timeout time.Duration) *CommandExecutor {
	return &CommandExecutor{
		timeout: timeout,
	}
}

// Execute 执行命令
func (e *CommandExecutor) Execute(ctx context.Context, command string, args ...string) (*CommandResult, error) {
	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// 创建命令
	cmd := exec.CommandContext(timeoutCtx, command, args...)

	// 执行命令并获取输出
	stdout, err := cmd.Output()
	result := &CommandResult{
		Stdout: string(stdout),
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.Stderr = string(exitError.Stderr)
			result.ExitCode = exitError.ExitCode()
		}
		result.Error = err
		return result, fmt.Errorf("command execution failed: %w", err)
	}

	return result, nil
}

// ExecuteWithStderr 执行命令并获取stdout和stderr
func (e *CommandExecutor) ExecuteWithStderr(ctx context.Context, command string, args ...string) (*CommandResult, error) {
	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// 创建命令
	cmd := exec.CommandContext(timeoutCtx, command, args...)

	// 执行命令并获取输出
	stdout, err := cmd.Output()
	stderr := ""

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			stderr = string(exitError.Stderr)
		}
	}

	result := &CommandResult{
		Stdout: string(stdout),
		Stderr: stderr,
		Error:  err,
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		}
		return result, fmt.Errorf("command execution failed: %w", err)
	}

	return result, nil
}