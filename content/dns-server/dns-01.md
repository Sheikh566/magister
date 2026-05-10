## Listening for DNS over UDP

DNS usually starts with a UDP datagram. Unlike TCP, there is no connection to accept: your server binds a UDP socket, waits for packets, and sends response packets back to the sender's address.

For this first lesson, you do not need to understand the whole DNS packet yet. The tester sends a valid DNS query and only checks that your process sends a UDP response before the timeout.

### Port Configuration

Listen on the port provided by the environment. Magister sets `DNS_SERVER_PORT`, `COURSE_PORT`, and `PORT` to the same value.

### External Resources

* [RFC 1035: DNS transport basics](https://www.rfc-editor.org/rfc/rfc1035)
* [Cloudflare Learning Center: What is DNS?](https://www.cloudflare.com/learning/dns/what-is-dns/)

### Your Task

Start a UDP server on the configured port. When a DNS query datagram arrives, send any response datagram back to the sender. Later lessons will make that response valid DNS.
