package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

// InitLogger initializes the zap logger
func InitLogger(debug bool) error {
	var config zap.Config
	
	if debug {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
	}
	
	// Set log level
	if debug {
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	} else {
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}
	
	var err error
	Logger, err = config.Build()
	if err != nil {
		return err
	}
	
	// Replace the global logger
	zap.ReplaceGlobals(Logger)
	return nil
}

// Sync flushes the logger
func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}