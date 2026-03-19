## Your First HTTP Conversation

Now that you can accept a raw TCP connection, it's time to speak the language of the web: Hypertext Transfer Protocol (HTTP). HTTP is a plain-text protocol where the client sends a **Request** and the server replies with a **Response**.

A valid minimal HTTP/1.1 response must contain a Status Line and Headers, followed by a blank line (which marks the end of the headers), and finally an optional Body. 

### Understanding HTTP Framing

HTTP/1.1 relies heavily on **CRLF** (Carriage Return and Line Feed, commonly written as `\r\n`) to delimit lines. The end of the headers is strictly defined by an empty line—meaning two consecutive CRLFs (`\r\n\r\n`).

Here is a perfectly valid response:
```http
HTTP/1.1 200 OK\r\n
Content-Length: 0\r\n
\r\n
```

### External Resources
* [MDN: An overview of HTTP](https://developer.mozilla.org/en-US/docs/Web/HTTP/Overview)
* [MDN: HTTP Messages](https://developer.mozilla.org/en-US/docs/Web/HTTP/Messages)
* [RFC 9112: HTTP/1.1 Message Syntax and Routing](https://httpwg.org/specs/rfc9112.html)

### Your Task
When the tester connects and sends a `GET / HTTP/1.1` request, read the request and write back a `200 OK` response with a `Content-Length` of `0`. You must ensure that your response is properly framed with CRLF line endings.
