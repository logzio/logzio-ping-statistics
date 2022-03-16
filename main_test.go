package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/golang/snappy"
	"github.com/jarcoal/httpmock"
	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getMetrics(request *http.Request) ([]map[string]interface{}, error) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}

	defer func(body io.ReadCloser) {
		if err = body.Close(); err != nil {
			panic(err)
		}

	}(request.Body)

	uncompressedBody, err := snappy.Decode(nil, body)
	if err != nil {
		return nil, err
	}

	writeRequest := &prompb.WriteRequest{}
	if err = writeRequest.Unmarshal(uncompressedBody); err != nil {
		return nil, err
	}

	metrics := make([]map[string]interface{}, 0)

	for _, timeseries := range writeRequest.Timeseries {
		metric := make(map[string]interface{})

		metric["value"] = timeseries.Samples[0].Value

		for _, label := range timeseries.Labels {
			metric[label.Name] = label.Value
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func TestNewLogzioPingStatistics_Success(t *testing.T) {
	err := os.Setenv(addressesEnvName, "www.google.com,https://listener.logz.io:8053,tcp://www.nytimes.com")
	require.NoError(t, err)

	err = os.Setenv(pingCountEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(pingIntervalEnvName, "1")
	require.NoError(t, err)

	err = os.Setenv(pingTimeoutEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	logzioPingStats, err := newLogzioPingStatistics(context.Background())
	require.NoError(t, err)
	require.NotNil(t, logzioPingStats)

	assert.Equal(t, []string{"www.google.com:80", "listener.logz.io:8053", "www.nytimes.com:80"}, logzioPingStats.addresses)
	assert.Equal(t, 10, logzioPingStats.pingCount)
	assert.Equal(t, 1*time.Second, logzioPingStats.pingInterval)
	assert.Equal(t, 10*time.Second, logzioPingStats.pingTimeout)
	assert.Equal(t, "https://listener.logz.io:8053", logzioPingStats.logzioMetricsListener)
	assert.Equal(t, "123456789a", logzioPingStats.logzioMetricsToken)

	os.Clearenv()
}

func TestNewLogzioPingStatistics_NoAddress(t *testing.T) {
	err := os.Setenv(pingCountEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(pingIntervalEnvName, "1")
	require.NoError(t, err)

	err = os.Setenv(pingTimeoutEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	_, err = newLogzioPingStatistics(context.Background())
	require.Error(t, err)

	os.Clearenv()
}

func TestNewLogzioPingStatistics_NoPingCount(t *testing.T) {
	err := os.Setenv(addressesEnvName, "www.google.com,https://listener.logz.io:8053,tcp://www.nytimes.com")
	require.NoError(t, err)

	err = os.Setenv(pingIntervalEnvName, "1")
	require.NoError(t, err)

	err = os.Setenv(pingTimeoutEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	_, err = newLogzioPingStatistics(context.Background())
	require.Error(t, err)

	os.Clearenv()
}

func TestNewLogzioPingStatistics_NoPingCountNumber(t *testing.T) {
	err := os.Setenv(addressesEnvName, "www.google.com,https://listener.logz.io:8053,tcp://www.nytimes.com")
	require.NoError(t, err)

	err = os.Setenv(pingCountEnvName, "pingCount")
	require.NoError(t, err)

	err = os.Setenv(pingIntervalEnvName, "1")
	require.NoError(t, err)

	err = os.Setenv(pingTimeoutEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	_, err = newLogzioPingStatistics(context.Background())
	require.Error(t, err)

	os.Clearenv()
}

func TestNewLogzioPingStatistics_NoPingCountPositiveNumber(t *testing.T) {
	err := os.Setenv(addressesEnvName, "www.google.com,https://listener.logz.io:8053,tcp://www.nytimes.com")
	require.NoError(t, err)

	err = os.Setenv(pingCountEnvName, "0")
	require.NoError(t, err)

	err = os.Setenv(pingIntervalEnvName, "1")
	require.NoError(t, err)

	err = os.Setenv(pingTimeoutEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	_, err = newLogzioPingStatistics(context.Background())
	require.Error(t, err)

	os.Clearenv()
}

func TestNewLogzioPingStatistics_NoPingInterval(t *testing.T) {
	err := os.Setenv(addressesEnvName, "www.google.com,https://listener.logz.io:8053,tcp://www.nytimes.com")
	require.NoError(t, err)

	err = os.Setenv(pingCountEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(pingTimeoutEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	_, err = newLogzioPingStatistics(context.Background())
	require.Error(t, err)

	os.Clearenv()
}

func TestNewLogzioPingStatistics_NoPingIntervalNumber(t *testing.T) {
	err := os.Setenv(addressesEnvName, "www.google.com,https://listener.logz.io:8053,tcp://www.nytimes.com")
	require.NoError(t, err)

	err = os.Setenv(pingCountEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(pingIntervalEnvName, "pingInterval")
	require.NoError(t, err)

	err = os.Setenv(pingTimeoutEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	_, err = newLogzioPingStatistics(context.Background())
	require.Error(t, err)

	os.Clearenv()
}

func TestNewLogzioPingStatistics_NoPingIntervalPositiveNumber(t *testing.T) {
	err := os.Setenv(addressesEnvName, "www.google.com,https://listener.logz.io:8053,tcp://www.nytimes.com")
	require.NoError(t, err)

	err = os.Setenv(pingCountEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(pingIntervalEnvName, "0")
	require.NoError(t, err)

	err = os.Setenv(pingTimeoutEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	_, err = newLogzioPingStatistics(context.Background())
	require.Error(t, err)

	os.Clearenv()
}

func TestNewLogzioPingStatistics_NoPingTimeout(t *testing.T) {
	err := os.Setenv(addressesEnvName, "www.google.com,https://listener.logz.io:8053,tcp://www.nytimes.com")
	require.NoError(t, err)

	err = os.Setenv(pingCountEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(pingIntervalEnvName, "1")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	_, err = newLogzioPingStatistics(context.Background())
	require.Error(t, err)

	os.Clearenv()
}

func TestNewLogzioPingStatistics_NoPingTimeoutNumber(t *testing.T) {
	err := os.Setenv(addressesEnvName, "www.google.com,https://listener.logz.io:8053,tcp://www.nytimes.com")
	require.NoError(t, err)

	err = os.Setenv(pingCountEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(pingIntervalEnvName, "1")
	require.NoError(t, err)

	err = os.Setenv(pingTimeoutEnvName, "pingTimeout")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	_, err = newLogzioPingStatistics(context.Background())
	require.Error(t, err)

	os.Clearenv()
}

func TestNewLogzioPingStatistics_NoPingTimeoutPositiveNumber(t *testing.T) {
	err := os.Setenv(addressesEnvName, "www.google.com,https://listener.logz.io:8053,tcp://www.nytimes.com")
	require.NoError(t, err)

	err = os.Setenv(pingCountEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(pingIntervalEnvName, "1")
	require.NoError(t, err)

	err = os.Setenv(pingTimeoutEnvName, "0")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	_, err = newLogzioPingStatistics(context.Background())
	require.Error(t, err)

	os.Clearenv()
}

func TestNewLogzioPingStatistics_NoLogzioMetricsListener(t *testing.T) {
	err := os.Setenv(addressesEnvName, "www.google.com,https://listener.logz.io:8053,tcp://www.nytimes.com")
	require.NoError(t, err)

	err = os.Setenv(pingCountEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(pingIntervalEnvName, "1")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	_, err = newLogzioPingStatistics(context.Background())
	require.Error(t, err)

	os.Clearenv()
}

func TestNewLogzioPingStatistics_NoLogzioMetricsToken(t *testing.T) {
	err := os.Setenv(addressesEnvName, "www.google.com,https://listener.logz.io:8053,tcp://www.nytimes.com")
	require.NoError(t, err)

	err = os.Setenv(pingCountEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(pingIntervalEnvName, "1")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	_, err = newLogzioPingStatistics(context.Background())
	require.Error(t, err)

	os.Clearenv()
}

func TestGetAllAddressesPingStatistics_Success(t *testing.T) {
	logzioPingStats := &logzioPingStatistics{
		ctx:                   context.Background(),
		logzioMetricsListener: "https://listener.logz.io:8053",
		logzioMetricsToken:    "123456789a",
		addresses:             []string{"www.google.com:80", "listener.logz.io:8053"},
		pingCount:             10,
		pingInterval:          1 * time.Second,
		pingTimeout:           10 * time.Second,
	}

	err := logzioPingStats.getAllAddressesPingStatistics()
	require.NoError(t, err)

	assert.Len(t, logzioPingStats.pingsStats, 2)

	for _, pingStats := range logzioPingStats.pingsStats {
		assert.NotNil(t, pingStats)
		assert.Len(t, pingStats.rtts, 10)

		for _, rtt := range pingStats.rtts {
			assert.NotEmpty(t, rtt)
		}

		assert.Equal(t, 10, pingStats.probesSent)
		assert.Equal(t, 10, pingStats.successfulProbes)
		assert.Equal(t, 0, pingStats.probesFailed)
		assert.Contains(t, logzioPingStats.addresses, pingStats.address)
		assert.NotEmpty(t, pingStats.addressIP)
	}
}

func TestCreateController_Success(t *testing.T) {
	logzioPingStats := &logzioPingStatistics{
		ctx:                   context.Background(),
		logzioMetricsListener: "https://listener.logz.io:8053",
		logzioMetricsToken:    "123456789a",
		addresses:             []string{"www.google.com:80", "listener.logz.io:8053"},
		pingCount:             10,
		pingInterval:          1 * time.Second,
		pingTimeout:           10 * time.Second,
	}

	cont, err := logzioPingStats.createController()
	require.NoError(t, err)
	require.NotNil(t, cont)
}

func TestCollectMetrics_Success(t *testing.T) {
	err := os.Setenv(awsRegionEnvName, "us-east-1")
	require.NoError(t, err)

	err = os.Setenv(awsLambdaFunctionNameEnvName, "test")
	require.NoError(t, err)

	logzioPingStats := &logzioPingStatistics{
		ctx:                   context.Background(),
		logzioMetricsListener: "https://listener.logz.io:8053",
		logzioMetricsToken:    "123456789a",
		addresses:             []string{"www.google.com:80", "listener.logz.io:8053"},
		pingCount:             3,
		pingInterval:          1 * time.Second,
		pingTimeout:           10 * time.Second,
	}

	err = logzioPingStats.getAllAddressesPingStatistics()
	require.NoError(t, err)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://listener.logz.io:8053",
		func(request *http.Request) (*http.Response, error) {
			metrics, err := getMetrics(request)
			require.NoError(t, err)
			require.NotNil(t, metrics)

			assert.Len(t, metrics, 12)

			for _, metric := range metrics {
				assert.Contains(t, []string{rttMetricName, probesSentMetricName, successfulProbesMetricName, probesFailedMetricName}, metric["__name__"])

				if metric["__name__"] == rttMetricName {
					assert.Len(t, metric, 9)
					assert.NotEmpty(t, metric["value"])
					assert.Contains(t, []string{"1", "2", "3"}, metric[rttMetricRttIndexLabelName])
					assert.Equal(t, "3", metric[rttMetricTotalRttsLabelName])
					assert.Equal(t, rttMetricUnitLabelValue, metric[unitLabelName])
				} else if metric["__name__"] == probesSentMetricName {
					assert.Len(t, metric, 6)
					assert.Equal(t, float64(3), metric["value"])
				} else if metric["__name__"] == successfulProbesMetricName {
					assert.Len(t, metric, 6)
					assert.Equal(t, float64(3), metric["value"])
				} else if metric["__name__"] == probesFailedMetricName {
					assert.Len(t, metric, 6)
					assert.Equal(t, float64(0), metric["value"])
				}

				assert.Contains(t, logzioPingStats.addresses, metric[addressLabelName])
				assert.NotEmpty(t, metric[ipLabelName])
				assert.Equal(t, "us-east-1", metric[awsRegionLabelName])
				assert.Equal(t, "test", metric[awsLambdaFunctionLabelName])
			}

			return httpmock.NewStringResponse(200, ""), nil
		})

	err = logzioPingStats.collectMetrics()
	require.NoError(t, err)

	os.Clearenv()
}

func TestRun_Success(t *testing.T) {
	err := os.Setenv(awsRegionEnvName, "us-east-1")
	require.NoError(t, err)

	err = os.Setenv(awsLambdaFunctionNameEnvName, "test")
	require.NoError(t, err)

	err = os.Setenv(addressesEnvName, "www.google.com,https://listener.logz.io:8053")
	require.NoError(t, err)

	err = os.Setenv(pingCountEnvName, "3")
	require.NoError(t, err)

	err = os.Setenv(pingIntervalEnvName, "1")
	require.NoError(t, err)

	err = os.Setenv(pingTimeoutEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://listener.logz.io:8053",
		func(request *http.Request) (*http.Response, error) {
			metrics, err := getMetrics(request)
			require.NoError(t, err)
			require.NotNil(t, metrics)

			assert.Len(t, metrics, 12)

			for _, metric := range metrics {
				assert.Contains(t, []string{rttMetricName, probesSentMetricName, successfulProbesMetricName, probesFailedMetricName}, metric["__name__"])

				if metric["__name__"] == rttMetricName {
					assert.Len(t, metric, 9)
					assert.NotEmpty(t, metric["value"])
					assert.Contains(t, []string{"1", "2", "3"}, metric[rttMetricRttIndexLabelName])
					assert.Equal(t, "3", metric[rttMetricTotalRttsLabelName])
					assert.Equal(t, rttMetricUnitLabelValue, metric[unitLabelName])
				} else if metric["__name__"] == probesSentMetricName {
					assert.Len(t, metric, 6)
					assert.Equal(t, float64(3), metric["value"])
				} else if metric["__name__"] == successfulProbesMetricName {
					assert.Len(t, metric, 6)
					assert.Equal(t, float64(3), metric["value"])
				} else if metric["__name__"] == probesFailedMetricName {
					assert.Len(t, metric, 6)
					assert.Equal(t, float64(0), metric["value"])
				}

				assert.Contains(t, []string{"www.google.com:80", "listener.logz.io:8053"}, metric[addressLabelName])
				assert.NotEmpty(t, metric[ipLabelName])
				assert.Equal(t, "us-east-1", metric[awsRegionLabelName])
				assert.Equal(t, "test", metric[awsLambdaFunctionLabelName])
			}

			return httpmock.NewStringResponse(200, ""), nil
		})

	err = run(context.Background())
	require.NoError(t, err)

	os.Clearenv()
}
