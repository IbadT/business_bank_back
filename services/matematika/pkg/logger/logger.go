package logger

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	// ANSI color codes
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
	colorGray   = "\033[90m"

	// Bright colors
	colorBrightRed    = "\033[91m"
	colorBrightCyan   = "\033[96m"
)

var (
	globalLogger *Logger
)

type Logger struct {
	entry *logrus.Entry
}

type Fields map[string]interface{}

func init() {
	logrus.SetFormatter(&CustomFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)
	globalLogger = &Logger{entry: logrus.NewEntry(logrus.StandardLogger())}
}

// GetLogger возвращает глобальный экземпляр логгера
func GetLogger() *Logger {
	return globalLogger
}

// WithOperation создает новый логгер с указанной операцией
func (l *Logger) WithOperation(op string) *Logger {
	return &Logger{
		entry: l.entry.WithField("operation", op),
	}
}

// WithFields добавляет дополнительные поля к логгеру
func (l *Logger) WithFields(fields Fields) *Logger {
	logrusFields := logrus.Fields{}
	for k, v := range fields {
		logrusFields[k] = v
	}
	return &Logger{
		entry: l.entry.WithFields(logrusFields),
	}
}

// Info логирует информационное сообщение
func (l *Logger) Info(message string, args ...interface{}) {
	msg := formatMessage(message, args...)
	l.entry.Info(msg)
}

// Success логирует успешную операцию
func (l *Logger) Success(message string, args ...interface{}) {
	msg := formatMessage(message, args...)
	l.entry.WithField("type", "success").Info(msg)
}

// Warn логирует предупреждение
func (l *Logger) Warn(message string, args ...interface{}) {
	msg := formatMessage(message, args...)
	l.entry.Warn(msg)
}

// Error логирует ошибку
func (l *Logger) Error(err error, message string, args ...interface{}) {
	msg := formatMessage(message, args...)
	if err != nil {
		l.entry.WithError(err).Error(msg)
	} else {
		l.entry.Error(msg)
	}
}

// Debug логирует отладочное сообщение
func (l *Logger) Debug(message string, args ...interface{}) {
	msg := formatMessage(message, args...)
	l.entry.Debug(msg)
}

// Fatal логирует критическую ошибку и завершает программу
func (l *Logger) Fatal(err error, message string, args ...interface{}) {
	msg := formatMessage(message, args...)
	if err != nil {
		l.entry.WithError(err).Fatal(msg)
	} else {
		l.entry.Fatal(msg)
	}
}

// formatMessage форматирует сообщение с аргументами
func formatMessage(message string, args ...interface{}) string {
	if len(args) == 0 {
		return message
	}
	return fmt.Sprintf(message, args...)
}

// CustomFormatter кастомный форматтер для logrus
type CustomFormatter struct{}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b strings.Builder
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	
	// Получаем операцию из полей
	operation, hasOp := entry.Data["operation"].(string)
	if !hasOp {
		operation = "unknown"
	}
	
	// Определяем цвет и иконку в зависимости от уровня
	var color, icon, levelStr string
	switch entry.Level {
	case logrus.PanicLevel, logrus.FatalLevel:
		color = colorBrightRed
		icon = "💥"
		levelStr = "FATAL"
	case logrus.ErrorLevel:
		color = colorRed
		icon = "❌"
		levelStr = "ERROR"
	case logrus.WarnLevel:
		color = colorYellow
		icon = "⚠️"
		levelStr = "WARN"
	case logrus.InfoLevel:
		// Проверяем тип для success
		if entry.Data["type"] == "success" {
			color = colorGreen
			icon = "✅"
			levelStr = "SUCCESS"
		} else {
			color = colorCyan
			icon = "ℹ️"
			levelStr = "INFO"
		}
	case logrus.DebugLevel:
		color = colorGray
		icon = "🔍"
		levelStr = "DEBUG"
	default:
		color = colorWhite
		icon = "📝"
		levelStr = "LOG"
	}
	
	// Формируем строку операции с цветом
	opColor := colorPurple
	opParts := strings.Split(operation, ".")
	if len(opParts) > 0 {
		// Последняя часть операции (метод) выделяем ярче
		opFormatted := strings.Join(opParts[:len(opParts)-1], ".") + 
			colorBrightCyan + "." + opParts[len(opParts)-1] + opColor
		operation = opFormatted
	}
	
	// Строим основную строку лога
	b.WriteString(fmt.Sprintf("%s[%s]%s ", colorGray, timestamp, colorReset))
	b.WriteString(fmt.Sprintf("%s%s%s ", color, icon, colorReset))
	b.WriteString(fmt.Sprintf("%s[%s]%s ", color, levelStr, colorReset))
	b.WriteString(fmt.Sprintf("%s%s%s ", opColor, operation, colorReset))
	b.WriteString(fmt.Sprintf("%s%s%s", colorWhite, entry.Message, colorReset))
	
	// Добавляем поля (кроме operation и type, которые уже обработаны)
	if len(entry.Data) > 0 {
		fields := []string{}
		for k, v := range entry.Data {
			if k != "operation" && k != "type" {
				fields = append(fields, fmt.Sprintf("%s%s%s=%v", colorGray, k, colorReset, v))
			}
		}
		if len(fields) > 0 {
			b.WriteString(fmt.Sprintf(" %s|%s %s", colorGray, colorReset, strings.Join(fields, " ")))
		}
	}
	
	// Добавляем ошибку, если есть
	if entry.Data["error"] != nil {
		errMsg := fmt.Sprintf("%v", entry.Data["error"])
		b.WriteString(fmt.Sprintf(" %s→%s %s%s%s", colorRed, colorReset, colorRed, errMsg, colorReset))
	}
	
	// Добавляем информацию о файле и строке для ошибок и fatal
	if entry.Level <= logrus.ErrorLevel {
		if entry.Caller != nil {
			file := entry.Caller.File
			line := entry.Caller.Line
			// Получаем только имя файла
			parts := strings.Split(file, "/")
			fileName := parts[len(parts)-1]
			b.WriteString(fmt.Sprintf(" %s[%s:%d]%s", colorGray, fileName, line, colorReset))
		}
	}
	
	b.WriteString("\n")
	
	return []byte(b.String()), nil
}

// GetCallerInfo получает информацию о вызывающем коде
func GetCallerInfo(skip int) (string, int) {
	_, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return "unknown", 0
	}
	parts := strings.Split(file, "/")
	fileName := parts[len(parts)-1]
	return fileName, line
}
