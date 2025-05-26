package main

import (
	"encoding/json"
	"log/slog"
	"os/exec"
	"slices"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "eduvpn"
	statusCmd = "vpn-user-portal-status"
)

var knownUsers = []string{}
var uniqueUsers = 0
var eduvpnLabelNames = []string{"profile", "vpn_proto"}

var (
	metrics = map[string]*prometheus.Desc{
		"ActiveConnectionCount": prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "active_connections"), "Number of active connections.", eduvpnLabelNames, nil),
		"MaxConnectionCount":    prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "max_connections"), "Maximum number of connections.", eduvpnLabelNames, nil),
		// "OpenVPNMaxConnectionCount":      prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "openvpn_max_connection_count"), "Maximum number of OpenVPN connections.", eduvpnLabelNames, nil),
		// "OpenVPNActiveConnectionCount":   prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "openvpn_active_connection_count"), "umber of active OpenVPN connections.", eduvpnLabelNames, nil),
		// "WireGuardMaxConnectionCount":    prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "wireguard_max_connection_count"), "Maximum number of WireGuard connections.", eduvpnLabelNames, nil),
		// "WireGuardActiveConnectionCount": prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "wireguard_active_connection_count"), "Number of active WireGuard connections.", eduvpnLabelNames, nil),
		"AllocatedIPCount": prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "allocated_ips"), "Number of allocated IP addresses.", eduvpnLabelNames, nil),
		"FreeIPCount":      prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "free_ips"), "Number of free IP addresses.", eduvpnLabelNames, nil),
		"ConnectionList":   prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "connection_list"), "Information about active connections.", []string{"profile", "user_id", "ip_list", "vpn_proto"}, nil),
		"UniqueUsers":      prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "unique_users"), "Number of unique users with active connections. Only has a value if the status flag `connections` is passed to the exporter.", nil, nil),
	}

	eduVPNUp = prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "up"), "Was the last scrape of eduVPN successful?", nil, nil)
)

type EduVPNCollector struct {
	up             prometheus.Gauge
	scrapeFailures prometheus.Counter
	metrics        map[string]*prometheus.Desc
	statusCmd      string
	statusFlags    []string
	logger         slog.Logger
}

type ConnectionList struct {
	UserID   string   `json:"user_id"`
	IPList   []string `json:"ip_list"`
	VPNProto string   `json:"vpn_proto"`
}

type EduVPNStatus struct {
	ProfileID                      string           `json:"profile_id"`
	ActiveConnectionCount          int              `json:"active_connection_count"`
	MaxConnectionCount             int              `json:"max_connection_count"`
	PercentageInUse                int              `json:"percentage_in_use"`
	OpenVPNMaxConnectionCount      int              `json:"openvpn_max_connection_count"`
	OpenVPNActiveConnectionCount   int              `json:"openvpn_active_connection_count"`
	WireGuardMaxConnectionCount    int              `json:"wireguard_max_connection_count"`
	WireGuardActiveConnectionCount int              `json:"wireguard_active_connection_count"`
	WireGuardAllocatedIPCount      int              `json:"wireguard_allocated_ip_count"`
	WireGuardFreeIPCount           int              `json:"wireguard_free_ip_count"`
	WireGuardPercentageAllocated   int              `json:"wireguard_percentage_allocated"`
	ConnectionList                 []ConnectionList `json:"connection_list"`
}

func newEduVPNCollector(metrics map[string]*prometheus.Desc, logger slog.Logger, statusCmd string, statusFlags []string) *EduVPNCollector {
	return &EduVPNCollector{
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "Was the last scrape of eduVPN successful?",
		}),
		scrapeFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "scrape_failures_total",
			Help:      "Total number of failed scrapes.",
		}),
		metrics:     metrics,
		statusCmd:   statusCmd,
		statusFlags: statusFlags,
		logger:      logger,
	}
}

func (collector *EduVPNCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range metrics {
		ch <- desc
	}

	ch <- eduVPNUp
	ch <- collector.scrapeFailures.Desc()
}

func (collector *EduVPNCollector) Collect(ch chan<- prometheus.Metric) {
	up := collector.scrape(ch)

	ch <- prometheus.MustNewConstMetric(eduVPNUp, prometheus.GaugeValue, up)
	ch <- collector.scrapeFailures
}

func (collector *EduVPNCollector) scrape(ch chan<- prometheus.Metric) (up float64) {
	flags := []string{"--json"}

	for _, flag := range collector.statusFlags {
		flags = append(flags, "--"+flag)
	}

	cmd := exec.Command(collector.statusCmd, flags...)
	if output, err := cmd.Output(); err != nil {
		collector.logger.Error("Failed getting user portal status.", "err", err)

		collector.scrapeFailures.Inc()

		return 0
	} else {
		if !json.Valid(output) {
			collector.logger.Error("Invalid JSON output from user portal status command.")
			collector.logger.Debug("Invalid JSON", "output", string(output))

			collector.scrapeFailures.Inc()

			return 0
		}

		var status []EduVPNStatus

		status, err := parseJson(output)
		if err != nil {
			collector.logger.Error("Failed to parse JSON output.", "err", err)
			collector.scrapeFailures.Inc()

			return 0
		}

		for _, s := range status {
			ch <- prometheus.MustNewConstMetric(
				collector.metrics["MaxConnectionCount"],
				prometheus.GaugeValue,
				float64(s.OpenVPNMaxConnectionCount),
				s.ProfileID,
				"openvpn",
			)

			ch <- prometheus.MustNewConstMetric(
				collector.metrics["MaxConnectionCount"],
				prometheus.GaugeValue,
				float64(s.WireGuardMaxConnectionCount),
				s.ProfileID,
				"wireguard",
			)

			ch <- prometheus.MustNewConstMetric(
				collector.metrics["ActiveConnectionCount"],
				prometheus.GaugeValue,
				float64(s.OpenVPNActiveConnectionCount),
				s.ProfileID,
				"openvpn",
			)

			ch <- prometheus.MustNewConstMetric(
				collector.metrics["ActiveConnectionCount"],
				prometheus.GaugeValue,
				float64(s.WireGuardActiveConnectionCount),
				s.ProfileID,
				"wireguard",
			)

			ch <- prometheus.MustNewConstMetric(
				collector.metrics["AllocatedIPCount"],
				prometheus.GaugeValue,
				float64(s.WireGuardAllocatedIPCount),
				s.ProfileID,
				"wireguard",
			)

			ch <- prometheus.MustNewConstMetric(
				collector.metrics["FreeIPCount"],
				prometheus.GaugeValue,
				float64(s.WireGuardFreeIPCount),
				s.ProfileID,
				"wireguard",
			)

			for _, conn := range s.ConnectionList {
				ch <- prometheus.MustNewConstMetric(
					collector.metrics["ConnectionList"],
					prometheus.GaugeValue,
					float64(1),
					s.ProfileID,
					conn.UserID,
					strings.Join(conn.IPList, ","),
					conn.VPNProto,
				)

				if !slices.Contains(knownUsers, conn.UserID) {
					knownUsers = append(knownUsers, conn.UserID)
					uniqueUsers++
				}
			}
		}

		ch <- prometheus.MustNewConstMetric(
			collector.metrics["UniqueUsers"],
			prometheus.GaugeValue,
			float64(uniqueUsers),
		)
	}

	return 1
}

func parseJson(output []byte) ([]EduVPNStatus, error) {
	var status []EduVPNStatus

	if err := json.Unmarshal(output, &status); err != nil {
		return status, err
	}

	return status, nil
}
