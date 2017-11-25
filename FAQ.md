# FAQ

### What metrics does this exporter report?

The BOSH TSDB Prometheus Exporter gets the metrics from the [BOSH OpenTSDB Health Monitor plugin][bosh_tsdb], who gathers them from each VM [BOSH Agent][bosh_agent]. The metrics that are being [reported][bosh_tsdb_exporter_metrics] are pretty basic, but include:

* Job metrics:
  * Health status
  * CPU
  * Load
  * Memory
  * Swap
  * System, Ephemeral and Persistent disk

### How can I get more detailed metrics from each VM?

If you want to get more detailed VM system metrics, like disk I/O, network traffic, ..., it is recommended to deploy the Prometheus [Node Exporter][node_exporter] on each VM.

### What is the recommended deployment strategy?

Prometheus advises to collocate exporters near the metrics source, in this case, that means colocating this exporter within your [BOSH Director][bosh_director] VM. We encourage you to follow this approach whenever is possible.

### I have a question but I don't see it answered at this FAQ

We will be glad to address any questions not answered here. Please, just open a [new issue][issues].

[bosh_agent]: https://bosh.io/docs/bosh-components.html#agent
[bosh_director]: http://bosh.io/docs/bosh-components.html#director
[bosh_tsdb]: http://bosh.io/docs/hm-config.html#tsdb
[bosh_tsdb_exporter_metrics]: https://github.com/bosh-prometheus/bosh_tsdb_exporter#metrics
[node_exporter]: https://github.com/prometheus/node_exporter
[issues]: https://github.com/bosh-prometheus/bosh_tsdb_exporter/issues
