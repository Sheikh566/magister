## Forwarding to an Upstream Resolver

A useful DNS server cannot know every record locally. When a client asks for a domain name that isn't in your local zone, your server can act as a middleman and forward the query to another DNS server (called an upstream resolver) to find the answer.

This process is known as **forwarding** or **recursive resolution**. 

### How Forwarding Works

1. **Receive:** Your server receives a DNS query from a client.
2. **Check Local:** You check if the requested domain is in your local zone (e.g., `magister.test`).
3. **Forward:** If it's not in your local zone, you open a new UDP socket and send the exact same DNS query packet to the upstream resolver.
4. **Wait:** You wait for the upstream resolver to send a response packet back to your server.
5. **Relay:** You take the response packet from the upstream resolver and send it back to the original client.

*Note: For this lesson, you can simply relay the exact bytes of the upstream response back to the client. You don't need to parse or modify the upstream response yet.*

### Upstream Configuration

Magister starts a fake upstream DNS server for this lesson and passes its address through these environment variables:

```text
DNS_UPSTREAM_ADDR
UPSTREAM_DNS_ADDR
COURSE_DNS_UPSTREAM_ADDR
```

They all contain the same `host:port` value (e.g., `127.0.0.1:5300`).

### External Resources

* [Cloudflare Learning Center: What is a DNS resolver?](https://www.cloudflare.com/learning/dns/dns-server-types/)
* [RFC 1035: Queries](https://www.rfc-editor.org/rfc/rfc1035#section-4)

### Your Task

When your local zone cannot answer a query, forward the original DNS query to the configured upstream resolver over UDP. Relay the upstream response back to the client. 

The tester will ask for `forward.magister.test`. Since this is not your hardcoded `magister.test` record, you should forward the query to the upstream address. The upstream will return `198.51.100.42`, which you must relay back to the tester.
