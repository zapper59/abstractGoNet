# abstractGoNet
The primary purpose of this library is to make client-server applications using raw golang connections testable.
This package is more flexible than an injection of net.Pipe would be because the underlying IO is buffered over two os.Pipe implementations.
This package is easy to inject because it matches the syntax of the Go std net library.
It also features Mutex locking around the virtual WAN object allowing simple testing of scenarios where listeners and connections may frequently flap.

## Documentation

See the public docs at [pkg.go.dev](https://pkg.go.dev/github.com/zapper59/abstractGoNet)

