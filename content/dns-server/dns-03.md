## Echoing the Question

A DNS question describes what the client wants to know. It contains a domain name (`QNAME`), a record type (`QTYPE`), and a class (`QCLASS`).

Domain names are not sent as dot-separated strings (like `google.com`). Instead, they are split into parts called **labels** (e.g., `google` and `com`). Each label is encoded with a single length byte followed by the actual text of the label.

The overall format of an encoded domain name is a sequence of these labels, terminated by a zero-length label (a null byte): `<len><text><len><text>...00`.

For example, `google.com` is encoded as:

```text
06 google 03 com 00
```

Here, `06` is the length of `google`, `03` is the length of `com`, and the final `00` marks the end of the domain name.

### External Resources

* [RFC 1035: Question section format](https://www.rfc-editor.org/rfc/rfc1035#section-4.1.2)
* [RFC 1035: Domain name representation](https://www.rfc-editor.org/rfc/rfc1035#section-3.1)

### Your Task

When the tester asks for the `A` record for `google.com`, include the same question in your response. Set `QDCOUNT` to `1`, encode the name as labels, and copy `QTYPE` and `QCLASS`.
