package abstractGoNet

import (
    "errors"
    "io"
    "net"
    "os"
    "sync"
    "time"
)

var ListenerClosedErr = errors.New("Listener Closed")

// An abstract template matching that of [net].
type Net interface {
    Listen(network, address string) (net.Listener, error)
}

type realNetImpl struct { }

// Get a [Net] implementation that forwards to [net].
func RealNet() Net {
    return &realNetImpl{ }
}

func (_ *realNetImpl) Listen(network, address string) (net.Listener, error) {
    return net.Listen(network, address)
}

// Thread-safe map of virtual hosts, each of which implements [Net].
// All functions on objects handed out by this type, as well as the
// [net.Listener] implementations they generate share a single Mutex to allow
// for concurrent host lookups. IO itself is synchronized by the underlying
// [os.File] handed out by [os.Pipe] for each virtual connection.
type VirtualWan struct {
    mutex sync.Mutex

    hosts map[string]virtualHost // Mutex
}

// Create a virtual Wide Area Network.
func NewVirtualWan() VirtualWan {
    return VirtualWan {
        // Mutex
        hosts: make(map[string]virtualHost),
    }
}

// Register a unique machine on the network where addr is an ip or hostname.
func (self *VirtualWan) NewVirtualHost(addr string) Net {
    self.mutex.Lock()
    defer self.mutex.Unlock()

    info := VirtualHostInfo {}

    ip := net.ParseIP(addr)
    if ip == nil {
        info.Names = append(info.Names, addr)
    } else {
        info.Addrs = append(info.Addrs, addr)
    }

    return self.NewVirtualHostWithInfo(info)
}

// Register a unique machine on the network.
func (self *VirtualWan) NewVirtualHostWithInfo(info VirtualHostInfo) Net {
    host := virtualHost { self, info, make(map[string]virtualListener) }

    for _, n := range info.Names {
        self.hosts[n] = host
    }
    for _, a := range info.Addrs {
        self.hosts[a] = host
    }

    return &host
}

// Collection of hostnames and ip's that a host can be reached at.
type VirtualHostInfo struct {
    // Hostname list.
    Names []string

    // IP Address list.
    Addrs []string
}

type virtualHost struct {
    wan *VirtualWan
    info VirtualHostInfo
    listenersByPort map[string]virtualListener
}

type virtualAddr struct {
    network string
    address string
}

func (self *virtualAddr) Network() string {
    return self.network
}

func (self *virtualAddr) String() string {
    return self.address
}

func (self *virtualHost) Listen(network, address string) (net.Listener, error) {
    self.wan.mutex.Lock()
    defer self.wan.mutex.Unlock()

    _, port, err := net.SplitHostPort(address)
    if err != nil {
        return nil, err
    }

    addr := virtualAddr { network, address }
    listener := virtualListener {
        self, port, addr, make(chan acceptPayload),
    }
    self.listenersByPort[port] = listener
    return &listener, nil
}

type acceptPayload struct {
    populated bool
    pipe os.File
    remoteAddr virtualAddr
}

type virtualListener struct {
    host *virtualHost
    port string
    addr virtualAddr
    acceptChan chan acceptPayload
}

func (self *virtualListener) Accept() (net.Conn, error) {
    payload := <-self.acceptChan
    if !payload.populated {
        return nil, ListenerClosedErr
    }

    conn := virtualConn {
        payload.pipe, self.addr, payload.remoteAddr, false,
    }
    return &conn, nil
}

func (self *virtualListener) Addr() net.Addr {
    return &self.addr
}

func (self *virtualListener) Close() error {
    self.host.wan.mutex.Lock()
    defer self.host.wan.mutex.Unlock()
    if _, exists := self.host.listenersByPort[self.port]; !exists {
        return errors.New("listener not found")
    }

    close(self.acceptChan)
    delete(self.host.listenersByPort, self.port)
    return nil
}

type virtualConn struct {
    pipe os.File
    localAddr virtualAddr
    remoteAddr virtualAddr
    closed bool
}

func (self *virtualConn) Read(b []byte) (int, error) {
    n, err := self.pipe.Read(b)

    if err != nil && err != io.EOF {
        err = &net.OpError{
            Op: "read",
            Net: self.localAddr.network,
            Source: &self.localAddr,
            Addr: &self.remoteAddr,
            Err: err,
        }
    }

    return n, err
}

func (self *virtualConn) Write(b []byte) (int, error) {
    n, err := self.pipe.Write(b)

    if err != nil && err != io.EOF {
        err = &net.OpError{
            Op: "write",
            Net: self.localAddr.network,
            Source: &self.localAddr,
            Addr: &self.remoteAddr,
            Err: err,
        }
    }

    return n, err
}

func (self *virtualConn) Close() error {
    return self.pipe.Close()
}

func (self *virtualConn) LocalAddr() net.Addr {
    return &self.localAddr
}

func (self *virtualConn) RemoteAddr() net.Addr {
    return &self.remoteAddr
}

func (self *virtualConn) SetDeadline(t time.Time) error {
    return nil
}

func (self *virtualConn) SetReadDeadline(t time.Time) error {
    return nil
}

func (self *virtualConn) SetWriteDeadline(t time.Time) error {
    return nil
}
