## Supporting IPv6 and AAAA Records

While `A` records map domain names to 32-bit IPv4 addresses, `AAAA` (pronounced "quad-A") records map domain names to 128-bit IPv6 addresses. 

The shape of the `AAAA` resource record is exactly the same as an `A` record, with two key differences:
1. The `TYPE` field is `28` (instead of `1`).
2. The `RDLENGTH` is `16` (instead of `4`), and the `RDATA` contains the 16 bytes of the IPv6 address.

### Encoding an IPv6 Address

IPv6 addresses are typically written as eight groups of four hexadecimal digits, separated by colons (e.g., `2001:0db8:0000:0000:0000:0000:0000:0053`). They are often abbreviated by removing leading zeros and replacing consecutive groups of zeros with a double colon (`::`).

For example, the abbreviated address `2001:db8::53` expands to:
`2001:0db8:0000:0000:0000:0000:0000:0053`

To encode this into the 16-byte `RDATA` payload, you write each 16-bit hex group as two bytes in network byte order (big-endian):
`20 01 0d b8 00 00 00 00 00 00 00 00 00 00 00 53`

### The Test Record

This course uses the documentation IPv6 prefix `2001:db8::/32`. Your local zone should contain this deterministic record:

```text
ipv6.magister.test.  AAAA  2001:db8::53
```

### External Resources

* [RFC 3596: DNS extensions to support IPv6](https://www.rfc-editor.org/rfc/rfc3596)
* [RFC 3849: IPv6 documentation prefix](https://www.rfc-editor.org/rfc/rfc3849)

### Your Task

Recognize incoming queries where `QTYPE=28` (`AAAA`). 

When the tester asks for `ipv6.magister.test`, answer with the `AAAA` record for `2001:db8::53`. Make sure to set `TYPE=28`, `RDLENGTH=16`, and encode the IPv6 address as exactly 16 bytes in the answer `RDATA`.
