# Go Load Balancing

## LB Algos

- [x] Round Robin
- [x] Weighted Round Robin
- [x] Random
- [ ] Least Connections

## Types

### Server

weak numConns, currently only wrks for server side, thus server side db(in memory)

```go
// helper global var
MaxConnsDB(in mem sync.Map)
```

- numConns (atomic)
- config
- context
- handlerFunc

### Service

- server
- type

### Instance

- id
- service
- state
- stateChan
- mutex

shuld be able to pass state chan to service

### Reverse Proxy

to be used before passing to a server()

- lb (loadbalancer)
- servers(slice of addresses)

### Load Balancer

- servers
- weights (optional)
