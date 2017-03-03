This is a resolver to aid with applications that do not support [RFC 2732](https://www.ietf.org/rfc/rfc2732.txt).

Run the server ```go run server``` and then any requests for subdomains
ipv6-literal in the form of an IPv6 address with colons replaced with - will
resolve to the corresponding IPv6 address. For example ```dig @localhost -p
8053 2001-db8--1.ipv6-literal AAAA``` will return the AAAA record for
```2001:db8::1```.

Eventually I intend to run this service under a real domain, but this project needs a bit of polishing up first.
