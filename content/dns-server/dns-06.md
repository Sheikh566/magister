## Parsing Questions

Hardcoded answers only work if you can read what the client asked for. This lesson moves your server from writing known bytes to parsing `QNAME`, `QTYPE`, and `QCLASS`.

Your local zone still has one IPv4 record:

```text
magister.test.  A  203.0.113.10
```

If the requested name is not in your local zone and no upstream resolver is configured yet, return `NXDOMAIN` (`RCODE=3`).
NXDOMAIN stands for Non-Existent Domain.

### External Resources

* [RFC 1035: Question section](https://www.rfc-editor.org/rfc/rfc1035#section-4.1.2)
* [Cloudflare Learning Center: What is an A record?](https://www.cloudflare.com/learning/dns/dns-records/dns-a-record/)

### Your Task

Decode the question section. Return the known `A` answer for `magister.test`, and return `NXDOMAIN` for a missing name such as `missing.magister.test`.
