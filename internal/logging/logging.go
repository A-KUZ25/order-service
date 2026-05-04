package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/syslog"
	"os"
	"strings"
	"sync"
	"time"
)

type contextKey string

const requestIDKey contextKey = "request_id"

var defaultLogger = newLogger("orders-service")

type logger struct {
	service string
	stdout  bool
	syslog  *syslog.Writer
	mu      sync.Mutex
}

func Init(service string) {
	defaultLogger = newLogger(service)
}

func newLogger(service string) *logger {
	target := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_TARGET")))
	if target == "" {
		if strings.TrimSpace(os.Getenv("SYSLOG_IDENTITY")) != "" {
			target = "both"
		} else {
			target = "stdout"
		}
	}

	l := &logger{
		service: service,
		stdout:  target == "stdout" || target == "both",
	}

	if target == "syslog" || target == "both" {
		identity := strings.TrimSpace(os.Getenv("SYSLOG_IDENTITY"))
		if identity == "" {
			identity = service
		}
		writer, err := syslog.New(syslog.LOG_INFO|syslog.LOG_USER, identity)
		if err != nil {
			log.Printf("syslog init failed: %v", err)
			l.stdout = true
		} else {
			l.syslog = writer
		}
	}

	return l
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	if strings.TrimSpace(requestID) == "" {
		return ctx
	}
	return context.WithValue(ctx, requestIDKey, requestID)
}

func RequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	value, _ := ctx.Value(requestIDKey).(string)
	return value
}

func Info(ctx context.Context, message string, args ...any) {
	defaultLogger.write(ctx, "info", message, nil, args...)
}

func Warn(ctx context.Context, message string, args ...any) {
	defaultLogger.write(ctx, "warning", message, nil, args...)
}

func Error(ctx context.Context, message string, err error, args ...any) {
	defaultLogger.write(ctx, "error", message, err, args...)
}

func (l *logger) write(ctx context.Context, level, message string, err error, args ...any) {
	entry := map[string]any{
		"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		"level":     level,
		"service":   l.service,
		"message":   message,
	}

	if requestID := RequestID(ctx); requestID != "" {
		entry["request_id"] = requestID
	}
	if err != nil {
		entry["error"] = err.Error()
	}

	for i := 0; i+1 < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok || key == "" {
			continue
		}
		entry[key] = args[i+1]
	}

	line, marshalErr := json.Marshal(entry)
	if marshalErr != nil {
		line = []byte(fmt.Sprintf(`{"level":"error","service":%q,"message":"log marshal failed","error":%q}`, l.service, marshalErr.Error()))
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.stdout {
		_, _ = os.Stdout.Write(append(line, '\n'))
	}
	if l.syslog != nil {
		switch level {
		case "error":
			_ = l.syslog.Err(string(line))
		case "warning":
			_ = l.syslog.Warning(string(line))
		default:
			_ = l.syslog.Info(string(line))
		}
	}
}
