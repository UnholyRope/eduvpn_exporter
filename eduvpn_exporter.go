package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	colversion "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
)

func main() {
	var (
		webConfig   = webflag.AddFlags(kingpin.CommandLine, ":10036")
		metricsPath = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics. Relative to web.route-prefix").Default("/metrics").String()
		statusCmd   = kingpin.Flag("status-cmd", "Path to the vpn-user-portal-status command.").Default(statusCmd).String()
		statusFlags = kingpin.Flag("status-flags", "Flags to use when getting the vpn user portal status.").HintOptions("connections", "all").Enums("connections", "all")
	)

	promlogConfig := &promslog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("eduvpn_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promslog.New(promlogConfig)

	registry := prometheus.NewPedanticRegistry()

	registry.MustRegister(
		collectors.NewGoCollector(),
		colversion.NewCollector("eduvpn_exporter"),
		newEduVPNCollector(metrics, *logger, *statusCmd, *statusFlags),
	)

	landingConfig := web.LandingConfig{
		Name:        "eduvpn_exporter",
		Description: "Scrapes metrics from the eduVPN user portal and transforms them into Prometheus metrics.",
		Links: []web.LandingLinks{
			{Address: *metricsPath, Text: "Metrics", Description: "Metrics from the eduVPN user portal."},
		},
		Version: version.Version,
	}

	landingPage, err := web.NewLandingPage(landingConfig)
	if err != nil {
		logger.Error("Error creating landing page.", "err", err)
		os.Exit(1)
	}

	var formattedMetricsPath string
	if !strings.HasPrefix(*metricsPath, "/") {
		formattedMetricsPath = "/" + *metricsPath
	} else {
		formattedMetricsPath = *metricsPath
	}

	http.Handle("/", landingPage)
	http.Handle(formattedMetricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{Registry: registry, EnableOpenMetrics: true}))
	srv := &http.Server{}
	if err := web.ListenAndServe(srv, webConfig, logger); err != nil {
		logger.Error("Error starting HTTP server.", "err", err)
		os.Exit(1)
	}
}
