FROM        quay.io/prometheus/busybox:latest
MAINTAINER  Ferran Rodenas <frodenas@gmail.com>

COPY bosh_tsdb_exporter /bin/bosh_tsdb_exporter

ENTRYPOINT ["/bin/bosh_tsdb_exporter"]
EXPOSE     9194
EXPOSE     13321