package profiling

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grafana/pyroscope-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordingLogger struct {
	infos  []string
	debugs []string
	errors []string
}

func (l *recordingLogger) Infof(format string, args ...any) {
	l.infos = append(l.infos, format)
}

func (l *recordingLogger) Debugf(format string, args ...any) {
	l.debugs = append(l.debugs, format)
}

func (l *recordingLogger) Errorf(format string, args ...any) {
	l.errors = append(l.errors, format)
}

func TestNewFromConfigDisabled(t *testing.T) {
	profiler, err := NewFromConfig("svc", Config{})
	require.NoError(t, err)
	assert.Nil(t, profiler)
}

func TestNewFromConfigStartsProfiler(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := &recordingLogger{}
	profiler, err := NewFromConfig("svc", Config{
		Enabled:           true,
		ServerAddress:     server.URL,
		Namespace:         "homelab",
		BasicAuthUser:     "user",
		BasicAuthPassword: "pass",
		TenantID:          "tenant",
	}, WithLogger(logger), WithProfileTypes(pyroscope.ProfileGoroutines))
	require.NoError(t, err)
	require.NotNil(t, profiler)
	require.NoError(t, Shutdown(profiler))
}

func TestNewReturnsStartError(t *testing.T) {
	profiler, err := New("svc", "://bad-url", WithProfileTypes(pyroscope.ProfileGoroutines))
	require.Error(t, err)
	assert.Nil(t, profiler)
	assert.ErrorContains(t, err, "profiling: start pyroscope")
}

func TestShutdownNilProfiler(t *testing.T) {
	assert.NoError(t, Shutdown(nil))
}

func TestOptions(t *testing.T) {
	logger := &recordingLogger{}
	cfg := pyroscope.Config{}

	WithNamespace("homelab")(&cfg)
	WithTags(map[string]string{
		"service_name": "override",
		"region":       "ap-southeast-1",
	})(&cfg)
	WithProfileTypes(pyroscope.ProfileGoroutines)(&cfg)
	WithLogger(logger)(&cfg)
	WithBasicAuth("user", "pass")(&cfg)
	WithTenantID("tenant")(&cfg)

	assert.Equal(t, "homelab", cfg.Tags["namespace"])
	assert.Equal(t, "override", cfg.Tags["service_name"])
	assert.Equal(t, "ap-southeast-1", cfg.Tags["region"])
	assert.Equal(t, []pyroscope.ProfileType{pyroscope.ProfileGoroutines}, cfg.ProfileTypes)
	assert.Same(t, logger, cfg.Logger)
	assert.Equal(t, "user", cfg.BasicAuthUser)
	assert.Equal(t, "pass", cfg.BasicAuthPassword)
	assert.Equal(t, "tenant", cfg.TenantID)
}

func TestOptionsInitializeNilTags(t *testing.T) {
	cfg := pyroscope.Config{}
	WithNamespace("homelab")(&cfg)
	assert.Equal(t, "homelab", cfg.Tags["namespace"])

	cfg = pyroscope.Config{}
	WithTags(map[string]string{"region": "ap-southeast-1"})(&cfg)
	assert.Equal(t, "ap-southeast-1", cfg.Tags["region"])
}

func TestSlogLogger(t *testing.T) {
	var out bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&out, nil)))
	defer slog.SetDefault(previous)

	logger := slogLogger{}
	logger.Infof("hello %s", "info")
	logger.Debugf("hello %s", "debug")
	logger.Errorf("hello %s", "error")

	data, err := io.ReadAll(&out)
	require.NoError(t, err)
	text := string(data)
	assert.Contains(t, text, "hello info")
	assert.Contains(t, text, "hello error")
}
