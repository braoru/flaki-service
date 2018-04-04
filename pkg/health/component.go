package health

//go:generate mockgen -destination=./mock/module.go -package=mock -mock_names=InfluxHealthChecker=InfluxHealthChecker,JaegerHealthChecker=JaegerHealthChecker,RedisHealthChecker=RedisHealthChecker,SentryHealthChecker=SentryHealthChecker  github.com/cloudtrust/flaki-service/pkg/health InfluxHealthChecker,JaegerHealthChecker,RedisHealthChecker,SentryHealthChecker

import (
	"context"
)

// Status is the status of the health check.
type Status int

const (
	// OK is the status for a successful health check.
	OK Status = iota
	// KO is the status for an unsuccessful health check.
	KO
	// Degraded is the status for a degraded service, e.g. the service still works, but the metrics DB is KO.
	Degraded
	// Deactivated is the status for a service that is deactivated, e.g. we can disable error tracking, instrumenting, tracing,...
	Deactivated
)

func (s Status) String() string {
	var names = []string{"OK", "KO", "Degraded", "Deactivated"}

	if s < OK || s > Deactivated {
		return "Unknown"
	}

	return names[s]
}

// InfluxHealthChecker is the interface of the influx health check module.
type InfluxHealthChecker interface {
	HealthChecks(context.Context) []InfluxReport
}

// JaegerHealthChecker is the interface of the jaeger health check module.
type JaegerHealthChecker interface {
	HealthChecks(context.Context) []JaegerReport
}

// RedisHealthChecker is the interface of the redis health check module.
type RedisHealthChecker interface {
	HealthChecks(context.Context) []RedisReport
}

// SentryHealthChecker is the interface of the sentry health check module.
type SentryHealthChecker interface {
	HealthChecks(context.Context) []SentryReport
}

// Component is the Health component.
type Component struct {
	influx InfluxHealthChecker
	jaeger JaegerHealthChecker
	redis  RedisHealthChecker
	sentry SentryHealthChecker
}

// NewComponent returns the health component.
func NewComponent(influx InfluxHealthChecker, jaeger JaegerHealthChecker, redis RedisHealthChecker, sentry SentryHealthChecker) *Component {
	return &Component{
		influx: influx,
		jaeger: jaeger,
		redis:  redis,
		sentry: sentry,
	}
}

// Report contains the result of one health test.
type Report struct {
	Name     string
	Duration string
	Status   string
	Error    string
}

// InfluxHealthChecks uses the health component to test the Influx health.
func (c *Component) InfluxHealthChecks(ctx context.Context) []Report {
	var reports = c.influx.HealthChecks(ctx)
	var out = []Report{}
	for _, r := range reports {
		out = append(out, Report{
			Name:     r.Name,
			Duration: r.Duration.String(),
			Status:   r.Status.String(),
			Error:    err(r.Error),
		})
	}
	return out
}

// JaegerHealthChecks uses the health component to test the Jaeger health.
func (c *Component) JaegerHealthChecks(ctx context.Context) []Report {
	var reports = c.jaeger.HealthChecks(ctx)
	var out = []Report{}
	for _, r := range reports {
		out = append(out, Report{
			Name:     r.Name,
			Duration: r.Duration.String(),
			Status:   r.Status.String(),
			Error:    err(r.Error),
		})
	}
	return out
}

// RedisHealthChecks uses the health component to test the Redis health.
func (c *Component) RedisHealthChecks(ctx context.Context) []Report {
	var reports = c.redis.HealthChecks(ctx)
	var out = []Report{}
	for _, r := range reports {
		out = append(out, Report{
			Name:     r.Name,
			Duration: r.Duration.String(),
			Status:   r.Status.String(),
			Error:    err(r.Error),
		})
	}
	return out
}

// SentryHealthChecks uses the health component to test the Sentry health.
func (c *Component) SentryHealthChecks(ctx context.Context) []Report {
	var reports = c.sentry.HealthChecks(ctx)
	var out = []Report{}
	for _, r := range reports {
		out = append(out, Report{
			Name:     r.Name,
			Duration: r.Duration.String(),
			Status:   r.Status.String(),
			Error:    err(r.Error),
		})
	}
	return out
}

// AllHealthChecks call all component checks and build a general health report.
func (c *Component) AllHealthChecks(ctx context.Context) map[string]string {
	var reports = map[string]string{}

	reports["influx"] = determineStatus(c.InfluxHealthChecks(ctx))
	reports["jaeger"] = determineStatus(c.JaegerHealthChecks(ctx))
	reports["redis"] = determineStatus(c.RedisHealthChecks(ctx))
	reports["sentry"] = determineStatus(c.SentryHealthChecks(ctx))

	return reports
}

// err return the string error that will be in the health report
func err(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// determineStatus parse all the tests reports and output a global status.
func determineStatus(reports []Report) string {
	var degraded = false
	for _, r := range reports {
		switch r.Status {
		case Deactivated.String():
			// If the status is Deactivated, we do not need to go through all tests reports, all
			// status will be the same.
			return Deactivated.String()
		case KO.String():
			return KO.String()
		case Degraded.String():
			degraded = true
		}
	}
	if degraded {
		return Degraded.String()
	}
	return OK.String()
}
