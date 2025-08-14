# GLOMERS

A series of distributed systems challenges that leverage Maelstrom to test distributed workloads and ensure systems are resilient and can survive all network delays and partitions.

The whole point of this is to ensure applications can tolerate network partition among distributed systems, basically CAP's theorem in practice.

## Prerequisites

-   [Maelstrom](https://github.com/jepsen-io/maelstrom) - A tool for simulating a toy distributed system that spins up multiple clients to interact with our servers.

## Usage

1. Test the unique ID generator by running the command below which spins up a `3-node cluster` and sends `10k requests per second` to the server while inducing network errors midway

```bash
make run-unique-ids
```

## Projects

-   `echo`: a simple echo server to demonstrate working with maelstrom in a distributed environment
-   `unique-ids`: a globally unique ID generator. The IDs are of length 24 and sortable by time while being unique.

## Considerations

Here are some decisions I took in the implementation of the projects listed.

1. Unique ID Generator

-   Timestamps are retrieved as unix epoch time, 32-bits, where they represented in big endian format
-   Node IDs are presumed to be unique on bootstrapping
-   A monotonically increasing atomic clock is used as a counter for each process of the application run or restarted
-   All length constraints not satisfied are padded with zeros to the left

## References

-   [Gossip Glomers](https://fly.io/blog/gossip-glomers)
-   [MongoDB ObjectID](https://www.mongodb.com/docs/manual/reference/method/objectid/)
