package main

import (
	"io/ioutil"
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

func TestNewLogzioPingStatistics_Success(t *testing.T) {
	err := os.Setenv(addressesEnvName, "www.google.com,www.nynews.com")
	require.NoError(t, err)

	err = os.Setenv(pingCountEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(pingIntervalEnvName, "1")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	logzioPingStats, err := newLogzioPingStatistics()
	require.NoError(t, err)
	require.NotNil(t, logzioPingStats)

	assert.Equal(t, []string{"www.google.com", "www.nynews.com"}, logzioPingStats.addresses)
	assert.Equal(t, 10, logzioPingStats.pingCount)
	assert.Equal(t, 1*time.Second, logzioPingStats.pingInterval)
	assert.Equal(t, "https://listener.logz.io:8053", logzioPingStats.logzioMetricsListener)
	assert.Equal(t, "123456789a", logzioPingStats.logzioMetricsToken)

	os.Clearenv()
}

func TestNewLogzioPingStatistics_NoAddress(t *testing.T) {
	err := os.Setenv(pingCountEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(pingIntervalEnvName, "1")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	_, err = newLogzioPingStatistics()
	require.Error(t, err)

	os.Clearenv()
}

func TestNewLogzioPingStatistics_NoPingCount(t *testing.T) {
	err := os.Setenv(addressesEnvName, "www.google.com,www.nynews.com")
	require.NoError(t, err)

	err = os.Setenv(pingIntervalEnvName, "1")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	_, err = newLogzioPingStatistics()
	require.Error(t, err)

	os.Clearenv()
}

func TestNewLogzioPingStatistics_NoPingInterval(t *testing.T) {
	err := os.Setenv(addressesEnvName, "www.google.com,www.nynews.com")
	require.NoError(t, err)

	err = os.Setenv(pingCountEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	_, err = newLogzioPingStatistics()
	require.Error(t, err)

	os.Clearenv()
}

func TestNewLogzioPingStatistics_NoLogzioMetricsListener(t *testing.T) {
	err := os.Setenv(addressesEnvName, "www.google.com,www.nynews.com")
	require.NoError(t, err)

	err = os.Setenv(pingCountEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(pingIntervalEnvName, "1")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	_, err = newLogzioPingStatistics()
	require.Error(t, err)

	os.Clearenv()
}

func TestNewLogzioPingStatistics_NoLogzioMetricsToken(t *testing.T) {
	err := os.Setenv(addressesEnvName, "www.google.com,www.nynews.com")
	require.NoError(t, err)

	err = os.Setenv(pingCountEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(pingIntervalEnvName, "1")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	_, err = newLogzioPingStatistics()
	require.Error(t, err)

	os.Clearenv()
}

func TestGetPingsStatistics_Success(t *testing.T) {
	logzioPingStats := &logzioPingStatistics{
		logzioMetricsListener: "https://listener.logz.io:8053",
		logzioMetricsToken:    "123456789a",
		addresses:             []string{"www.google.com", "www.nynews.com"},
		pingCount:             10,
		pingInterval:          1 * time.Second,
	}

	err := logzioPingStats.getPingsStatistics()
	require.NoError(t, err)

	assert.Len(t, logzioPingStats.pingsStats, 2)

	for _, pingStats := range logzioPingStats.pingsStats {
		assert.NotNil(t, pingStats)
		assert.Len(t, pingStats.Rtts, 10)
	}
}

func TestCollectMetrics_Success(t *testing.T) {
	logzioPingStats := &logzioPingStatistics{
		logzioMetricsListener: "https://listener.logz.io:8053",
		logzioMetricsToken:    "123456789a",
		addresses:             []string{"www.google.com", "www.nynews.com"},
		pingCount:             3,
		pingInterval:          1 * time.Second,
	}

	err := logzioPingStats.getPingsStatistics()
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

			assert.Len(t, writeRequest.Timeseries, 16)

			for _, timeseries := range writeRequest.Timeseries {
				for _, label := range timeseries.Labels {
					if label.Name == "__name__" {
						if label.Value == rttMetricName {
							assert.Len(t, timeseries.Labels, 8)

							for _, rttMetricLabel := range timeseries.Labels {
								assert.Contains(t, []string{"__name__", "aws_region", "aws_lambda_function", "address", "ip", "rtt_index", "total_rtts", "unit"}, rttMetricLabel.Name)
							}
						} else if label.Value == stdDevRttMetricName {
							assert.Len(t, timeseries.Labels, 6)

							for _, rttMetricLabel := range timeseries.Labels {
								assert.Contains(t, []string{"__name__", "aws_region", "aws_lambda_function", "address", "ip", "unit"}, rttMetricLabel.Name)
							}
						} else if label.Value == packetsSentMetricName {
							assert.Len(t, timeseries.Labels, 5)

							for _, rttMetricLabel := range timeseries.Labels {
								assert.Contains(t, []string{"__name__", "aws_region", "aws_lambda_function", "address", "ip"}, rttMetricLabel.Name)
							}
						} else if label.Value == packetsLossMetricName {
							assert.Len(t, timeseries.Labels, 5)

							for _, rttMetricLabel := range timeseries.Labels {
								assert.Contains(t, []string{"__name__", "aws_region", "aws_lambda_function", "address", "ip"}, rttMetricLabel.Name)
							}
						} else if label.Value == packetsRecvMetricName {
							assert.Len(t, timeseries.Labels, 5)

							for _, rttMetricLabel := range timeseries.Labels {
								assert.Contains(t, []string{"__name__", "aws_region", "aws_lambda_function", "address", "ip"}, rttMetricLabel.Name)
							}
						} else if label.Value == packetsRecvDuplicatesMetricName {
							assert.Len(t, timeseries.Labels, 5)

							for _, rttMetricLabel := range timeseries.Labels {
								assert.Contains(t, []string{"__name__", "aws_region", "aws_lambda_function", "address", "ip"}, rttMetricLabel.Name)
							}
						} else {
							assert.Failf(t, "", "")
						}
					}
				}
			}

			return httpmock.NewStringResponse(200, ""), nil
		})

	err = logzioPingStats.collectMetrics()
	require.NoError(t, err)
}
