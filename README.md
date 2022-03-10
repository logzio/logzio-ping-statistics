# Ping Statistics Auto-Deployment

Auto-deployment of Lambda function that collects ping statistics metrics of addresses and sends them to Logz.io.

* The Lambda function will be deployed with the layer LogzioLambdaExtensionLogs.
  For more information about the extension [click here](https://github.com/logzio/logzio-lambda-extensions/tree/main/logzio-lambda-extensions-logs).

## Getting Started

To start just press the button and follow the instructions:

[![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=us-east-1#/stacks/create/template?templateURL=https://logzio-aws-integrations-us-east-1.s3.amazonaws.com/ping-statistics-auto-deployment/auto-deployment.yaml&stackName=logzio-ping-statistics-auto-deployment)

### Parameters

| Parameter | Description | Required/Optional | Default |
| --- | --- | --- | --- |
| Address | The addresses to ping. You can add port for each address (default port for address is 80). Addresses must be separated by comma. | Required | - |
| PingCount | The number of pings for each address. | Required | `3` |
| PingInterval | The time to wait (seconds) between each ping. | Required | `1 (second)` |
| PingTimeout | The timeout (seconds) for each ping. | Required | `10 (seconds)` |
| LogzioMetricsListener | The Logz.io metrics listener URL for your region. (For more details, see the regions page: https://docs.logz.io/user-guide/accounts/account-region.html) | Required | `https://listener.logz.io:8053` |
| LogzioMetricsToken | Your Logz.io metrics token (Can be retrieved from the Manage Token page). | Required | - |
| CloudWatchEventScheduleExpression | The scheduling expression that determines when and how often the Lambda function runs. | Required | `rate(30 minutes)` |
| LogzioLogsListener | The Logz.io logs listener URL for your region. (For more details, see the regions page: https://docs.logz.io/user-guide/accounts/account-region.html) | Required | `https://listener.logz.io:8071` |
| LogzioLogsToken | Your Logz.io logs token (Can be retrieved from the Manage Token page). | Required | - |
| LogsExtensionLogLevel | Log level of the extension. Can be set to one of the following: debug, info, warn, error, fatal, panic | Required | `info` |
| EnableExtensionInnerLogs | Set to true if you wish the extension inner logs will be shipped to your Logz.io account. | Required | `false` |
| EnablePlatformLogs | The platform log captures runtime or execution environment errors. Set to true if you wish the platform logs will be shipped to your Logz.io account. | Required | `false` |
| GrokPatters | Must be set with LogsFormat. Use this if you want to parse your logs into fields. A minified JSON list that contains the field name and the regex that will match the field. | Optional | - |
| LogsFormat | Must be set with GrokPatters. Use this if you want to parse your logs into fields. The format in which the logs will appear, in accordance to grok conventions. | Optional | - |
| LogsCustomFields | Include additional fields with every message sent, formatted as fieldName1=fieldValue1;fieldName2=fieldValue2 (NO SPACES). A custom key that clashes with a key from the log itself will be ignored. | Optional | - |

## Searching in Logz.io

All metrics that were sent from the lambda function will have the prefix `ping_stats` in their name. 