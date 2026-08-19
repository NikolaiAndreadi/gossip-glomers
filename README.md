# Gossip Glomers

Go solutions to [Fly.io's Gossip Glomers](https://fly.io/dist-sys/) distributed systems challenges, run against [Maelstrom](https://github.com/jepsen-io/maelstrom).

## Requirements

- Go 1.26+
- JDK, gnuplot, graphviz (Maelstrom's dependencies)

On macOS:

```sh
make setup
```

Maelstrom itself is downloaded automatically on the first test run.

## Usage

Build and run all solutions:

```sh
make test-all
```

Run a single challenge:

```sh
make test-echo
```

Run every workload in sequence (takes a few minutes; failures are collected and reported at the end):

```sh
make test-all
```

Browse the results of the last run in Maelstrom's web UI:

```sh
make serve
```

## Challenges

Some challenges share an implementation: 3c passes with the multi-node broadcast solution, 
3d and 3e with the efficient one, and 6c with the 6b solution. Challenge 4 has two implementations:
- a stateless one built on Maelstrom's sequentially consistent key/value store, which is what the challenge intends;
- a CRDT-based one, just because I wanted to do it :-)

Challenge 6b additionally has `test-2-distributed-read-uncommitted-transactions`, which runs the same workload under 
network partitions with total availability required.

## License

MIT — see [LICENSE](LICENSE).
