package profiling

import (
	"fmt"
	"maps"

	"github.com/grafana/pyroscope-go"
)

type Option func(*pyroscope.Config)

func New(serviceName, serverAddress string, opts ...Option) (*pyroscope.Profiler, error) {
	cfg := pyroscope.Config{
		ApplicationName: serviceName,
		ServerAddress:   serverAddress,
		Logger:          pyroscope.StandardLogger,
		Tags: map[string]string{
			"service_name": serviceName,
		},
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
			pyroscope.ProfileGoroutines,
		},
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	profiler, err := pyroscope.Start(cfg)
	if err != nil {
		return nil, fmt.Errorf("profiling: start pyroscope: %w", err)
	}

	return profiler, nil
}

func Shutdown(profiler *pyroscope.Profiler) error {
	if profiler == nil {
		return nil
	}

	return profiler.Stop()
}

func WithNamespace(namespace string) Option {
	return func(cfg *pyroscope.Config) {
		if cfg.Tags == nil {
			cfg.Tags = make(map[string]string)
		}

		cfg.Tags["namespace"] = namespace
	}
}

func WithTags(tags map[string]string) Option {
	return func(cfg *pyroscope.Config) {
		if cfg.Tags == nil {
			cfg.Tags = make(map[string]string, len(tags))
		}

		maps.Copy(cfg.Tags, tags)
	}
}

func WithProfileTypes(types ...pyroscope.ProfileType) Option {
	return func(cfg *pyroscope.Config) {
		cfg.ProfileTypes = types
	}
}

func WithLogger(logger pyroscope.Logger) Option {
	return func(cfg *pyroscope.Config) {
		cfg.Logger = logger
	}
}

func WithBasicAuth(user, password string) Option {
	return func(cfg *pyroscope.Config) {
		cfg.BasicAuthUser = user
		cfg.BasicAuthPassword = password
	}
}

func WithTenantID(tenantID string) Option {
	return func(cfg *pyroscope.Config) {
		cfg.TenantID = tenantID
	}
}
