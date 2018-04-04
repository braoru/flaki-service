package health

//go:generate mockgen -destination=./mock/sentry.go -package=mock -mock_names=SentryClient=SentryClient  github.com/cloudtrust/flaki-service/pkg/health SentryClient

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// SentryModule is the health check module for sentry.
type SentryModule struct {
	sentry     sentryClient
	httpClient sentryHTTPClient
	enabled    bool
}

// sentryClient is the interface of the sentry client.
type sentryClient interface {
	URL() string
}

// sentryHTTPClient is the interface of the http client.
type sentryHTTPClient interface {
	Get(string) (*http.Response, error)
}

// NewSentryModule returns the sentry health module.
func NewSentryModule(sentry sentryClient, httpClient sentryHTTPClient, enabled bool) *SentryModule {
	return &SentryModule{
		sentry:     sentry,
		httpClient: httpClient,
		enabled:    enabled,
	}
}

// SentryReport is the health report returned by the sentry module.
type SentryReport struct {
	Name     string
	Duration time.Duration
	Status   Status
	Error    error
}

// HealthChecks executes all health checks for Sentry.
func (m *SentryModule) HealthChecks(context.Context) []SentryReport {
	var reports = []SentryReport{}
	reports = append(reports, m.sentryPingCheck())
	return reports
}

func (m *SentryModule) sentryPingCheck() SentryReport {
	var healthCheckName = "ping"

	if !m.enabled {
		return SentryReport{
			Name:   healthCheckName,
			Status: Deactivated,
		}
	}

	var dsn = m.sentry.URL()

	// Get Sentry health status.
	var now = time.Now()
	var err = pingSentry(dsn, m.httpClient)
	var duration = time.Since(now)

	var hcErr error
	var s Status
	switch {
	case err != nil:
		hcErr = errors.Wrap(err, "could not ping sentry")
		s = KO
	default:
		s = OK
	}

	return SentryReport{
		Name:     healthCheckName,
		Duration: duration,
		Status:   s,
		Error:    hcErr,
	}
}

func pingSentry(dsn string, httpClient sentryHTTPClient) error {

	// Build sentry health url from sentry dsn. The health url is <sentryURL>/_health
	var url string
	if idx := strings.LastIndex(dsn, "/api/"); idx != -1 {
		url = fmt.Sprintf("%s/_health", dsn[:idx])
	}

	// Query sentry health endpoint.
	var res *http.Response
	{
		var err error
		res, err = httpClient.Get(url)
		if err != nil {
			return err
		}
		if res != nil {
			defer res.Body.Close()
		}
	}

	// Chesk response status.
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("http response status code: %v", res.Status)
	}

	// Chesk response body. The sentry health endpoint returns "ok" when there is no issue.
	var response []byte
	{
		var err error
		response, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
	}

	if strings.Compare(string(response), "ok") == 0 {
		return nil
	}

	return fmt.Errorf("response should be 'ok' but is: %v", string(response))
}
