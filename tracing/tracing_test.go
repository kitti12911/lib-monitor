package tracing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
)

func TestNewFromConfigDisabled(t *testing.T) {
	tp, err := NewFromConfig(context.Background(), "svc", Config{})
	require.NoError(t, err)
	assert.Nil(t, tp)
}

func TestNewFromConfigGRPC(t *testing.T) {
	tp, err := NewFromConfig(context.Background(), "svc", Config{
		Enabled:     true,
		Endpoint:    "localhost:4317",
		Protocol:    ProtocolGRPC,
		Insecure:    true,
		SampleRatio: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, tp)
	assert.Same(t, tp, otel.GetTracerProvider())
	require.NoError(t, Shutdown(context.Background(), tp))
}

func TestNewFromConfigHTTP(t *testing.T) {
	tp, err := NewFromConfig(context.Background(), "svc", Config{
		Enabled:     true,
		Endpoint:    "localhost:4318",
		Protocol:    ProtocolHTTP,
		Insecure:    true,
		SampleRatio: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, tp)
	assert.Same(t, tp, otel.GetTracerProvider())
	require.NoError(t, Shutdown(context.Background(), tp))
}

func TestNewUsesGRPCDefaults(t *testing.T) {
	tp, err := New(context.Background(), "svc", "localhost:4317")
	require.NoError(t, err)
	require.NotNil(t, tp)
	require.NoError(t, Shutdown(context.Background(), tp))
}

func TestNewFromConfigExporterError(t *testing.T) {
	_, err := NewFromConfig(context.Background(), "svc", Config{
		Enabled:  true,
		Endpoint: "localhost:4317",
		Protocol: "zipkin",
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, `unsupported protocol "zipkin"`)
}

func TestNewExporterDefaultsToGRPC(t *testing.T) {
	exporter, err := newExporter(context.Background(), Config{
		Endpoint: "localhost:4317",
		Insecure: true,
	})
	require.NoError(t, err)
	require.NotNil(t, exporter)
	_ = exporter.Shutdown(context.Background())
}

func TestNewExporterRejectsUnsupportedProtocol(t *testing.T) {
	_, err := newExporter(context.Background(), Config{
		Endpoint: "localhost:4317",
		Protocol: "zipkin",
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, `unsupported protocol "zipkin"`)
}

func TestNewGRPCExporterError(t *testing.T) {
	_, err := newGRPCExporter(context.Background(), Config{
		Endpoint: "\n",
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "tracing: create exporter")
}

func TestNewHTTPExporterError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := newHTTPExporter(ctx, Config{
		Endpoint: "localhost:4318",
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "tracing: create exporter")
}

func TestShutdownNilProvider(t *testing.T) {
	assert.NoError(t, Shutdown(context.Background(), nil))
}
