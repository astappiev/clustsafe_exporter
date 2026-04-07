package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	"github.com/prometheus/exporter-toolkit/web/kingpinflag"

	"github.com/astappiev/clustsafe_exporter/collector"
)

const (
	envVarUser     = "CLUSTSAFE_USER"
	envVarPassword = "CLUSTSAFE_PASSWORD"
)

var (
	metricsPath      = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
	clustsafePath    = kingpin.Flag("clustsafe.path", "The ClustSafe command to use. If empty, remote fetching will be used.").String()
	clustsafeCommand = kingpin.Flag("clustsafe.command", "The command to execute, can be `clustsafes`, `sensors` or `all`.").Default("all").String()
	toolkitFlags     = kingpinflag.AddFlags(kingpin.CommandLine, ":9879")
)

func remoteHandler(config *collector.Config, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("target")
		if target == "" {
			http.Error(w, "'target' parameter must be specified", http.StatusBadRequest)
			return
		}

		registry := prometheus.NewRegistry()
		registry.MustRegister(collector.NewExporter(target, config, logger))
		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	}
}

func main() {
	promlogConfig := &promslog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("clustsafe_exporter"))
	kingpin.CommandLine.UsageWriter(os.Stdout)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promslog.New(promlogConfig)

	logger.Info("Starting clustsafe_exporter", "version", version.Info())
	logger.Info("Build context", "context", version.BuildContext())

	clustsafeUser := os.Getenv(envVarUser)
	clustsafePassword := os.Getenv(envVarPassword)

	if clustsafeUser == "" || clustsafePassword == "" {
		logger.Error("No credentials given!", "details", fmt.Sprintf("Please set %s and %s environment variables", envVarUser, envVarPassword))
		os.Exit(1)
	}

	config := &collector.Config{
		Path:     *clustsafePath,
		Command:  *clustsafeCommand,
		User:     clustsafeUser,
		Password: clustsafePassword,
	}

	if *clustsafePath == "" {
		config.FetchFn = collector.NewRemoteFetcher(config.User, config.Password, config.Command)
	}

	prometheus.MustRegister(versioncollector.NewCollector("clustsafe_exporter"))
	http.HandleFunc(*metricsPath, remoteHandler(config, logger))

	if *metricsPath != "/" && *metricsPath != "" {
		landingConfig := web.LandingConfig{
			Name:        "ClustSafe Exporter",
			Description: "Prometheus Exporter for ClustSafe Rack PDU",
			Version:     version.Info(),
			Form: web.LandingForm{
				Action: *metricsPath,
				Inputs: []web.LandingFormInput{
					{
						Label:       "Target host",
						Name:        "target",
						Placeholder: "10.0.0.1",
					},
				},
			},
		}
		landingPage, err := web.NewLandingPage(landingConfig)
		if err != nil {
			logger.Error("Error creating landing page", "err", err)
			os.Exit(1)
		}
		http.Handle("/", landingPage)
	}

	server := &http.Server{}
	if err := web.ListenAndServe(server, toolkitFlags, logger); err != nil {
		logger.Error("Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}
