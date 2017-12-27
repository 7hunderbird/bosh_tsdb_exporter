package main

import (
	"net"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/bosh-prometheus/bosh_tsdb_exporter/collectors"
)

var (
	metricsNamespace = kingpin.Flag(
		"metrics.namespace", "Metrics Namespace ($BOSH_TSDB_EXPORTER_METRICS_NAMESPACE)",
	).Envar("BOSH_TSDB_EXPORTER_METRICS_NAMESPACE").Default("bosh_tsdb").String()

	metricsEnvironment = kingpin.Flag(
		"metrics.environment", "Environment label to be attached to metrics ($BOSH_TSDB_EXPORTER_METRICS_ENVIRONMENT)",
	).Envar("BOSH_TSDB_EXPORTER_METRICS_ENVIRONMENT").Required().String()

	tsdbListenAddress = kingpin.Flag(
		"tsdb.listen-address", "Address to listen on for the TSDB collector ($BOSH_TSDB_EXPORTER_TSDB_LISTEN_ADDRESS)",
	).Envar("BOSH_TSDB_EXPORTER_TSDB_LISTEN_ADDRESS").Default(":13321").String()

	listenAddress = kingpin.Flag(
		"web.listen-address", "Address to listen on for web interface and telemetry ($BOSH_TSDB_EXPORTER_WEB_LISTEN_ADDRESS)",
	).Envar("BOSH_TSDB_EXPORTER_WEB_LISTEN_ADDRESS").Default(":9194").String()

	metricsPath = kingpin.Flag(
		"web.telemetry-path", "Path under which to expose Prometheus metrics ($BOSH_TSDB_EXPORTER_WEB_TELEMETRY_PATH)",
	).Envar("BOSH_TSDB_EXPORTER_WEB_TELEMETRY_PATH").Default("/metrics").String()

	authUsername = kingpin.Flag(
		"web.auth.username", "Username for web interface basic auth ($BOSH_TSDB_EXPORTER_WEB_AUTH_USERNAME)",
	).Envar("BOSH_TSDB_EXPORTER_WEB_AUTH_USERNAME").String()

	authPassword = kingpin.Flag(
		"web.auth.password", "Password for web interface basic auth ($BOSH_TSDB_EXPORTER_WEB_AUTH_PASSWORD)",
	).Envar("BOSH_TSDB_EXPORTER_WEB_AUTH_PASSWORD").String()

	tlsCertFile = kingpin.Flag(
		"web.tls.cert_file", "Path to a file that contains the TLS certificate (PEM format). If the certificate is signed by a certificate authority, the file should be the concatenation of the server's certificate, any intermediates, and the CA's certificate ($BOSH_TSDB_EXPORTER_WEB_TLS_CERTFILE)",
	).Envar("BOSH_TSDB_EXPORTER_WEB_TLS_CERTFILE").ExistingFile()

	tlsKeyFile = kingpin.Flag(
		"web.tls.key_file", "Path to a file that contains the TLS private key (PEM format) ($BOSH_TSDB_EXPORTER_WEB_TLS_KEYFILE)",
	).Envar("BOSH_TSDB_EXPORTER_WEB_TLS_KEYFILE").ExistingFile()
)

func init() {
	prometheus.MustRegister(version.NewCollector(*metricsNamespace))
}

type basicAuthHandler struct {
	handler  http.HandlerFunc
	username string
	password string
}

func (h *basicAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok || username != h.username || password != h.password {
		log.Errorf("Invalid HTTP auth from `%s`", r.RemoteAddr)
		w.Header().Set("WWW-Authenticate", "Basic realm=\"metrics\"")
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}
	h.handler(w, r)
	return
}

func prometheusHandler() http.Handler {
	handler := prometheus.Handler()

	if *authUsername != "" && *authPassword != "" {
		handler = &basicAuthHandler{
			handler:  prometheus.Handler().ServeHTTP,
			username: *authUsername,
			password: *authPassword,
		}
	}

	return handler
}

func main() {
	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("bosh_tsdb_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.Infoln("Starting bosh_tsdb_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	log.Infoln("TSDB listening on", *tsdbListenAddress)
	tsdbListener, err := net.Listen("tcp", *tsdbListenAddress)
	if err != nil {
		log.Errorf("Could not open TSDB listen address: %v", err)
		os.Exit(1)
	}
	defer tsdbListener.Close()

	tsdbCollector := collectors.NewHMTSDBCollector(
		*metricsNamespace,
		*metricsEnvironment,
		tsdbListener,
	)
	prometheus.MustRegister(tsdbCollector)

	handler := prometheusHandler()
	http.Handle(*metricsPath, handler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>BOSH TSDB Exporter</title></head>
             <body>
             <h1>BOSH TSDB Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})

	if *tlsCertFile != "" && *tlsKeyFile != "" {
		log.Infoln("Listening TLS on", *listenAddress)
		log.Fatal(http.ListenAndServeTLS(*listenAddress, *tlsCertFile, *tlsKeyFile, nil))
	} else {
		log.Infoln("Listening on", *listenAddress)
		log.Fatal(http.ListenAndServe(*listenAddress, nil))
	}
}
