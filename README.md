# gosocks

In this project I try to familiarize myself with the [Go](https://go.dev/) language as well
as the [WebSocket](https://de.wikipedia.org/wiki/WebSocket) protocol.

## Goals

My ultimate project goal is to build a non-buffering message broker.
It should be provide the following features:

1. Control Plane API in HTTP/1.1
    1. Create / delete topics
    2. Fetch data plane Metrics
2. Data Plane API in WebSocket
    1. Send messages to topics
    2. Receive messages from topics

## Progress

I will **attempt** to log my progress in this project in this section.

### 2024-05-12

- Complete ws frame deserialization
- Implement WebSocket handshake
- Add routing for WebSockets

Receival of WebSocket messages is now functional!

### 2024-05-10

- Switch from manual bitshifting to encoding/binary for multi-byte int deserialization
- Deserialize Masking Key

### 2024-05-07

- Begin WebSocket implementation with deserialization of ws frames:
    - FIN, RSV1-3 flags
    - Opcode
    - Masked flag
    - Payload length

### 2024-05-05

- Add basic server with async request handling
- Fix request parsing by relying on `Content-Length` header value when determining the number of bytes to read as the body
- Response handling + serialization
- Plaintext + JSON responses

### 2024-05-02

- Refactor and extend request routing test suite
- Refactor `HttpRequest.Path` and `HttpRequest.getPath()` function

### 2024-04-30

- Refactor request parsing and add convenience method
- Fix handling of empty request content
- Add request routing

### 2024-04-22

- Added request content parsing
- Added `HttpRequest.String()` function for pretty-printing requests

### 2024-04-21

- Reset repo
- Implement partial request parsing (missing request content)
- Created test for request parsing basic case
