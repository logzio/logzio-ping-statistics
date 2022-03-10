package main

import (
	"context"
	"github.com/golang/snappy"
	"github.com/jarcoal/httpmock"
	"github.com/prometheus/prometheus/prompb"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogzioPingStatistics_Success(t *testing.T) {
	err := os.Setenv(addressesEnvName, "www.google.com,https://listener.logz.io:8053")
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

	assert.Equal(t, []string{"www.google.com:80", "listener.logz.io:8053"}, logzioPingStats.addresses)
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
	err := os.Setenv(addressesEnvName, "www.google.com,https://listener.logz.io:8053")
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
	err := os.Setenv(addressesEnvName, "www.google.com,https://listener.logz.io:8053")
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
	err := os.Setenv(addressesEnvName, "www.google.com,https://listener.logz.io:8053")
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
	err := os.Setenv(addressesEnvName, "www.google.com,https://listener.logz.io:8053")
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
	err := os.Setenv(addressesEnvName, "www.google.com,https://listener.logz.io:8053")
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
	err := os.Setenv(addressesEnvName, "www.google.com,https://listener.logz.io:8053")
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
	err := os.Setenv(addressesEnvName, "www.google.com,https://listener.logz.io:8053")
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
	err := os.Setenv(addressesEnvName, "www.google.com,https://listener.logz.io:8053")
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
	err := os.Setenv(addressesEnvName, "www.google.com,https://listener.logz.io:8053")
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
	err := os.Setenv(addressesEnvName, "www.google.com,https://listener.logz.io:8053")
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
	err := os.Setenv(addressesEnvName, "www.google.com,https://listener.logz.io:8053")
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
		assert.Contains(t, []string{"www.google.com:80", "listener.logz.io:8053"}, pingStats.address)
		assert.NotEmpty(t, pingStats.addressIP)
	}
}

func TestCreateController(t *testing.T) {
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
			body, err := ioutil.ReadAll(request.Body)
			require.NoError(t, err)

			uncompressedBody, err := snappy.Decode(nil, body)
			require.NoError(t, err)

			writeRequest := &prompb.WriteRequest{}
			err = writeRequest.Unmarshal(uncompressedBody)
			require.NoError(t, err)

			assert.Len(t, writeRequest.Timeseries, 12)

			metricsToSend := make([]map[string]interface{}, 0)

			for _, timeseries := range writeRequest.Timeseries {
				metricToSend := make(map[string]interface{})

				metricToSend["value"] = timeseries.Samples[0].Value

				for _, label := range timeseries.Labels {
					metricToSend[label.Name] = label.Value
				}

				metricsToSend = append(metricsToSend, metricToSend)
			}

			assert.Len(t, metricsToSend, 12)

			for _, metricToSend := range metricsToSend {
				assert.Contains(t, []string{rttMetricName, probesSentMetricName, successfulProbesMetricName, probesFailedMetricName}, metricToSend["__name__"])

				if metricToSend["__name__"] == rttMetricName {
					assert.Len(t, metricToSend, 9)

					assert.NotEmpty(t, metricToSend["value"])
					assert.Contains(t, []string{"www.google.com:80", "listener.logz.io:8053"}, metricToSend["address"])
					assert.NotEmpty(t, metricToSend["ip"])
					assert.Contains(t, []string{"1", "2", "3"}, metricToSend["rtt_index"])
					assert.Equal(t, "3", metricToSend["total_rtts"])
					assert.Equal(t, "milliseconds", metricToSend["unit"])
				} else if metricToSend["__name__"] == probesSentMetricName {
					assert.Len(t, metricToSend, 6)

					assert.Equal(t, float64(3), metricToSend["value"])
					assert.Contains(t, []string{"www.google.com:80", "listener.logz.io:8053"}, metricToSend["address"])
					assert.NotEmpty(t, metricToSend["ip"])
				} else if metricToSend["__name__"] == successfulProbesMetricName {
					assert.Len(t, metricToSend, 6)

					assert.Equal(t, float64(3), metricToSend["value"])
					assert.Contains(t, []string{"www.google.com:80", "listener.logz.io:8053"}, metricToSend["address"])
					assert.NotEmpty(t, metricToSend["ip"])
				} else if metricToSend["__name__"] == probesFailedMetricName {
					assert.Len(t, metricToSend, 6)

					assert.Equal(t, float64(0), metricToSend["value"])
					assert.Contains(t, []string{"www.google.com:80", "listener.logz.io:8053"}, metricToSend["address"])
					assert.NotEmpty(t, metricToSend["ip"])
				}

				assert.Equal(t, "us-east-1", metricToSend["aws_region"])
				assert.Equal(t, "test", metricToSend["aws_lambda_function"])
			}

			return httpmock.NewStringResponse(200, ""), nil
		})

	err = logzioPingStats.collectMetrics()
	require.NoError(t, err)
}
