## Returning an A Record

Resource records are the answers DNS is built around. An `A` record maps a domain name to an IPv4 address.

### Resource Record Layout

Every resource record (including answers) follows a standard format. It starts with a preamble containing the name and metadata, followed by the record-specific data.

* **NAME** (Label Sequence): The domain name this record is for, encoded just like in the question section (e.g., `<len><text>...00`).
* **TYPE** (2 bytes): The record type. For an `A` record, this is `1`.
* **CLASS** (2 bytes): The class. In practice, this is always `1` (for `IN` or Internet).
* **TTL** (4 bytes): Time-To-Live. How long (in seconds) a resolver is allowed to cache this record before it should be requeried.
* **RDLENGTH** (2 bytes): The length of the record-specific data (`RDATA`) in bytes.
* **RDATA** (Variable): The actual data for the record.

### The A Record

For an `A` record, the `RDATA` is simply the IPv4 address stored as four bytes. Since an IPv4 address is 4 bytes long, the `RDLENGTH` is always `4`.

For example, the IP address `203.0.113.10` is encoded as the 4 bytes: `CB 00 71 0A` (in hex).

For this course, your first local zone contains one deterministic record:

```text
magister.test.  A  203.0.113.10
```

*(Note: The `203.0.113.0/24` range is reserved for documentation, so it is safe to use in tests and examples.)*

### External Resources

* [RFC 1035: Resource record format](https://www.rfc-editor.org/rfc/rfc1035#section-4.1.3)
* [RFC 5737: IPv4 documentation address blocks](https://www.rfc-editor.org/rfc/rfc5737)

### Your Task

Answer `A` queries for `magister.test` with one `IN` answer record containing `203.0.113.10`. 

To do this:
1. Set `ANCOUNT` to `1` in the DNS Header.
2. Append the Answer Record immediately after the Question section.
3. Encode the `NAME` (e.g., `magister.test` as labels), set `TYPE=1`, `CLASS=1`, choose a `TTL` (e.g., `60`), set `RDLENGTH=4`, and write the 4 bytes of the IP address in `RDATA`.
