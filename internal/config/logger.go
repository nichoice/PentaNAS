package config

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
)

// LoggerConfig 日志配置结构体
type LoggerConfig struct {
	Logger struct {
		Level              string   `yaml:"level"`
		Encoding           string   `yaml:"encoding"`
		OutputPaths        []string `yaml:"output_paths"`
		ErrorOutputPaths   []string `yaml:"error_output_paths"`
		Development        bool     `yaml:"development"`
		DisableCaller      bool     `yaml:"disable_caller"`
		DisableStacktrace  bool     `yaml:"disable_stacktrace"`
		Sampling           struct {
			Initial    int `yaml:"initial"`
			Thereafter int `yaml:"thereafter"`
		} `yaml:"sampling"`
		EncoderConfig struct {
			TimeKey        string `yaml:"time_key"`
			LevelKey       string `yaml:"level_key"`
			NameKey        string `yaml:"name_key"`
			CallerKey      string `yaml:"caller_key"`
			MessageKey     string `yaml:"message_key"`
			StacktraceKey  string `yaml:"stacktrace_key"`
			LineEnding     string `yaml:"line_ending"`
			LevelEncoder   string `yaml:"level_encoder"`
			TimeEncoder    string `yaml:"time_encoder"`
			DurationEncoder string `yaml:"duration_encoder"`
			CallerEncoder  string `yaml:"caller_encoder"`
		} `yaml:"encoder_config"`
	} `yaml:"logger"`
}

// InitLogger 初始化日志器
func InitLogger(configPath string) (*zap.Logger, error) {
	// 读取配置文件
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取日志配置文件失败: %w", err)
	}

	// 解析配置
	var config LoggerConfig
	if err := yaml.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("解析日志配置失败: %w", err)
	}

	// 创建日志目录
	for _, path := range config.Logger.OutputPaths {
		if path != "stdout" && path != "stderr" {
			dir := filepath.Dir(path)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("创建日志目录失败: %w", err)
			}
		}
	}

	// 构建 zap 配置
	zapConfig := zap.Config{
		Level:             zap.NewAtomicLevelAt(parseLogLevel(config.Logger.Level)),
		Development:       config.Logger.Development,
		DisableCaller:     config.Logger.DisableCaller,
		DisableStacktrace: config.Logger.DisableStacktrace,
		Sampling: &zap.SamplingConfig{
			Initial:    config.Logger.Sampling.Initial,
			Thereafter: config.Logger.Sampling.Thereafter,
		},
		Encoding:         config.Logger.Encoding,
		EncoderConfig:    buildEncoderConfig(config.Logger.EncoderConfig),
		OutputPaths:      config.Logger.OutputPaths,
		ErrorOutputPaths: config.Logger.ErrorOutputPaths,
	}

	// 构建日志器
	logger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("构建日志器失败: %w", err)
	}

	return logger, nil
}

// parseLogLevel 解析日志级别
func parseLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// buildEncoderConfig 构建编码器配置
func buildEncoderConfig(config struct {
	TimeKey         string `yaml:"time_key"`
	LevelKey        string `yaml:"level_key"`
	NameKey         string `yaml:"name_key"`
	CallerKey       string `yaml:"caller_key"`
	MessageKey      string `yaml:"message_key"`
	StacktraceKey   string `yaml:"stacktrace_key"`
	LineEnding      string `yaml:"line_ending"`
	LevelEncoder    string `yaml:"level_encoder"`
	TimeEncoder     string `yaml:"time_encoder"`
	DurationEncoder string `yaml:"duration_encoder"`
	CallerEncoder   string `yaml:"caller_encoder"`
}) zapcore.EncoderConfig {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        config.TimeKey,
		LevelKey:       config.LevelKey,
		NameKey:        config.NameKey,
		CallerKey:      config.CallerKey,
		MessageKey:     config.MessageKey,
		StacktraceKey:  config.StacktraceKey,
		LineEnding:     zapcore.DefaultLineEnding,
	}

	// 设置级别编码器
	switch config.LevelEncoder {
	case "lowercase":
		encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	case "capital":
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	case "color":
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	default:
		encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	}

	// 设置时间编码器
	switch config.TimeEncoder {
	case "iso8601":
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	case "millis":
		encoderConfig.EncodeTime = zapcore.EpochMillisTimeEncoder
	case "nanos":
		encoderConfig.EncodeTime = zapcore.EpochNanosTimeEncoder
	case "rfc3339":
		encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	default:
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	// 设置持续时间编码器
	switch config.DurationEncoder {
	case "seconds":
		encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	case "nanos":
		encoderConfig.EncodeDuration = zapcore.NanosDurationEncoder
	case "ms":
		encoderConfig.EncodeDuration = zapcore.MillisDurationEncoder
	default:
		encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	}

	// 设置调用者编码器
	switch config.CallerEncoder {
	case "full":
		encoderConfig.EncodeCaller = zapcore.FullCallerEncoder
	case "short":
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	default:
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}

	return encoderConfig
}