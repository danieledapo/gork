# Gork
ZMachine v3 implemented in Go just to play Zork and learn Go :smile:.
It's far from being complete and useful, but the core features are already
implemented. It should be just a matter of adding the missing instructions.


### How to Install
```bash
$ go get github.com/d-dorazio/gork
$ go install github.com/d-dorazio/gork/cmd/gork
$ go install github.com/d-dorazio/gork/cmd/gork-ztools
```

### Usage
```bash
$ gork zork1.z5
```

Start SSH server with
```
$ gork -address 127.0.0.1:4273 -identity ~/.ssh/id_rsa zork1.z5
```

### Resources
- [Standard](http://inform-fiction.org/zmachine/standards/index.html)
- [ZTools](http://inform-fiction.org/zmachine/ztools.html)
