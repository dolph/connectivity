# `connectivity`

`connectivity` is a tool for validating network connectivity requirements.

Given a set of URLs, connectivity attempts to validate each one as thoroughly
as possible, step by step.

1. The URL is parsed for validity, and to understand what the destination port
   should be. For example, `https://` URLs assume an implicit port 443. The
   protocol and port of the intended destination are logged as `connectivity`
   interprets them.
1. If a URL contains a hostname, it is resolved.
2. Each address returned by DNS is validated for routability.
3. Each address returned by DNS is dialed to validate the network path (or
   pinged, in the case of URLs like `icmp://<host>`.
4. Supported schemes are then validated at the application-level (attempting to
   make real HTTP requests in the case of `http://` or `https://`, for
   example).

If any step in the validation process fails, relevant debugging information is
logged.
