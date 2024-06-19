package collector

import (
	"encoding/xml"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"os/exec"
	"sync"
)

type Config struct {
	Path     string
	Command  string
	User     string
	Password string
	FetchFn  func(string) ([]byte, error)
}

type Exporter struct {
	mutex         sync.RWMutex
	clustsafeHost string
	fetchFn       func(string) ([]byte, error)

	humidity         *prometheus.Desc
	temperature      *prometheus.Desc
	outletStatus     *prometheus.Desc
	lineStatus       *prometheus.Desc
	powerConsumption *prometheus.Desc
	status           *prometheus.Desc
	logger           log.Logger
}

func NewExporter(clustsafeHost string, config *Config, logger log.Logger) *Exporter {
	if config.FetchFn == nil {
		config.FetchFn = func(host string) ([]byte, error) {
			return exec.Command(config.Path, "--host", host,
				"--user", config.User, "--password", config.Password,
				"-a",           // accept all SSL certificates
				"-x",           // output in XML format
				config.Command, // request information (default: sensors)
			).Output()
		}
	}

	return &Exporter{
		clustsafeHost: clustsafeHost,
		fetchFn:       config.FetchFn,
		logger:        logger,

		humidity:         prometheus.NewDesc("clustsafe_humidity", "The humidity in percentage.", []string{"sensor"}, nil),
		temperature:      prometheus.NewDesc("clustsafe_temperature", "The temperature in celsius.", []string{"sensor"}, nil),
		outletStatus:     prometheus.NewDesc("clustsafe_outlet_up", "The status of an outlet in the module.", []string{"module", "outlet"}, nil),
		lineStatus:       prometheus.NewDesc("clustsafe_line_up", "The status of in input power line in the module.", []string{"module", "line"}, nil),
		powerConsumption: prometheus.NewDesc("clustsafe_power_consumption_watts", "The real power consumption in Watts.", []string{"module"}, nil),
		status:           prometheus.NewDesc("clustsafe_up", "Was the last scrape of ClustSafe successful.", nil, nil),
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.humidity
	ch <- e.temperature
	ch <- e.outletStatus
	ch <- e.lineStatus
	ch <- e.powerConsumption
	ch <- e.status
}

func (e *Exporter) parseOutput(data []byte) (ClustsafeSchema, error) {
	var response ClustsafeSchema
	err := xml.Unmarshal(data, &response)
	return response, err
}

func (e *Exporter) scrape(ch chan<- prometheus.Metric) (up float64) {
	out, err := e.fetchFn(e.clustsafeHost)

	if err != nil {
		level.Error(e.logger).Log("msg", "Can't execute clustsafe command", "err", err)
		return 0
	}

	data, err := e.parseOutput(out)
	if err != nil {
		level.Debug(e.logger).Log("msg", "Received clustsafe output", "output", string(out))
		level.Error(e.logger).Log("msg", "Failed parsing clustsafe output", "err", err)
		return 0
	}

	for _, module := range data.Modules.Clustsafe {
		if module.Status == "connected" {
			ch <- prometheus.MustNewConstMetric(e.powerConsumption, prometheus.GaugeValue, module.Power.RealPower, module.ID)

			for _, outlet := range module.Outlets.Outlet {
				var status float64
				if outlet.Status == "on" {
					status = 1.0
				}

				ch <- prometheus.MustNewConstMetric(e.outletStatus, prometheus.GaugeValue, status, module.ID, outlet.ID)
			}

			for _, line := range module.Lines.Line {
				var status float64
				if line.Status == "connected" {
					status = 1.0
				}

				ch <- prometheus.MustNewConstMetric(e.lineStatus, prometheus.GaugeValue, status, module.ID, line.ID)
			}
		}
	}

	for _, sensor := range data.Sensors.Sensor {
		if sensor.Type == "humidity" {
			ch <- prometheus.MustNewConstMetric(e.humidity, prometheus.GaugeValue, sensor.Value, "humidity")
		} else if sensor.Type == "temperature" {
			ch <- prometheus.MustNewConstMetric(e.temperature, prometheus.GaugeValue, sensor.Value, "temperature")
		} else if sensor.Type == "dallas" && sensor.Value > 0 {
			ch <- prometheus.MustNewConstMetric(e.temperature, prometheus.GaugeValue, sensor.Value, "dallas"+sensor.ID)
		}
	}

	return 1
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()

	up := e.scrape(ch)

	ch <- prometheus.MustNewConstMetric(e.status, prometheus.GaugeValue, up)
}
