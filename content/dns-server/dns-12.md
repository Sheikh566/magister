## TCP Fallback and Truncated UDP Responses

Classic DNS over UDP is strictly limited to 512 bytes (unless the client and server negotiate larger messages using an extension called EDNS). 

When a server generates a response that is larger than 512 bytes, it cannot fit the entire payload into a single UDP datagram. Instead, the server must:
1. Truncate the response so it fits within 512 bytes.
2. Set the `TC` (Truncated) flag to `1` in the DNS Header.

When a client receives a response with the `TC` flag set, it knows the answer is incomplete. The client will then automatically open a TCP connection to the server and retry the exact same query. Because TCP is a stream-oriented protocol, it does not have the 512-byte limitation.

### DNS over TCP

Because TCP is a continuous stream of bytes (unlike UDP, which sends discrete datagrams), the server and client need a way to know where one DNS message ends and the next begins.

To solve this, DNS over TCP wraps each DNS message with a **two-byte big-endian length prefix**. This prefix specifies the size of the DNS message that immediately follows it.

For example, if your DNS message is 47 bytes long, the TCP payload will look like this:

```text
00 2f <47 bytes of DNS message>
```
*(Note: `00 2f` is `47` in hexadecimal).*

When reading from a TCP socket, you must first read these 2 bytes to determine how much data to read next. When writing a response to a TCP socket, you must prepend the 2-byte length of your response before sending the actual DNS packet.

### Test Record

For this lesson, queries for `large.magister.test` should produce enough `A` records to make the complete DNS response larger than 512 bytes. You can achieve this by returning 30+ `A` records (each with a different IP, or even the same IP) in the Answer section.

### External Resources

* [RFC 7766: DNS transport over TCP](https://www.rfc-editor.org/rfc/rfc7766)
* [RFC 1035: DNS over TCP and UDP](https://www.rfc-editor.org/rfc/rfc1035#section-4.2)

### Your Task

1. **UDP Truncation:** When `large.magister.test` is queried over UDP, generate the large response. Realize it exceeds 512 bytes, truncate it, and return a valid DNS response with the `TC` flag set to `1`.
2. **TCP Server:** Listen for TCP connections on the same port your UDP server is using.
3. **TCP Handling:** When the tester connects via TCP and retries the query, read the 2-byte length prefix, then read the DNS query. 
4. **TCP Response:** Generate the full, un-truncated response (which will be >512 bytes). Ensure the `TC` flag is `0`. Prepend the 2-byte length of your response, and send it back over the TCP connection.
