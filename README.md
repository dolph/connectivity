# `connectivity`

`connectivity` is a tool for verifying network connectivity requirements.

Given a set of URLs, connectivity attempts to reach each one as thoroughly as
possible. If a URL contains a hostname, it is first resolved. Each DNS record
is then individually dialed or pinged (in the case of `icmp://`). Supported
schemes are then verified at the application level (attempting to make real
HTTP requests in the case of `http://` or `https://`, for example).
