## Concurrency and Responsiveness

A real-world web server rarely handles clients one at a time. Dozens, hundreds, or even thousands of clients might be connected simultaneously.

If your server reads from a socket synchronously on its main thread, a slow client—one that opens a connection but takes 10 seconds to send its HTTP request—will block all other clients from connecting or receiving responses. This is a classic denial-of-service vulnerability (often called a Slowloris attack).

### Handling Multiple Clients

To stay responsive, your server needs a concurrency model. Common approaches include:
* **Multi-threading/Multi-processing:** Spawning a new OS thread or process for each connection.
* **Asynchronous I/O (Event Loop):** Using a single thread and non-blocking I/O (like Node.js or Python's `asyncio`).
* **Lightweight Threads (Goroutines):** Using language-level concurrency primitives (like Go's goroutines) which are mapped efficiently to OS threads.

### External Resources
* [Wikipedia: Concurrency (computer science)](https://en.wikipedia.org/wiki/Concurrency_(computer_science))
* [Cloudflare: What is a Slowloris attack?](https://www.cloudflare.com/learning/ddos/ddos-attack-tools/slowloris/)

### Your Task
Ensure your server can handle concurrent requests. The tester will simulate slow clients by opening connections and sending data very slowly. While these slow clients are connected and still trickling data, the tester will open a new connection and expect a fast response. Your server must serve the fast client without waiting for the slow clients to finish.
