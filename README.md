Prometheus ClustSafe Exporter
========================

This is an ClustSafe exporter for [Prometheus](https://prometheus.io).

Read the configuration section to setup dynamic targets, the exporter doesn't have any default metrics of the host it is running on.

## Exposed metrics
| Metric                            | Meaning                                          | Labels         |
|-----------------------------------|--------------------------------------------------|----------------|
| clustsafe_humidity                | The humidity in percentage.                      | sensor         |
| clustsafe_temperature             | The temperature in Celsius.                      | sensor         |
| clustsafe_outlet_up               | The status of an outlet in the module.           | module, outlet |
| clustsafe_line_up                 | The status of in input power line in the module. | module, line   |
| clustsafe_power_consumption_watts | The real power consumption in Watts.             | module         |
| clustsafe_up                      | Was the last scrape of ClustSafe successful.     |                |

## Installation

For most use-cases, simply download [the latest release](https://github.com/astappiev/clustsafe_exporter/releases).

### Building from source

You need a Go development environment. Then, simply run `make` to build the executables:

    make build

## Running

A minimal invocation looks like this:

    CLUSTSAFE_USER=username CLUSTSAFE_PASSWORD=password ./clustsafe_exporter

You have to provide username and password as environment variables, it's not safe to pass them as request parameters.

Supported parameters include:

 - `web.listen-address`: the address/port to listen on (default: `":9879"`)
 - `web.telemetry-path`: the path under which to expose metrics (default: `"/metrics"`)
 - `clustsafe.path`: the ClustSafe command to use (default: `"cw-clustsafe"`, can be `"clustsafeX"` for older version)
 - `clustsafe.command`: the command to execute, can be `clustsafes`, `sensors` or `all` (default: `"all"`)

### Configuration

To add your ClustSafe targets to Prometheus, you can use any of the supported service discovery mechanism of your choice.
The following example uses the file-based SD and should be easy to adjust to other scenarios.

Create a YAML file that contains a list of targets, e.g.:

```
---
- targets:
  - 10.0.0.1
  - 10.0.0.2
  - 10.0.0.3
  - 10.0.0.4
  labels:
    job: clustsafe_exporter
```

This file needs to be stored on the Prometheus server host.  Assuming that this file is called
`/etc/prometheus/clustsafe_exporter/targets.yml`, and the ClustSafe exporter is running on a host that has
the DNS name `internal.example.com`, add the following to your Prometheus config:

```
- job_name: clustsafe
  scrape_interval: 1m
  scrape_timeout: 30s
  file_sd_configs:
  - files:
    - /etc/prometheus/clustsafe_exporter/targets.yml
    refresh_interval: 5m
  relabel_configs:
  - source_labels: [__address__]
    separator: ;
    regex: (.*)
    target_label: __param_target
    replacement: ${1}
    action: replace
  - source_labels: [__param_target]
    separator: ;
    regex: (.*)
    target_label: instance
    replacement: ${1}
    action: replace
  - separator: ;
    regex: .*
    target_label: __address__
    replacement: internal.example.com:9879
    action: replace
```

For more information, e.g. how to use mechanisms other than a file to discover the list of hosts to scrape,
please refer to the [Prometheus documentation](https://prometheus.io/docs).
