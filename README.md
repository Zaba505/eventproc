# eventproc

A proof-of-concept for implementing a RPC style gateway for processing
actions leveraging bi-directional gRPC streams.

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
+---------+         +-+---v---+         +-------------+
|         +-------->|         +-------->|             |
| Client  |         | Gateway |         | Processor 2 |
|         |<--------+         |<--------+             |
+---------+         +--^---+--+         +-------------+
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

**Gateway:** implements an API for clients to simply send a request to and then
receive a response. The response could be nothing or it could contain content
from one of the backend Processors, which the Gateway would relay to the
corresponding client. This provides a uniform for both clients as well as
processors.

**Processor:** a gRPC service which actions and there responses are streamed
to and from.

## Implementation

Located in the `action` folder are gRPC service definitions for a Gateway and
Processor along with a rough implementation of a Gateway.

Communication between the `Gateway` and the backend `Processors` is done
over a bi-directional gRPC stream. The `Gateway` uses a API defined enum for
its basis of mapping events to their respective `Processors`. By specifying
an action type enum, supporting new action types and their respective processors
becomes trivial since all a gateway needs is to map action type(s) to processor(s).

`gateway/main.go` runs a Gateway service exposing it through a HTTP REST-ish API
and a gRPC API.

`echo/main.go` runs a Processor "backend" that simply echoes back any action
content streamed to it. It only exposes the streaming gRPC API, per the Processor
service definition.

## Research resources

* [The Mysterious Gotcha of gRPC Stream Performance](https://ably.com/blog/grpc-stream-performance)
* [gRPC On HTTP/2: Engineering a robust, high performance protocol](https://www.cncf.io/blog/2018/08/31/grpc-on-http-2-engineering-a-robust-high-performance-protocol/)
