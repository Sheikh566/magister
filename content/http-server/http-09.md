## Robust Parsing and Error Handling

On the open internet, your server will receive all kinds of garbage data. Broken clients, network glitches, or malicious scanners will frequently send bytes that don't look like valid HTTP at all.

A production-grade server must be robust. It cannot panic, crash, or enter an infinite loop when given malformed input. 

### The 400 Bad Request Status

If a client sends a request that violates the HTTP specification (e.g., a missing HTTP version, missing headers, or invalid framing), the server should gracefully return a `400 Bad Request` status code. 

More importantly, returning an error should be handled like any other request—the server should send the error response and then go back to waiting for the next request (or close the connection if keep-alive isn't possible). It must not crash the entire process.

### External Resources
* [MDN: 400 Bad Request](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/400)
* [OWASP: What is a Bad Request?](https://owasp.org/www-community/attacks/HTTP_Request_Smuggling) (Context on why robust parsing matters)

### Your Task
Ensure your parser fails gracefully. If the tester sends intentionally malformed bytes (like "NOT HTTP AT ALL"), your server should reply with a `400 Bad Request`. The tester will also send well-formed requests to verify that your server stays alive and handles both malformed and valid input correctly—the order of these requests is not fixed, so your parser must actually detect and reject bad input rather than hardcoding responses.
