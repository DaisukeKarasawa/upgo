package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func Init(level string, output string, filePath string) error {
	// ログレベルの設定
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// エンコーダー設定
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.MessageKey = "message"
	encoderConfig.LevelKey = "level"
	encoderConfig.CallerKey = "caller"

	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// 出力先の設定
	var writeSyncer zapcore.WriteSyncer
	if output == "file" {
		// ログディレクトリの作成
		if err := os.MkdirAll(filePath[:len(filePath)-len("/upgo.log")], 0755); err != nil {
			return err
		}

		file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		writeSyncer = zapcore.AddSync(file)
	} else {
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	// コアの作成
	core := zapcore.NewCore(encoder, writeSyncer, zapLevel)

	// ロガーの作成
	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return nil
}

func Get() *zap.Logger {
	if Logger == nil {
		// デフォルトロガーを返す
		logger, _ := zap.NewProduction()
		return logger
	}
	return Logger
}
