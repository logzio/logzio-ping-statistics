package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	metricsExporter "github.com/logzio/go-metrics-sdk"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

const (
	addressesEnvName             = "ADDRESSES"
	pingCountEnvName             = "PING_COUNT"
	pingIntervalEnvName          = "PING_INTERVAL"
	pingTimeoutEnvName           = "PING_TIMEOUT"
	logzioMetricsListenerEnvName = "LOGZIO_METRICS_LISTENER"
	logzioMetricsTokenEnvName    = "LOGZIO_METRICS_TOKEN"
	awsRegionEnvName             = "AWS_REGION"
	awsLambdaFunctionNameEnvName = "AWS_LAMBDA_FUNCTION_NAME"
	addressHttpsPrefix           = "https://"
	addressHttpPrefix            = "http://"
	addressTcpPrefix			 = "tcp://"
	addressSuffixDefaultPort     = ":80"
	meterName                    = "ping_stats"
	rttMetricName                = meterName + "_rtt"
	probesSentMetricName         = meterName + "_probes_sent"
	successfulProbesMetricName   = meterName + "_successful_probes"
	probesFailedMetricName       = meterName + "_probes_failed"
)

var (
	debugLogger = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	infoLogger  = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
)

type logzioPingStatistics struct {
	ctx                   context.Context
	logzioMetricsListener string
	logzioMetricsToken    string
	addresses             []string
	pingCount             int
	pingInterval          time.Duration
	pingTimeout           time.Duration
	pingsStats            []*pingStatistics
}

type pingStatistics struct {
	probesSent       int
	successfulProbes int
	probesFailed     int
	addressIP        string
	address          string
	rtts             []float64
}

func newLogzioPingStatistics(ctx context.Context) (*logzioPingStatistics, error) {
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

	addresses := getAddresses(addressesString)

	pingCount, err := getNumberEnvValue(os.Getenv(pingCountEnvName), pingCountEnvName)
	if err != nil {
		return nil, err
	}

	pingInterval, err := getNumberEnvValue(os.Getenv(pingIntervalEnvName), pingIntervalEnvName)
	if err != nil {
		return nil, err
	}

	pingTimeout, err := getNumberEnvValue(os.Getenv(pingTimeoutEnvName), pingTimeoutEnvName)
	if err != nil {
		return nil, err
	}

	return &logzioPingStatistics{
		ctx:                   ctx,
		logzioMetricsListener: logzioMetricsListener,
		logzioMetricsToken:    logzioMetricsToken,
		addresses:             addresses,
		pingCount:             *pingCount,
		pingInterval:          time.Duration(*pingInterval) * time.Second,
		pingTimeout:           time.Duration(*pingTimeout) * time.Second,
		pingsStats:            make([]*pingStatistics, 0),
	}, nil
}

func (lps *logzioPingStatistics) getAddressPingStatistics(address string) (*pingStatistics, error) {
	debugLogger.Println("Getting ping statistics for address:", address)

	rtts := make([]float64, 0)
	successfulProbes := 0
	var addressIP string

	for count := 0; count < lps.pingCount; count++ {
		time.Sleep(lps.pingInterval)

		start := time.Now()
		conn, err := net.DialTimeout("tcp", address, lps.pingTimeout)
		if err != nil {
			errorLogger.Println("Error connecting to address:", address, ":", err)
			continue
		}

		end := time.Now()

		if count+1 == lps.pingCount {
			addressIP = conn.RemoteAddr().String()
		}

		if err = conn.Close(); err != nil {
			return nil, fmt.Errorf("error closing connection: %v", err)
		}

		rtt := float64(end.Sub(start)) / float64(time.Millisecond)
		successfulProbes++

		rtts = append(rtts, rtt)
	}

	if len(rtts) == 0 {
		errorLogger.Println("Did not get ping statistics rtts for address:", address)
	}

	return &pingStatistics{
		probesSent:       lps.pingCount,
		successfulProbes: successfulProbes,
		probesFailed:     lps.pingCount - successfulProbes,
		addressIP:        addressIP,
		address:          address,
		rtts:             rtts,
	}, nil
}

func (lps *logzioPingStatistics) getAllAddressesPingStatistics() error {
	debugLogger.Println("Getting ping statistics for all addresses...")

	pingStatsChan := make(chan *pingStatistics)
	defer close(pingStatsChan)

	for _, address := range lps.addresses {
		go func(address string) {
			pingStats, err := lps.getAddressPingStatistics(address)
			if err != nil {
				errorLogger.Println("Error getting ping statistics for address", address, ":", err)
				pingStatsChan <- nil
				return
			}

			pingStatsChan <- pingStats
		}(address)
	}

	for range lps.addresses {
		if pingStats := <-pingStatsChan; pingStats != nil {
			lps.pingsStats = append(lps.pingsStats, pingStats)
		}
	}

	if len(lps.pingsStats) == 0 {
		return fmt.Errorf("did not get ping statistics for any given address")
	}

	return nil
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
			for index, rtt := range pingStats.rtts {
				result.Observe(rtt,
					attribute.String("address", pingStats.address),
					attribute.String("ip", pingStats.addressIP),
					attribute.Int("rtt_index", index+1),
					attribute.Int("total_rtts", len(pingStats.rtts)),
					attribute.String("unit", "milliseconds"),
				)
			}
		}
	}
}

func (lps *logzioPingStatistics) getProbesSentObserverCallback() func(context.Context, metric.Int64ObserverResult) {
	return func(_ context.Context, result metric.Int64ObserverResult) {
		debugLogger.Println("Running probes sent observer callback...")

		for _, pingStats := range lps.pingsStats {
			result.Observe(int64(pingStats.probesSent),
				attribute.String("address", pingStats.address),
				attribute.String("ip", pingStats.addressIP))
		}
	}
}

func (lps *logzioPingStatistics) getSuccessfulProbesObserverCallback() func(context.Context, metric.Int64ObserverResult) {
	return func(_ context.Context, result metric.Int64ObserverResult) {
		debugLogger.Println("Running successful probes observer callback...")

		for _, pingStats := range lps.pingsStats {
			result.Observe(int64(pingStats.successfulProbes),
				attribute.String("address", pingStats.address),
				attribute.String("ip", pingStats.addressIP))
		}
	}
}

func (lps *logzioPingStatistics) getProbesFailedObserverCallback() func(context.Context, metric.Int64ObserverResult) {
	return func(_ context.Context, result metric.Int64ObserverResult) {
		debugLogger.Println("Running probes failed observer callback...")

		for _, pingStats := range lps.pingsStats {
			result.Observe(int64(pingStats.probesFailed),
				attribute.String("address", pingStats.address),
				attribute.String("ip", pingStats.addressIP))
		}
	}
}

func (lps *logzioPingStatistics) collectMetrics() error {
	cont, err := lps.createController()
	if err != nil {
		panic(fmt.Errorf("error creating controller: %v", err))
	}

	debugLogger.Println("Collecting metrics...")

	defer func() {
		handleErr(cont.Stop(lps.ctx))
	}()

	meter := cont.Meter(meterName)

	_ = metric.Must(meter).NewFloat64GaugeObserver(
		rttMetricName,
		lps.getRttObserverCallback(),
		metric.WithDescription("Ping RTT"),
	)

	_ = metric.Must(meter).NewInt64GaugeObserver(
		probesSentMetricName,
		lps.getProbesSentObserverCallback(),
		metric.WithDescription("Ping probes sent"),
	)

	_ = metric.Must(meter).NewInt64GaugeObserver(
		successfulProbesMetricName,
		lps.getSuccessfulProbesObserverCallback(),
		metric.WithDescription("Ping successful probes"),
	)

	_ = metric.Must(meter).NewInt64GaugeObserver(
		probesFailedMetricName,
		lps.getProbesFailedObserverCallback(),
		metric.WithDescription("Ping probes failed"),
	)

	return nil
}

func run(ctx context.Context) error {
	logzioPingStats, err := newLogzioPingStatistics(ctx)
	if err != nil {
		return fmt.Errorf("error creating logzioPingStatistics instance: %v", err)
	}

	if err = logzioPingStats.getAllAddressesPingStatistics(); err != nil {
		return fmt.Errorf("error getting all addresses ping statistics: %v", err)
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

func getAddresses(addressesString string) []string {
	addresses := strings.Split(addressesString, ",")
	re := regexp.MustCompile(":[0-9]+$")

	for index := 0; index < len(addresses); index++ {
		addresses[index] = strings.Replace(addresses[index], " ", "", -1)

		if strings.Contains(addresses[index], addressHttpsPrefix) {
			addresses[index] = strings.Replace(addresses[index], addressHttpsPrefix, "", 1)
		} else if strings.Contains(addresses[index], addressHttpPrefix) {
			addresses[index] = strings.Replace(addresses[index], addressHttpPrefix, "", 1)
		} else if strings.Contains(addresses[index], addressTcpPrefix) {
			addresses[index] = strings.Replace(addresses[index], addressTcpPrefix, "", 1)
		}

		if re.FindStringSubmatch(addresses[index]) == nil {
			addresses[index] = addresses[index] + addressSuffixDefaultPort
		}
	}

	return addresses
}

func getNumberEnvValue(envValue string, envName string) (*int, error) {
	numberEnvValue, err := strconv.Atoi(envValue)
	if err != nil {
		return nil, fmt.Errorf("%s must be a number", envName)
	}

	if numberEnvValue < 1 {
		return nil, fmt.Errorf("%s must be a positive number", envName)
	}

	return &numberEnvValue, nil
}

func HandleRequest(ctx context.Context) error {
	infoLogger.Println("Starting to get ping statistics for all addresses...")

	if err := run(ctx); err != nil {
		return err
	}

	infoLogger.Println("The ping statistics have been sent to Logz.io successfully")
	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
