# EduVPN Exporter for Prometheus

This is a simple exporter that transforms the status info obtained from `vpn-user-portal-status` into metrics for Prometheus consumption.

For more information about monitoring the EduVPN User Portal can be found in the [eduVPN docs](https://docs.eduvpn.org/server/v3/monitoring.html).

## Configuration

### Web configuration

Information about the web configration can be found in the [web-configuration docs of the exporter-toolkit](https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md).

The exporter by default listens on `:10036`.

### Exporter configuration
The following flags are available:
- `web.telemetry-path`: Specify the path under which the metrics can be scraped. Defaults to `/metrics`.
- `status-cmd`: The path to `vpn-user-portal-status`. Useful if it is not on the `$PATH`.
- `status-flags`: Flags to pass to `vpn-user-portal-status`. `connections` includes information about the profiles connections. `all` includes all wireguard connections in the connections info.

> [!NOTE]
> The metric `eduvpn_unique_users` only has a value if the status flag `connections` is passed to the exporter.

> [!WARNING]
> Passing the `connections` flag will introduce high cardinality metrics.

### Logging configuration
The log level can be set via `log.level`. Supported levels are `debug`, `info`, `warn` and `error`. 

`log.format` sets the output format. `logfmt` and `json` are currently supported.
