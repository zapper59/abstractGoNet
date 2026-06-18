package abstractGoNet

import (
    "testing"
)

func TestRealNet(t *testing.T) {
    listener, err := RealNet().Listen("tcp", ":")
    if err != nil {
        t.Error(err)
    }
    if err = listener.Close(); err != nil {
        t.Error(err)
    }
}

func TestClosedListener(t *testing.T) {
    wan := NewVirtualWan()
    host := wan.NewVirtualHost("a")
    listener, err := host.Listen("tcp", ":123")
    if err != nil {
        t.Error(err)
    }
    
    if err := listener.Close(); err != nil {
        t.Error(err)
    }

    _, err = listener.Accept()
    if err != ListenerClosedErr {
        t.Error(err)
    }
}
