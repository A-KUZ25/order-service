package logging

import (
	"context"
	"encoding/json"
	"fmt"
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
	service       string
	target        string
	identity      string
	stdout        bool
	syslog        *syslog.Writer
	syslogInitErr error
	mu            sync.Mutex
}

func Init(service string) {
	defaultLogger = newLogger(service)
	Info(context.Background(), "logging initialized",
		"log_target", defaultLogger.target,
		"stdout_enabled", defaultLogger.stdout,
		"syslog_enabled", defaultLogger.syslog != nil,
		"syslog_identity", defaultLogger.identity,
		"syslog_error", errorString(defaultLogger.syslogInitErr),
	)
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
		service:  service,
		target:   target,
		stdout:   target == "stdout" || target == "both",
		identity: strings.TrimSpace(os.Getenv("SYSLOG_IDENTITY")),
	}

	if target == "syslog" || target == "both" {
		if l.identity == "" {
			l.identity = service
		}
		writer, err := syslog.New(syslog.LOG_INFO|syslog.LOG_USER, l.identity)
		if err != nil {
			l.syslogInitErr = err
			l.stdout = true
		} else {
			l.syslog = writer
		}
	}

	return l
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
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
