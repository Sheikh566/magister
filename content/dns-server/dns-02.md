## Writing the DNS Header

Every DNS message begins with a 12-byte header. The header contains information about the query/response, such as the transaction ID, various flags, and the number of entries in the other sections of the packet.

### Header Layout

The 12-byte header is structured as follows:

* **ID** (Packet Identifier, 16 bits): A random identifier assigned to query packets. Response packets must reply with the same ID. This is needed to differentiate responses due to the stateless nature of UDP.
* **QR** (Query Response, 1 bit): `0` for queries, `1` for responses.
* **OPCODE** (Operation Code, 4 bits): Typically always `0`, see RFC 1035 for details.
* **AA** (Authoritative Answer, 1 bit): Set to `1` if the responding server is authoritative - that is, it "owns" - the domain queried.
* **TC** (Truncated Message, 1 bit): Set to `1` if the message length exceeds 512 bytes. Traditionally a hint that the query can be reissued using TCP.
* **RD** (Recursion Desired, 1 bit): Set by the sender of the request if the server should attempt to resolve the query recursively if it does not have an answer readily available.
* **RA** (Recursion Available, 1 bit): Set by the server to indicate whether or not recursive queries are allowed.
* **Z** (Reserved, 3 bits): Originally reserved for later use, but now used for DNSSEC queries.
* **RCODE** (Response Code, 4 bits): Set by the server to indicate the status of the response, i.e. whether or not it was successful or failed.
* **QDCOUNT** (Question Count, 16 bits): The number of entries in the Question Section.
* **ANCOUNT** (Answer Count, 16 bits): The number of entries in the Answer Section.
* **NSCOUNT** (Authority Count, 16 bits): The number of entries in the Authority Section.
* **ARCOUNT** (Additional Count, 16 bits): The number of entries in the Additional Section.

All multi-byte fields are encoded in network byte order, also called big-endian.

For this lesson, you will primarily be concerned with:
- **ID**: Needs to match the request's ID.
- **QR**: Needs to be set to `1` (indicating a response).
- **RCODE**: Needs to be set to `0` (indicating no error).

### External Resources

* [Emil Hernvall's DNS Guide](https://github.com/EmilHernvall/dnsguide/blob/b52da3b32b27c81e5c6729ac14fe01fef8b1b593/chapter1.md#1---the-dns-protocol)
* [RFC 1035: Message format](https://www.rfc-editor.org/rfc/rfc1035#section-4.1.1)
* [Wikipedia: Domain Name System message format](https://en.wikipedia.org/wiki/Domain_Name_System#DNS_message_format)

### Your Task

Read the request header and return a DNS response header. Copy the request transaction ID, set `QR=1`, and return `RCODE=0` for a standard query.
