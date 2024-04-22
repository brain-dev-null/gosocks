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

### 2024-04-21

- Reset repo
- Implement partial request parsing (missing request content)
- Created test for request parsing basic case

### 2024-04-22

- Added request content parsing
- Added `HttpRequest.String()` function for pretty-printing requests
