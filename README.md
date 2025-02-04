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
| Addresses | The addresses to ping. You can add port for each address (default port for address is 80). Addresses must be separated by comma. (Example addresses: `www.google.com`, `tcp://www.google.com`, `https://www.google.com`, `http://www.google.com`). | Required | - |
| PingCount | The number of pings for each address. | Required | `3` |
| PingInterval | The time to wait (seconds) between each ping. | Required | `1 (second)` |
| PingTimeout | The timeout (seconds) for each ping. | Required | `10 (seconds)` |
| LogzioListener | The Logz.io listener URL for your region. (For more details, see the regions page: https://docs.logz.io/user-guide/accounts/account-region.html) | Required | `https://listener.logz.io` |
| LogzioMetricsToken | Your Logz.io metrics token (Can be retrieved from the Manage Token page). | Required | - |
| LogzioLogsToken | Your Logz.io logs token (Can be retrieved from the Manage Token page). | Required | - |
| SchedulingInterval | The scheduling expression that determines when and how often the Lambda function runs. Rate below 6 minutes will cause the lambda to behave unexpectedly due to cold start and custom resource invocation. | Required | `rate(30 minutes)` |

## Searching in Logz.io

All metrics that were sent from the Lambda function will have the prefix `ping_stats` in their name. 

## Changelog
**v1.0.3**:
 - Fix custom resource run logic
 - Update `LogzioLambdaExtensionLogs` version 16 -> 18

**v1.0.2**:
 - Update logzio logs lambda extention version

**v1.0.1**:
 - Remove IP label from the metrics.
 - After deployment, the Lambda function will be invoked immediately instead of waiting for scheduler to invoke it, and will be invoked each interval (according the scheduling expression).

**v1.0.0**
