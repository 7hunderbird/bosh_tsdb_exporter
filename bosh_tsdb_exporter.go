package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"

	"github.com/bosh-prometheus/bosh_tsdb_exporter/collectors"
)

var (
	metricsNamespace = flag.String(
		"metrics.namespace", "bosh_tsdb",
		"Metrics Namespace ($BOSH_TSDB_EXPORTER_METRICS_NAMESPACE).",
	)

	metricsEnvironment = flag.String(
		"metrics.environment", "",
		"Environment label to be attached to metrics ($BOSH_TSDB_EXPORTER_METRICS_ENVIRONMENT).",
	)

	tsdbListenAddress = flag.String(
		"tsdb.listen-address", ":13321",
		"Address to listen on for the TSDB collector ($BOSH_TSDB_EXPORTER_TSDB_LISTEN_ADDRESS).",
	)

	showVersion = flag.Bool(
		"version", false,
		"Print version information.",
	)

	listenAddress = flag.String(
		"web.listen-address", ":9194",
		"Address to listen on for web interface and telemetry ($BOSH_TSDB_EXPORTER_WEB_LISTEN_ADDRESS).",
	)

	metricsPath = flag.String(
		"web.telemetry-path", "/metrics",
		"Path under which to expose Prometheus metrics ($BOSH_TSDB_EXPORTER_WEB_TELEMETRY_PATH).",
	)

	authUsername = flag.String(
		"web.auth.username", "",
		"Username for web interface basic auth ($BOSH_TSDB_EXPORTER_WEB_AUTH_USERNAME).",
	)

	authPassword = flag.String(
		"web.auth.password", "",
		"Password for web interface basic auth ($BOSH_TSDB_EXPORTER_WEB_AUTH_PASSWORD).",
	)

	tlsCertFile = flag.String(
		"web.tls.cert_file", "",
		"Path to a file that contains the TLS certificate (PEM format). If the certificate is signed by a certificate authority, the file should be the concatenation of the server's certificate, any intermediates, and the CA's certificate ($BOSH_TSDB_EXPORTER_WEB_TLS_CERTFILE).",
	)

	tlsKeyFile = flag.String(
		"web.tls.key_file", "",
		"Path to a file that contains the TLS private key (PEM format) ($BOSH_TSDB_EXPORTER_WEB_TLS_KEYFILE).",
	)
)

func init() {
	prometheus.MustRegister(version.NewCollector(*metricsNamespace))
}

func overrideFlagsWithEnvVars() {
	overrideWithEnvVar("BOSH_TSDB_EXPORTER_METRICS_NAMESPACE", metricsNamespace)
	overrideWithEnvVar("BOSH_TSDB_EXPORTER_METRICS_ENVIRONMENT", metricsEnvironment)
	overrideWithEnvVar("BOSH_TSDB_EXPORTER_TSDB_LISTEN_ADDRES", tsdbListenAddress)
	overrideWithEnvVar("BOSH_TSDB_EXPORTER_WEB_LISTEN_ADDRESS", listenAddress)
	overrideWithEnvVar("BOSH_TSDB_EXPORTER_WEB_TELEMETRY_PATH", metricsPath)
	overrideWithEnvVar("BOSH_TSDB_EXPORTER_WEB_AUTH_USERNAME", authUsername)
	overrideWithEnvVar("BOSH_TSDB_EXPORTER_WEB_AUTH_PASSWORD", authPassword)
	overrideWithEnvVar("BOSH_TSDB_EXPORTER_WEB_TLS_CERTFILE", tlsCertFile)
	overrideWithEnvVar("BOSH_TSDB_EXPORTER_WEB_TLS_KEYFILE", tlsKeyFile)
}

func overrideWithEnvVar(name string, value *string) {
	envValue := os.Getenv(name)
	if envValue != "" {
		*value = envValue
	}
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
	flag.Parse()
	overrideFlagsWithEnvVars()

	if *showVersion {
		fmt.Fprintln(os.Stdout, version.Print("bosh_tsdb_exporter"))
		os.Exit(0)
	}

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
