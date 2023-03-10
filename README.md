# `connectivity`

`connectivity` is a tool for validating network connectivity requirements.

Given a set of URLs, connectivity attempts to validate each one as thoroughly as possible, step by step.

1. The URL is parsed for validity, and to understand what the destination port should be. For example, `https://` URLs assume an implicit port 443. The protocol and port of the intended destination are logged as `connectivity` interprets them.

2. If a URL contains a hostname, it is resolved.

3. The routability of each address is evaluated. Failures in routability are non-fatal; that is, the attempt to determine a viable route is only useful for producing robust logging upon failure, but does not indicate a failure in the actual routability from the host.

4. Each address returned by DNS is dialed to validate the network path (or pinged, in the case of URLs like `icmp://<host>`.

5. Supported schemes are then validated at the application-level (attempting to make real HTTP requests in the case of `http://` or `https://`, for example).

If any step in the validation process fails, relevant debugging information is logged.

## Configuration

`connectivity` looks for a configuration file in the following three locations, and uses the first one it finds:

- `./connectivity.yml`
- `~/.connectivity.yml`
- `/etc/connectivity.yml`

Without a configuration file, you might use `connectivity` to validate multiple types of connectivity to one or more destinations:

```bash
connectivity [check|wait|monitor] icmp://example.com tcp://example.com:443/ udp://example.com:53 http://example.com/health https://example.com/health
```

With a YAML configuration file, you can simply invoke `connectivity` using `connectivity [check|wait|monitor]`. Configuration uses arbitrary key-value pairs, where the key is used as a label for logging purposes (instead of logging the entire URL), and the value is the URL to be validated (which would otherwise be passed on the command line).

```yaml
---
Ping: icmp://example.com
TCP: tcp://example.com:443
UDP: udp://example.com:53
HTTP: http://example.com/health
HTTPS: https://example.com/health
```

## Supported schemes

`connectivity` can be used to validate connectivity at various different layers of the [OSI model](https://en.wikipedia.org/wiki/OSI_model).

OSI Layer 3 (Network):

- `icmp://`: Validate the destination by pinging it (ICMP). You must not specify a port. Response time metrics are emitted via statsd.

OSI Layer 4 (Transport):

- `tcp://`: Simply dial the host at the specified port and hangup (a port is required). This is useful for validating raw connectivity (similar to `netcat`) without validating anything futher about the connection. Layer 7 firewalls may allow this check to succeed, but deny the application-specific traffic, such as TLS negotiation.
- `udp://`: Simply dial the host at the specified port (a port is required). It is impossible to guarantee the destination was actually reached, only that packets _can_ be sent.

OSI Layer 7 (Application):

- `http://`: Make an HTTP `GET` request to the destination. An `HTTP 2xx` response is expected.
- `https://`: Make an HTTPS `GET` connection, including TLS validation. An `HTTP 2xx` response is expected.
