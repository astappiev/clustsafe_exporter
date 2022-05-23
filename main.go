package main

import (
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

var (
	listenAddress = flag.String("web.listen-address", ":9879", "Address to listen on for telemetry")
	metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics")
)

func remoteHandler(w http.ResponseWriter, r *http.Request) {
	clustsafeUser := os.Getenv("CLUSTSAFE_USER")
	clustsafePassword := os.Getenv("CLUSTSAFE_PASSWORD")

	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, "'target' parameter must be specified", 400)
		return
	}

	registry := prometheus.NewRegistry()
	remoteCollector := ClustsafeCollector(target, clustsafeUser, clustsafePassword)
	registry.MustRegister(remoteCollector)
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func main() {
	flag.Parse()

	godotenv.Load()
	clustsafeUser := os.Getenv("CLUSTSAFE_USER")
	clustsafePassword := os.Getenv("CLUSTSAFE_PASSWORD")

	if len(clustsafeUser) == 0 || len(clustsafePassword) == 0 {
		log.Println("No credentials given! Check env variables.")
	}

	log.Println("Listening on address: " + *listenAddress)
	http.HandleFunc(*metricsPath, remoteHandler)
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
				<label>Target:</label> <input type="text" name="target" placeholder="X.X.X.X" value="1.2.3.4"><br>
				<input type="submit" value="Submit">
				</form>
			 </body>
            </html>`))
	})
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
