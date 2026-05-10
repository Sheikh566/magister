## Parsing Header Flags

Now your server needs to make decisions based on the incoming header. The transaction ID still needs to be echoed, but the flags tell you whether the client sent a standard query, an unsupported opcode, or malformed counts.

DNS uses the low four bits of the flags field for `RCODE`, the response code. For unsupported opcodes, return `RCODE=4`, which means `Not Implemented`.

### External Resources

* [RFC 1035: Header flags and response codes](https://www.rfc-editor.org/rfc/rfc1035#section-4.1.1)
* [IANA: DNS parameters](https://www.iana.org/assignments/dns-parameters/dns-parameters.xhtml)

### Your Task

Parse the DNS header from each request. If `OPCODE` is not `0`, return a response with the same transaction ID, `QR=1`, and `RCODE=4`.
