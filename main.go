package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	envVarUser     = "CLUSTSAFE_USER"
	envVarPassword = "CLUSTSAFE_PASSWORD"
)

var (
	listenAddress    = flag.String("web.listen-address", ":9879", "Address to listen on for telemetry")
	metricsPath      = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics")
	clustsafePath    = flag.String("clustsafe.path", "cw-clustsafe", "The ClustSafe command to use")
	clustsafeCommand = flag.String("clustsafe.command", "all", "The command to execute, can be `clustsafes`, `sensors` or `all`")
)

func remoteHandler(config CollectorConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("target")
		if target == "" {
			http.Error(w, "'target' parameter must be specified", http.StatusBadRequest)
			return
		}

		registry := prometheus.NewRegistry()
		remoteCollector := ClustsafeCollector(target, config)
		registry.MustRegister(remoteCollector)
		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	}
}

func main() {
	flag.Parse()

	clustsafeUser := os.Getenv(envVarUser)
	clustsafePassword := os.Getenv(envVarPassword)

	if len(clustsafeUser) == 0 || len(clustsafePassword) == 0 {
		log.Fatalf("No credentials given! Please set %s and %s environment variables", envVarUser, envVarPassword)
	}

	config := CollectorConfig{
		clustsafePath:     *clustsafePath,
		clustsafeCommand:  *clustsafeCommand,
		clustsafeUser:     clustsafeUser,
		clustsafePassword: clustsafePassword,
	}

	log.Println("Listening on address: " + *listenAddress)
	http.HandleFunc(*metricsPath, remoteHandler(config))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>ClustSafe Exporter</title></head>
			 <style>
			  label{
				display:inline-block;
			    width:75px;
			  }
			  form label {
			    margin: 10px;
			  }
			  form input {
			    margin: 10px;
			  }
			 </style>
			 </head>
			 <body>
				<h1>ClustSafe Exporter</h1>
				<form action="` + *metricsPath + `">
				<label>Target host:</label> <input type="text" name="target" placeholder="10.0.0.1"><br>
				<input type="submit" value="Submit">
				</form>
			 </body>
            </html>`))
	})

	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		log.Fatal("HTTP listener stopped", err)
	}
}
