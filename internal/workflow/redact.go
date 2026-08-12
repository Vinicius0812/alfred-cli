package workflow

import (
	"io"
	"strings"
	"sync"
)

func redact(value string, secrets []string) string {
	for _, secret := range secrets {
		if secret != "" {
			value = strings.ReplaceAll(value, secret, "[REDACTED]")
		}
	}
	return value
}

type redactingWriter struct {
	mu      sync.Mutex
	target  io.Writer
	secrets []string
	buffer  string
}

func newRedactingWriter(target io.Writer, secrets []string) *redactingWriter {
	return &redactingWriter{target: target, secrets: secrets}
}

func (w *redactingWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.buffer += string(data)
	redacted := redact(w.buffer, w.secrets)
	keep := partialSecretSuffixLength(redacted, w.secrets)
	flushLength := len(redacted) - keep
	if flushLength == 0 {
		return len(data), nil
	}
	if _, err := io.WriteString(w.target, redacted[:flushLength]); err != nil {
		return 0, err
	}
	w.buffer = redacted[flushLength:]
	return len(data), nil
}

func (w *redactingWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.buffer == "" {
		return nil
	}
	_, err := io.WriteString(w.target, redact(w.buffer, w.secrets))
	w.buffer = ""
	return err
}

func partialSecretSuffixLength(value string, secrets []string) int {
	longest := 0
	for _, secret := range secrets {
		for length := 1; length < len(secret) && length <= len(value); length++ {
			if length > longest && strings.HasSuffix(value, secret[:length]) {
				longest = length
			}
		}
	}
	return longest
}
