## The Foundation of the Web

Every HTTP/1.1 conversation starts with a solid foundation: a Transmission Control Protocol (TCP) connection. TCP provides a reliable, ordered, and error-checked delivery of a stream of bytes between applications running on hosts communicating via an IP network.

Before your server can understand HTTP requests, it must first be able to listen on a port and accept inbound connections.

### Key Concepts

* **Ports and Sockets:** A port is a logical endpoint for a connection. A socket is the software abstraction that applications use to read and write to that connection.
* **The Handshake:** TCP uses a three-way handshake (SYN, SYN-ACK, ACK) to establish a connection. When you "accept" a connection in your code, the operating system has usually handled this handshake for you.

### External Resources
* [Transmission Control Protocol (Wikipedia)](https://en.wikipedia.org/wiki/Transmission_Control_Protocol)
* [Beej's Guide to Network Programming](https://beej.us/guide/bgnet/)
* [MDN: How the Web works](https://developer.mozilla.org/en-US/docs/Learn/Getting_started/How_the_Web_works)

### Your Task
For this lesson, you don't need to read any bytes or send any data. Your only job is to bind to a specific port, wait for an incoming connection, and accept it successfully without immediately dropping it.
