# eventproc

A proof-of-concept for exploring event based architecture utilizing bi-directional
gRPC streams.

## Architecture

```ascii
                +-------------+
                |             |
                | Processor 1 |
                |             |
                +-----^---+---+
                      |   |
                      |   |
                      |   |
                      |   |
+---------+         +-+---v--+         +-------------+
|         +-------->|        +-------->|             |
| Client  |         | Sink   |         | Processor 2 |
|         |<--------+        |<--------+             |
+---------+         +--^---+-+         +-------------+
                       |   |
                       |   |
                       |   |
                       |   |
                       |   |
                       |   |
                  +----+---v----+
                  |             |
                  | Processor 3 |
                  |             |
                  +-------------+
```

**Client:** a gRPC client (mobile device) or a web
based HTTP client (a browser).

**Event Sink:** implements an API for
clients to simply send an Event and then receive a response. The response
could be nothing or it could be an event from one of the backend Event Processors.
This allows for both request/response style events as well as PubSub style events.

**Event Processor:** a gRPC service which events are streamed to and from.

## Implementation

Located in the `event` folder are gRPC service definitions for an Event Sink and
Event Processors along with some initial implementations of both.

Communication between the `Event Sink` and the backend `Event Processors` is done
over a bi-directional gRPC stream. The `Event Sink` uses a API defined enum for
its basis of mapping events to their respective `Event Processors`. By specifying
an event type enum, supporting new event types and their respective processors
becomes trivial since all a sink needs is to map event type(s) to processor(s).

`eventproc/main.go` exposes an Event Sink through a REST-style HTTP API.

`proc/main.go` implements a very simple gRPC based Event Processor "backend" that
simply echoes back any events streamed to it.

## Research resources

* [The Mysterious Gotcha of gRPC Stream Performance](https://ably.com/blog/grpc-stream-performance)
* [gRPC On HTTP/2: Engineering a robust, high performance protocol](https://www.cncf.io/blog/2018/08/31/grpc-on-http-2-engineering-a-robust-high-performance-protocol/)
