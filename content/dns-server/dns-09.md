## Following Referrals Recursively

Recursive resolution is what turns DNS from a simple forwarding proxy into a true resolver. Instead of just asking one upstream server to do the work, a recursive resolver does the legwork itself. It starts by asking a root nameserver, follows `NS` (Name Server) referrals, and keeps asking more specific authoritative servers until it finds the final answer.

### The Recursive Process

When a client asks your server for `recursive.magister.test`, your server must perform the following steps:

1. **Query the Root:** Send the query for `recursive.magister.test` to the Root Nameserver.
2. **Handle the Referral:** The Root server won't know the IP address, but it knows who handles the `.test` Top-Level Domain (TLD). It responds with a referral: an `NS` record pointing to the TLD nameserver (e.g., `ns.test.magister`).
3. **Query the TLD:** Send the same query for `recursive.magister.test` to the TLD nameserver.
4. **Handle the Next Referral:** The TLD server responds with another referral, pointing to the Authoritative nameserver for `magister.test`.
5. **Query the Authoritative Server:** Send the query one last time to the Authoritative nameserver.
6. **Return the Answer:** The Authoritative server responds with the final `A` record containing the IP address. You return this final answer to the original client.

### Test Environment Configuration

In real DNS, nameservers are reached on port 53, and their IP addresses are often provided in the "Additional" section of the referral response (known as glue records). 

Because Magister tests run on unprivileged local ports, the tester provides a map from nameserver names to their local `host:port` addresses via environment variables:

```text
DNS_ROOT_HINTS
DNS_ROOT_ADDR
DNS_AUTHORITY_ADDRS
DNS_AUTHORITY_MAP
```

* Use `DNS_ROOT_HINTS` or `DNS_ROOT_ADDR` as the address for the very first query (Step 1).
* Use the authority map (`DNS_AUTHORITY_MAP` or `DNS_AUTHORITY_ADDRS`) to look up the `host:port` for the nameservers returned in the referrals (e.g., translating `ns.test.magister` to `127.0.0.1:5301`). The map is usually provided as a comma-separated list of `name=address` pairs (e.g., `ns.test.magister=127.0.0.1:5301,ns2.test.magister=127.0.0.1:5302`).

### External Resources

* [Cloudflare Learning Center: Recursive DNS resolver](https://www.cloudflare.com/learning/dns/what-is-recursive-dns/)
* [RFC 1034: Domain concepts and facilities](https://www.rfc-editor.org/rfc/rfc1034)

### Your Task

Resolve `recursive.magister.test` without using the forwarding upstream from the previous lesson. 

1. Query the fake root server.
2. Parse the response to find the `NS` referral.
3. Look up the referred nameserver's address in the environment variables.
4. Query the fake TLD server.
5. Follow the next referral to the fake authoritative server.
6. Return the final `A` answer (`198.51.100.99`) to the client.
