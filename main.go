package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-ping/ping"
	metricsExporter "github.com/logzio/go-metrics-sdk"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

const (
	addressesEnvName                = "ADDRESSES"
	pingCountEnvName                = "PING_COUNT"
	pingIntervalEnvName             = "PING_INTERVAL"
	logzioMetricsListenerEnvName    = "LOGZIO_METRICS_LISTENER"
	logzioMetricsTokenEnvName       = "LOGZIO_METRICS_TOKEN"
	awsRegionEnvName                = "AWS_REGION"
	awsLambdaFunctionNameEnvName    = "AWS_LAMBDA_FUNCTION_NAME"
	maxPingCount                    = 25
	meterName                       = "ping_stats"
	rttMetricName                   = meterName + "_rtt"
	stdDevRttMetricName             = meterName + "_std_dev_rtt"
	packetsSentMetricName           = meterName + "_packets_sent"
	packetsLossMetricName           = meterName + "_packets_loss"
	packetsRecvMetricName           = meterName + "_packets_recv"
	packetsRecvDuplicatesMetricName = meterName + "_packets_recv_duplicates"
)

var (
	debugLogger = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	infoLogger  = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
)

type logzioPingStatistics struct {
	logzioMetricsListener string
	logzioMetricsToken    string
	addresses             []string
	pingCount             int
	pingInterval          time.Duration
	pingsStats            []*ping.Statistics
}

func newLogzioPingStatistics() (*logzioPingStatistics, error) {
	logzioMetricsListener := os.Getenv(logzioMetricsListenerEnvName)
	if logzioMetricsListener == "" {
		return nil, fmt.Errorf("%s must not be empty", logzioMetricsListenerEnvName)
	}

	logzioMetricsToken := os.Getenv(logzioMetricsTokenEnvName)
	if logzioMetricsToken == "" {
		return nil, fmt.Errorf("%s must not be empty", logzioMetricsTokenEnvName)
	}

	addressesString := os.Getenv(addressesEnvName)
	if addressesString == "" {
		return nil, fmt.Errorf("%s must not be empty", addressesEnvName)
	}

	pingCount, err := strconv.Atoi(os.Getenv(pingCountEnvName))
	if err != nil {
		return nil, fmt.Errorf("%s must be a number", pingCountEnvName)
	}

	if pingCount < 1 || pingCount > maxPingCount {
		return nil, fmt.Errorf("%s must be between 1 and %d (include)", pingCountEnvName, maxPingCount)
	}

	pingInterval, err := strconv.Atoi(os.Getenv(pingIntervalEnvName))
	if err != nil {
		return nil, fmt.Errorf("%s must be a number", pingIntervalEnvName)
	}

	if pingInterval < 1 {
		return nil, fmt.Errorf("%s must be greater or equal to 1", pingIntervalEnvName)
	}

	debugLogger.Println("Creating logzioPingStatistics instance...")

	return &logzioPingStatistics{
		logzioMetricsListener: logzioMetricsListener,
		logzioMetricsToken:    logzioMetricsToken,
		addresses:             strings.Split(addressesString, ","),
		pingCount:             pingCount,
		pingInterval:          time.Duration(pingInterval) * time.Second,
		pingsStats:            make([]*ping.Statistics, 0),
	}, nil
}

func (lps *logzioPingStatistics) getPingStatistics(pinger *ping.Pinger) (*ping.Statistics, error) {
	debugLogger.Println("Getting ping statistics for address:", pinger.Addr())

	if err := pinger.Run(); err != nil {
		return nil, fmt.Errorf("error running pinger: %v", err)
	}

	return pinger.Statistics(), nil
}

func (lps *logzioPingStatistics) getPingsStatistics() error {
	debugLogger.Println("Getting ping statistics for all addresses...")

	pingStatsChan := make(chan *ping.Statistics)
	defer close(pingStatsChan)

	for _, address := range lps.addresses {
		go func(address string, pingStatsChan chan *ping.Statistics) {
			pinger, err := lps.createPinger(address)
			if err != nil {
				errorLogger.Println(err)
				pingStatsChan <- nil
				return
			}

			pingStats, err := lps.getPingStatistics(pinger)
			if err != nil {
				errorLogger.Println("Error getting ping statistics for address", address, ":", err)
				pingStatsChan <- nil
				return
			}

			pingStatsChan <- pingStats
		}(address, pingStatsChan)
	}

	for range lps.addresses {
		if pingStats := <-pingStatsChan; pingStats != nil {
			lps.pingsStats = append(lps.pingsStats, pingStats)
		}
	}

	pingStatsSize := len(lps.pingsStats)

	if pingStatsSize == 0 {
		return fmt.Errorf("did not get ping statistics for all addresses")
	}

	debugLogger.Println("Got", pingStatsSize, "ping/s statistics")

	return nil
}

func (lps *logzioPingStatistics) createPinger(address string) (*ping.Pinger, error) {
	debugLogger.Println("Creating Pinger for address: ", address)

	pinger, err := ping.NewPinger(address)
	if err != nil {
		return nil, fmt.Errorf("error creating pinger: %v", err)
	}

	pinger.Count = lps.pingCount
	pinger.Interval = lps.pingInterval
	pinger.SetPrivileged(true)

	return pinger, nil
}

func (lps *logzioPingStatistics) createController() (*controller.Controller, error) {
	debugLogger.Println("Creating controller...")

	config := metricsExporter.Config{
		LogzioMetricsListener: lps.logzioMetricsListener,
		LogzioMetricsToken:    lps.logzioMetricsToken,
		RemoteTimeout:         30 * time.Second,
		PushInterval:          15 * time.Second,
	}

	return metricsExporter.InstallNewPipeline(config,
		controller.WithCollectPeriod(5*time.Second),
		controller.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				attribute.String("aws_region", os.Getenv(awsRegionEnvName)),
				attribute.String("aws_lambda_function", os.Getenv(awsLambdaFunctionNameEnvName)),
			),
		),
	)
}

func (lps *logzioPingStatistics) getRttObserverCallback() func(context.Context, metric.Float64ObserverResult) {
	return func(_ context.Context, result metric.Float64ObserverResult) {
		debugLogger.Println("Running RTT observer callback...")

		for _, pingStats := range lps.pingsStats {
			for index, rtt := range pingStats.Rtts {
				result.Observe(float64(rtt)/float64(time.Millisecond),
					attribute.String("address", pingStats.Addr),
					attribute.String("ip", pingStats.IPAddr.String()),
					attribute.Int("rtt_index", index+1),
					attribute.Int("total_rtts", lps.pingCount),
					attribute.String("unit", "milliseconds"),
				)
			}
		}
	}
}

func (lps *logzioPingStatistics) getStdDevRttObserverCallback() func(context.Context, metric.Float64ObserverResult) {
	return func(_ context.Context, result metric.Float64ObserverResult) {
		for _, pingStats := range lps.pingsStats {
			debugLogger.Println("Running standard deviation RTT observer callback...")

			result.Observe(float64(pingStats.StdDevRtt)/float64(time.Millisecond),
				attribute.String("address", pingStats.Addr),
				attribute.String("ip", pingStats.IPAddr.String()),
				attribute.String("unit", "milliseconds"),
			)
		}
	}
}

func (lps *logzioPingStatistics) getPacketsSentObserverCallback() func(context.Context, metric.Int64ObserverResult) {
	return func(_ context.Context, result metric.Int64ObserverResult) {
		debugLogger.Println("Running packets sent observer callback...")

		for _, pingStats := range lps.pingsStats {
			result.Observe(int64(pingStats.PacketsSent),
				attribute.String("address", pingStats.Addr),
				attribute.String("ip", pingStats.IPAddr.String()))
		}
	}
}

func (lps *logzioPingStatistics) getPacketLossObserverCallback() func(context.Context, metric.Float64ObserverResult) {
	return func(_ context.Context, result metric.Float64ObserverResult) {
		debugLogger.Println("Running packet loss observer callback...")

		for _, pingStats := range lps.pingsStats {
			result.Observe(pingStats.PacketLoss,
				attribute.String("address", pingStats.Addr),
				attribute.String("ip", pingStats.IPAddr.String()))
		}
	}
}

func (lps *logzioPingStatistics) getPacketsRecvObserverCallback() func(context.Context, metric.Int64ObserverResult) {
	return func(_ context.Context, result metric.Int64ObserverResult) {
		debugLogger.Println("Running packets received observer callback...")

		for _, pingStats := range lps.pingsStats {
			result.Observe(int64(pingStats.PacketsRecv),
				attribute.String("address", pingStats.Addr),
				attribute.String("ip", pingStats.IPAddr.String()))
		}
	}
}

func (lps *logzioPingStatistics) getPacketsRecvDuplicatesObserverCallback() func(context.Context, metric.Int64ObserverResult) {
	return func(_ context.Context, result metric.Int64ObserverResult) {
		debugLogger.Println("Running packets received duplicates observer callback...")

		for _, pingStats := range lps.pingsStats {
			result.Observe(int64(pingStats.PacketsRecvDuplicates),
				attribute.String("address", pingStats.Addr),
				attribute.String("ip", pingStats.IPAddr.String()))
		}
	}
}

func (lps *logzioPingStatistics) collectMetrics() error {
	ctx := context.Background()
	cont, err := lps.createController()
	if err != nil {
		panic(fmt.Errorf("error creating controller: %v", err))
	}

	debugLogger.Println("Collecting metrics...")

	defer func() {
		handleErr(cont.Stop(ctx))
	}()

	meter := cont.Meter(meterName)

	_ = metric.Must(meter).NewFloat64GaugeObserver(
		rttMetricName,
		lps.getRttObserverCallback(),
		metric.WithDescription("Ping RTT"),
	)

	_ = metric.Must(meter).NewFloat64GaugeObserver(
		stdDevRttMetricName,
		lps.getStdDevRttObserverCallback(),
		metric.WithDescription("Ping standard deviation RTT"),
	)

	_ = metric.Must(meter).NewInt64GaugeObserver(
		packetsSentMetricName,
		lps.getPacketsSentObserverCallback(),
		metric.WithDescription("Ping packets sent"),
	)

	_ = metric.Must(meter).NewFloat64GaugeObserver(
		packetsLossMetricName,
		lps.getPacketLossObserverCallback(),
		metric.WithDescription("Ping packet loss"),
	)

	_ = metric.Must(meter).NewInt64GaugeObserver(
		packetsRecvMetricName,
		lps.getPacketsRecvObserverCallback(),
		metric.WithDescription("Ping packets received"),
	)

	_ = metric.Must(meter).NewInt64GaugeObserver(
		packetsRecvDuplicatesMetricName,
		lps.getPacketsRecvDuplicatesObserverCallback(),
		metric.WithDescription("Ping packets received duplicates"),
	)

	return nil
}

func run() error {
	logzioPingStats, err := newLogzioPingStatistics()
	if err != nil {
		return fmt.Errorf("error creating logzioPingStatistics instance: %v", err)
	}

	if err = logzioPingStats.getPingsStatistics(); err != nil {
		return fmt.Errorf("error creating pings statistics: %v", err)
	}

	if err = logzioPingStats.collectMetrics(); err != nil {
		return fmt.Errorf("error collecting metrics: %v", err)
	}

	return nil
}

func handleErr(err error) {
	if err != nil {
		panic(fmt.Errorf("something went wrong: %v", err))
	}
}

func main() {
	infoLogger.Println("Starting to get ping statistics for all addresses...")

	if err := run(); err != nil {
		panic(err)
	}

	infoLogger.Println("The ping statistics have been sent to Logz.io successfully")
}
