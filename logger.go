package steam

// Logger is the minimal logging contract used by the SDK.
type Logger interface {
	Debug(msg string, args ...any)
	Error(msg string, args ...any)
}

type noopLogger struct{}

func (noopLogger) Debug(string, ...any) {}

func (noopLogger) Error(string, ...any) {}
