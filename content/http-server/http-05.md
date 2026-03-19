## Inspecting Headers

HTTP Headers are key-value pairs that let the client and the server pass additional information with an HTTP request or response.

Examples of common request headers include:
* `Host`: The domain name of the server (required in HTTP/1.1).
* `User-Agent`: Information about the client's browser or software.
* `Accept`: What type of content the client can understand.

### Header Formatting Rules

Headers follow a specific format:
```
Header-Name: Header-Value\r\n
```

Importantly, **header names are case-insensitive**. `User-Agent: curl/7.68.0` is completely equivalent to `user-agent: curl/7.68.0` or `USER-AGENT: curl/7.68.0`.

### External Resources
* [MDN: HTTP Headers](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers)
* [RFC 9110: Field Names](https://httpwg.org/specs/rfc9110.html#field.names)

### Your Task
Implement a `GET /headers/<name>` route. The server should read all headers sent by the client. It should then look up the header specified by `<name>` (ignoring case) and return the value of that header as plain text in the response body.
