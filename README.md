# go-reuseport

This package enables listening and dialing from _the same_ TCP or UDP port. This means that the following sockopts are set:

```
SO_REUSEADDR
SO_REUSEPORT
```

- godoc: https://godoc.org/github.com/jbenet/go-reuseport

## Examples


```Go
// listen on the same port. oh yeah.
l1, _ := reuse.Listen("tcp", "127.0.0.1:1234")
l2, _ := reuse.Listen("tcp", "127.0.0.1:1234")
```

```Go
// dial from the same port. oh yeah.
l1, _ := reuse.Listen("tcp", "127.0.0.1:1234")
l2, _ := reuse.Listen("tcp", "127.0.0.1:1235")
c, _ := reuse.Dial("tcp", "127.0.0.1:1234", "127.0.0.1:1235")
```

**Note: cant dial self because tcp/ip stacks use 4-tuples to identify connections, and doing so would clash.**

## Tested

Tested on `darwin` and `linux`.
