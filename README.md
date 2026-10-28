# goCache

A Redis-compatible, in-memory key-value store written from scratch in Go.

goCache is a learning project that rebuilds the core ideas behind Redis: speaking the RESP wire protocol, handling many clients with `epoll` instead of one goroutine per connection, sharding the keyspace across worker goroutines, expiring keys, evicting with approximated LRU, and implementing the data structures that power Redis commands (hash tables, sets, a B+ tree-backed sorted set, Count-Min Sketch and Bloom filter) by hand.

Because it speaks RESP, you can talk to it with the standard `redis-cli`.

## Highlights

- **RESP protocol** – hand-written encoder/decoder for simple strings, bulk strings, integers, errors and arrays.
- **I/O multiplexing with epoll** – connections are monitored with raw `epoll_create1` / `epoll_ctl` / `epoll_wait` syscalls rather than Go's netpoller-per-goroutine model.
- **Shared-nothing multi-threading** – keys are hashed (MurmurHash3) to a fixed worker; each worker owns its own store, so commands run without locks.
- **Key expiration** – per-key TTL with lazy (on-access) deletion.
- **Approximated LRU eviction** – sampled eviction pool, modelled on Redis's `allkeys-lru`.
- **Hand-built data structures** – B+ tree for sorted sets, Count-Min Sketch, Bloom filter.
- **Graceful shutdown** on `SIGINT` / `SIGTERM`.

## Architecture

```
                 ┌──────────────┐
  redis-cli ───▶ │   Listener   │  net.Listen(":3000"), Accept()
                 └──────┬───────┘
                        │ round-robin
          ┌─────────────┴─────────────┐
          ▼                           ▼
   ┌─────────────┐             ┌─────────────┐
   │ I/O Handler │   ...       │ I/O Handler │   each owns an epoll instance,
   │   (epoll)   │             │   (epoll)   │   reads + parses RESP commands
   └──────┬──────┘             └──────┬──────┘
          │   murmur3(key) % numWorkers   │
          └─────────────┬─────────────────┘
          ┌─────────────┴─────────────┐
          ▼                           ▼
   ┌─────────────┐             ┌─────────────┐
   │  Worker 0   │   ...       │  Worker N   │   single goroutine per shard,
   │ dict/cms/bf │             │ dict/cms/bf │   executes command, replies via channel
   └─────────────┘             └─────────────┘
```

1. **Listener** (`internal/server/server.go`) accepts TCP connections on `:3000` and hands each one to an I/O handler in round-robin order.
2. **I/O handlers** (`internal/server/io_handler.go`) register the connection's file descriptor with their own `epoll` instance. When a socket becomes readable, the handler reads the request, parses it with the RESP decoder, and wraps it in a `Task` with a reply channel.
3. **Dispatch** hashes the command's first argument (the key) with MurmurHash3 and sends the task to `workers[hash % numWorkers]`. Every command for a given key therefore always lands on the same worker.
4. **Workers** (`internal/core/worker.go`) each run a single goroutine that drains a buffered task channel, looks the command up in a registry, executes it against the worker's private stores, and sends the RESP-encoded response back. Since a worker's data is touched by only that goroutine, no mutexes are needed on the data path.

On startup the server uses `runtime.NumCPU() / 2` workers and the same number of I/O handlers.

## Supported commands

| Group            | Commands                                                                                                                                             |
| ---------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| Server           | `PING [msg]`, `INFO`                                                                                                                                 |
| Strings / keys   | `SET key value [EX seconds]`, `GET key`, `DEL key [key ...]`, `EXISTS key [key ...]`, `EXPIRE key seconds`, `TTL key`                                |
| Sets             | `SADD key member [member ...]`, `SREM key member [member ...]`, `SISMEMBER key member`, `SMEMBERS key`                                               |
| Sorted sets      | `ZADD key score member [score member ...]`, `ZSCORE key member`, `ZRANK key member`                                                                  |
| Count-Min Sketch | `CMS.INITBYDIM key width depth`, `CMS.INITBYPROB key error probability`, `CMS.INCRBY key item incr [item incr ...]`, `CMS.QUERY key item [item ...]` |
| Bloom filter     | `BF.RESERVE key error_rate capacity`, `BF.ADD key item`, `BF.EXISTS key item`                                                                        |

Command names are case-insensitive. Using a command on a key of the wrong type returns Redis's `WRONGTYPE` error.

## Under the hood

### Storage and expiration

Each worker holds a `Dict` (`internal/data_structure/dict.go`) made of two maps: one from key to an `Obj` (value, type tag, last-access time) and one from key to its absolute expiry timestamp in milliseconds. Expiry is enforced **lazily**: whenever a key is read, it is checked and deleted if its deadline has passed. `TTL` follows Redis conventions (`-2` = key missing, `-1` = no expiry).

### Approximated LRU eviction

Before every write, the dict checks whether it has reached `MinKeyRatioForEviction` of `MaxKeyNumber`. If so, it samples `EpoolLruSampleSize` keys into an **eviction pool** sorted by last access time (capped at `EpoolMaxSize` entries), then deletes the oldest `EvictionRatio × MaxKeyNumber` keys. This mirrors how Redis approximates LRU without keeping a full linked list.

### Sorted set (B+ tree)

`ZADD`/`ZSCORE`/`ZRANK` are backed by a B+ tree (`bplus_tree.go`) of `(score, member)` items ordered by score and then by member, with linked leaves. The default degree is 4.

### Count-Min Sketch

A `depth × width` matrix of counters, with one MurmurHash3 seed per row. `INCRBY` increments one cell per row; `QUERY` returns the minimum across rows. `INITBYPROB` derives the dimensions from the desired error bound and probability (`width = ⌈2/ε⌉`, `depth = ⌈log(δ)/log(½)⌉`).

### Bloom filter

A bit array sized from the expected capacity and target false-positive rate, using double hashing on top of MurmurHash3. `BF.ADD` on a missing key creates a filter with defaults (capacity 100, error rate 1%).

## Project layout

```
cmd/main.go                         entry point, signal handling
internal/
  config/config.go                  port, limits, eviction settings
  constant/                         RESP literals, data-type tags, tunables
  server/
    server.go                       listener, worker pool, key-based dispatch
    io_handler.go                   epoll-driven connection handlers
  core/
    resp.go                         RESP encoder/decoder + command parser
    worker.go                       worker loop and command registry
    commands_*.go                   command implementations per data type
    io_multiplexing/                epoll wrapper behind an IOMultiplexer interface
  data_structure/
    dict.go                         key/value + expiry store
    eviction*.go                    eviction pool and LRU ordering
    simple_set.go, sorted_set.go, bplus_tree.go
    cms.go, bloom_filter.go
```

## Getting started

**Requirements:** Go 1.25+ and **Linux** (the I/O layer uses `epoll` syscalls directly). On macOS or Windows, run it in Docker or WSL.

```bash
git clone https://github.com/TrienThongLu/go-cache.git
cd go-cache
go run ./cmd
```

The server listens on port `3000`. Connect with `redis-cli`:

```bash
redis-cli -p 3000
```

```
127.0.0.1:3000> SET greeting hello EX 60
OK
127.0.0.1:3000> GET greeting
"hello"
127.0.0.1:3000> TTL greeting
(integer) 57
127.0.0.1:3000> ZADD leaderboard 100 alice 85 bob
(integer) 2
127.0.0.1:3000> ZRANK leaderboard alice
(integer) 1
127.0.0.1:3000> BF.ADD visitors user:42
OK
127.0.0.1:3000> BF.EXISTS visitors user:42
(integer) 1
127.0.0.1:3000> CMS.INITBYDIM clicks 2000 5
OK
127.0.0.1:3000> CMS.INCRBY clicks home 3 about 1
...
```

Run the tests:

```bash
go test ./...
```

## Configuration

Settings live in `internal/config/config.go`:

| Setting                  | Default       | Meaning                                                                            |
| ------------------------ | ------------- | ---------------------------------------------------------------------------------- |
| `Port`                   | `:3000`       | TCP listen address                                                                 |
| `MaxConnection`          | `20000`       | Max events returned per `epoll_wait`                                               |
| `MaxKeyNumber`           | `10`          | Key limit per worker before eviction (kept small to make eviction easy to observe) |
| `MinKeyRatioForEviction` | `0.9`         | Fill ratio that triggers eviction                                                  |
| `EvictionRatio`          | `0.1`         | Fraction of `MaxKeyNumber` removed per eviction                                    |
| `EvictionPolicy`         | `allkeys-lru` | Eviction policy                                                                    |
| `EpoolMaxSize`           | `16`          | Max candidates kept in the eviction pool                                           |
| `EpoolLruSampleSize`     | `5`           | Keys sampled per eviction round                                                    |

## Limitations and ideas for future work

This is an educational project, not a production cache. Known gaps:

- **Linux only** because of the direct `epoll` dependency; a `kqueue` backend behind the existing `IOMultiplexer` interface would add macOS support.
- **Lazy expiry only.** An active expiry cycle (sampling keys every 100 ms, as Redis does) is sketched in `internal/core/expire.go` but currently disabled.
- **Single-key routing.** Multi-key commands such as `DEL k1 k2` are routed by the first key, so other keys may live on a different worker.
- **Shared eviction pool.** `EPool` is a package-level variable used by all workers; it should be per worker to keep the lock-free design safe.
- **Fixed 512-byte read buffer** with no handling of partial or pipelined requests.
- **No persistence** (RDB/AOF) and no replication.
- CMS and Bloom filter keys live in separate maps, so they don't count toward `INFO`, `EXISTS`, `DEL`, TTL or eviction.

## Acknowledgements

Inspired by the design of [Redis](https://redis.io/) and its modules (RedisBloom for `BF.*` and `CMS.*`).
