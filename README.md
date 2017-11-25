# BOSH TSDB Prometheus Exporter [![Build Status](https://travis-ci.org/bosh-prometheus/bosh_tsdb_exporter.png)](https://travis-ci.org/bosh-prometheus/bosh_tsdb_exporter)

A [Prometheus][prometheus] exporter for [BOSH OpenTSDB Health Monitor plugin][bosh-tsdb] metrics. Please refer to the [FAQ][faq] for general questions about this exporter.

## Architecture overview

![](https://cdn.rawgit.com/bosh-prometheus/bosh_tsdb_exporter/master/architecture/architecture.svg)

## Installation

### Binaries

Download the already existing [binaries][binaries] for your platform:

```bash
$ ./bosh_tsdb_exporter <flags>
```

### From source

Using the standard `go install` (you must have [Go][golang] already installed in your local machine):

```bash
$ go install github.com/bosh-prometheus/bosh_tsdb_exporter
$ bosh_tsdb_exporter <flags>
```

### Docker

To run the BOSH TSDB exporter as a Docker container, run:

```bash
docker run -p 9194:9194 -p 13321:13321 boshprometheus/bosh-tsdb-exporter <flags>
```

### BOSH

This exporter can be deployed using the [Prometheus BOSH Release][prometheus-boshrelease].

## Usage

### Flags

| Flag / Environment Variable | Required | Default | Description |
| --------------------------- | -------- | ------- | ----------- |
| `metrics.namespace`<br />`BOSH_TSDB_EXPORTER_METRICS_NAMESPACE` | No | `bosh_tsdb` | Metrics Namespace |
| `metrics.environment`<br />`BOSH_TSDB_EXPORTER_METRICS_ENVIRONMENT` | No | | Environment label to be attached to metrics |
| `tsdb.listen-address`<br />`BOSH_TSDB_EXPORTER_TSDB_LISTEN_ADDRESS` | No | `:13321` | Address to listen on for the TSDB collector |
| `web.listen-address`<br />`BOSH_TSDB_EXPORTER_WEB_LISTEN_ADDRESS` | No | `:9194` | Address to listen on for web interface and telemetry |
| `web.telemetry-path`<br />`BOSH_TSDB_EXPORTER_WEB_TELEMETRY_PATH` | No | `/metrics` | Path under which to expose Prometheus metrics |
| `web.auth.username`<br />`BOSH_TSDB_EXPORTER_WEB_AUTH_USERNAME` | No | | Username for web interface basic auth |
| `web.auth.password`<br />`BOSH_TSDB_EXPORTER_WEB_AUTH_PASSWORD` | No | | Password for web interface basic auth |
| `web.tls.cert_file`<br />`BOSH_TSDB_EXPORTER_WEB_TLS_CERTFILE` | No | | Path to a file that contains the TLS certificate (PEM format). If the certificate is signed by a certificate authority, the file should be the concatenation of the server's certificate, any intermediates, and the CA's certificate |
| `web.tls.key_file`<br />`BOSH_TSDB_EXPORTER_WEB_TLS_KEYFILE` | No | | Path to a file that contains the TLS private key (PEM format) |

### Metrics

The exporter returns the following metrics:

| Metric | Description | Labels |
| ------ | ----------- | ------ |
| *metrics.namespace*_received_tsdb_messages_total | Total number of BOSH HM TSDB received messages | `environment` |
| *metrics.namespace*_invalid_tsdb_messages_total | Total number of BOSH HM TSDB invalid messages | `environment` |
| *metrics.namespace*_discarded_tsdb_messages_total | Total number of BOSH HM TSDB discarded messages | `environment` |
| *metrics.namespace*_last_tsdb_received_message_timestamp | Number of seconds since 1970 since last received message from BOSH HM TSDB | `environment` |
| *metrics.namespace*_last_hm_tsdb_scrape_timestamp | Number of seconds since 1970 since last scrape of BOSH HM TSDB collector | `environment` |
| *metrics.namespace*_last_hm_tsdb_scrape_duration_seconds | Duration of the last scrape of BOSH HM TSDB collector | `environment` |

The exporter returns the following `Job` metrics:

| Metric | Description | Labels |
| ------ | ----------- | ------ |
| *metrics.namespace*_job_healthy | BOSH Job Healthy (1 for healthy, 0 for unhealthy) | `environment`, `bosh_deployment`, `bosh_job_name`, `bosh_job_id`, `bosh_job_index` |
| *metrics.namespace*_job_load_avg01 | BOSH Job Load avg01 | `environment`, `bosh_deployment`, `bosh_job_name`, `bosh_job_id`, `bosh_job_index` |
| *metrics.namespace*_job_cpu_sys | BOSH Job CPU System | `environment`, `bosh_deployment`, `bosh_job_name`, `bosh_job_id`, `bosh_job_index` |
| *metrics.namespace*_job_cpu_user | BOSH Job CPU User | `environment`, `bosh_deployment`, `bosh_job_name`, `bosh_job_id`, `bosh_job_index` |
| *metrics.namespace*_job_cpu_wait | BOSH Job CPU Wait | `environment`, `bosh_deployment`, `bosh_job_name`, `bosh_job_id`, `bosh_job_index` |
| *metrics.namespace*_job_mem_kb | BOSH Job Memory KB | `environment`, `bosh_deployment`, `bosh_job_name`, `bosh_job_id`, `bosh_job_index` |
| *metrics.namespace*_job_mem_percent | BOSH Job Memory Percent | `environment`, `bosh_deployment`, `bosh_job_name`, `bosh_job_id`, `bosh_job_index` |
| *metrics.namespace*_job_swap_kb | BOSH Job Swap KB | `environment`, `bosh_deployment`, `bosh_job_name`, `bosh_job_id`, `bosh_job_index` |
| *metrics.namespace*_job_swap_percent | BOSH Job Swap Percent | `environment`, `bosh_deployment`, `bosh_job_name`, `bosh_job_id`, `bosh_job_index` |
| *metrics.namespace*_job_system_disk_inode_percent | BOSH Job System Disk Inode Percent | `environment`, `bosh_deployment`, `bosh_job_name`, `bosh_job_id`, `bosh_job_index` |
| *metrics.namespace*_job_system_disk_percent | BOSH Job System Disk Percent | `environment`, `bosh_deployment`, `bosh_job_name`, `bosh_job_id`, `bosh_job_index` |
| *metrics.namespace*_job_ephemeral_disk_inode_percent | BOSH Job Ephemeral Disk Inode Percent | `environment`, `bosh_deployment`, `bosh_job_name`, `bosh_job_id`, `bosh_job_index` |
| *metrics.namespace*_job_ephemeral_disk_percent | BOSH Job Ephemeral Disk Percent | `environment`, `bosh_deployment`, `bosh_job_name`, `bosh_job_id`, `bosh_job_index` |
| *metrics.namespace*_job_persistent_disk_inode_percent | BOSH Job Persistent Disk Inode Percent | `environment`, `bosh_deployment`, `bosh_job_name`, `bosh_job_id`, `bosh_job_index` |
| *metrics.namespace*_job_persistent_disk_percent | BOSH Job Persistent Disk Percent | `environment`, `bosh_deployment`, `bosh_job_name`, `bosh_job_id`, `bosh_job_index` |

## Contributing

Refer to the [contributing guidelines][contributing].

## License

Apache License 2.0, see [LICENSE][license].

[binaries]: https://github.com/bosh-prometheus/bosh_tsdb_exporter/releases
[bosh]: https://bosh.io
[bosh-tsdb]: http://bosh.io/docs/hm-config.html#tsdb
[contributing]: https://github.com/bosh-prometheus/bosh_tsdb_exporter/blob/master/CONTRIBUTING.md
[faq]: https://github.com/bosh-prometheus/bosh_tsdb_exporter/blob/master/FAQ.md
[golang]: https://golang.org/
[license]: https://github.com/bosh-prometheus/bosh_tsdb_exporter/blob/master/LICENSE
[prometheus]: https://prometheus.io/
[prometheus-boshrelease]: https://github.com/bosh-prometheus/prometheus-boshrelease
