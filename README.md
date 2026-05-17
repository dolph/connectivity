# `connectivity`

`connectivity` is a tool for validating network connectivity requirements.

Given a set of URLs, connectivity attempts to validate each one as thoroughly as possible, step by step.

1. The URL is parsed for validity, and to understand what the destination port should be. For example, `https://` URLs assume an implicit port 443. The protocol and port of the intended destination are logged as `connectivity` interprets them.

2. If a URL contains a hostname, it is resolved.

3. The routability of each address is evaluated. Failures in routability are non-fatal; that is, the attempt to determine a viable route is only useful for producing robust logging upon failure, but does not indicate a failure in the actual routability from the host.

4. Each address returned by DNS is dialed to validate the network path (or pinged, in the case of URLs like `icmp://<host>`). Every returned address must pass for the destination to be considered reachable.

5. Supported schemes are then validated at the application-level (attempting to make real HTTP requests in the case of `http://` or `https://`, for example).

If any step in the validation process fails, relevant debugging information is logged.

## Configuration

`connectivity` looks for a configuration file in the following locations, and uses the first one it finds:

- `./connectivity.yml`
- `./connectivity.yaml`
- `~/.connectivity.yml`
- `~/.connectivity.yaml`
- `/etc/connectivity.yml`
- `/etc/connectivity.yaml`

Without a configuration file, you might use `connectivity` to validate multiple types of connectivity to one or more destinations:

```bash
connectivity check icmp://example.com tcp://example.com:443/ udp://example.com:53 http://example.com/health https://example.com/health
```

With a YAML configuration file, you can simply invoke `connectivity` using `connectivity check`, `connectivity wait`, or `connectivity monitor`. Configuration uses arbitrary key-value pairs, where the key is used as a label for logging purposes (instead of logging the entire URL), and the value is the URL to be validated (which would otherwise be passed on the command line).

If both a configuration file and command-line URLs are provided, command-line URLs replace the configured destinations for that run.

```yaml
---
Ping: icmp://example.com
TCP: tcp://example.com:443
UDP: udp://example.com:53
HTTP: http://example.com/health
HTTPS: https://example.com/health
```

### Commands

`connectivity` supports the following subcommands:

- `check`: check all destinations once and exit.
- `wait`: wait until all destinations validate successfully at least once.
- `waitfor`: alias for `wait`.
- `monitor`: continuously monitor all destinations.
- `validate-config`: load and validate configuration without making network requests.
- `version`: print build/version metadata.
- `help`: print general help; `connectivity help <command>` prints per-command help.

### Statsd

Connectivity emits result and timing metrics through statsd. By default metrics are sent to `127.0.0.1:8125` over UDP. The statsd settings and defaults are:

| Setting | Default |
| --- | --- |
| `statsd_host` | `127.0.0.1` |
| `statsd_port` | `8125` |
| `statsd_protocol` | `udp` |

`statsd_protocol` may be `udp` or `tcp`. Parsing custom YAML values for these settings is tracked separately in #6.

### Exit codes

- `0`: success, help/version output, or successful configuration validation.
- `1`: a connectivity check failed, or no destinations were parsed.
- `2`: invalid command or destination parse error.

## Supported schemes

`connectivity` can be used to validate connectivity at various different layers of the [OSI model](https://en.wikipedia.org/wiki/OSI_model).

OSI Layer 3 (Network):

- `icmp://`: Validate the destination by pinging it (ICMP). You must not specify a port. ICMP checks usually require root or `CAP_NET_RAW` on Linux, equivalent container capability such as `--cap-add=NET_RAW`, or platform-specific unprivileged ICMP support. Response time metrics are emitted via statsd.

OSI Layer 4 (Transport):

- `tcp://`: Simply dial the host at the specified port and hangup (a port is required). This is useful for validating raw connectivity (similar to `netcat`) without validating anything futher about the connection. Layer 7 firewalls may allow this check to succeed, but deny the application-specific traffic, such as TLS negotiation.
- `udp://`: Simply dial the host at the specified port (a port is required). It is impossible to guarantee the destination was actually reached, only that packets _can_ be sent.
- Schemes known to the system services database, such as `mysql://`, `postgresql://`, `ssh://`, or `ftp://` when those service names are registered locally, are treated as TCP dial checks using the registered default port. Service-name availability is OS-dependent. `nats://` is also supported as a special case with default port `4222`.

OSI Layer 7 (Application):

- `http://`: Make an HTTP `GET` request to the destination. An `HTTP 2xx` response is expected.
- `https://`: Make an HTTPS `GET` connection, including TLS validation. An `HTTP 2xx` response is expected.