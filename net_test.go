package abstractGoNet

import (
    "errors"
    "io"
    "net"
    "testing"
)

func TestRealNetListen(t *testing.T) {
    listener, err := RealNet().Listen("tcp", ":")
    if err != nil {
        t.Error(err)
    }
    if err = listener.Close(); err != nil {
        t.Error(err)
    }
}

func TestRealNetDial(t *testing.T) {
    _, err := RealNet().Dial("tcp", "localhost:70000")
    if err == nil {
        t.Errorf("Connected to invalid port")
    }
}

func TestListenerClosed(t *testing.T) {
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

func TestListenerConflict(t *testing.T) {
    wan := NewVirtualWan()
    host := wan.NewVirtualHost("a")
    _, err := host.Listen("tcp", ":123")
    if err != nil {
        t.Error(err)
    }

    _, err = host.Listen("unix", ":123")
    if err != ListenerConflictErr {
        t.Error(err)
    }
}

func dialAsync(
    t *testing.T,
    host Net,
    network, address string,
    response chan net.Conn,
) {
    c, err := host.Dial(network, address)
    if err != nil {
        t.Error(err)
    }
    response <- c
}

const testData = "a piece of data with some length"

func roundTripData(a net.Conn, b net.Conn) error {
    size, err := io.WriteString(a, testData)
    if err != nil {
        return err
    }

    _, err = io.CopyN(b, b, int64(size))
    if err != nil {
        return err
    }

    buff := make([]byte, size)
    _, err = io.ReadFull(a, buff)
    if err != nil {
        return err
    }

    if string(buff) != testData {
        return errors.New(string(buff))
    }
    return nil
}

func TestDialBasic(t *testing.T) {
    wan := NewVirtualWan()
    host1 := wan.NewVirtualHost("a")
    listener, err := host1.Listen("tcp", ":123")
    if err != nil {
        t.Error(err)
    }

    host2 := wan.NewVirtualHost("b")
    connChan := make(chan net.Conn)
    go dialAsync(t, host2, "tcp", "a:123", connChan)

    conn1, err := listener.Accept()
    if err != nil {
        t.Error(err)
    }
    conn2 := <-connChan

    la1 := conn1.LocalAddr().String()
    if la1 != "a:123" {
        t.Errorf("Wrong addr %s", la1)
    }
    ra1 := conn1.RemoteAddr().String()
    if ra1 != "b" {
        t.Errorf("Wrong addr %s", ra1)
    }
    la2 := conn2.LocalAddr().String()
    if la2 != "b" {
        t.Errorf("Wrong addr %s", la2)
    }
    ra2 := conn2.RemoteAddr().String()
    if ra2 != "a:123" {
        t.Errorf("Wrong addr %s", ra2)
    }

    err = roundTripData(conn1, conn2)
    if err != nil {
        t.Error(err)
    }
}

func TestDialWrongPort(t *testing.T) {
    wan := NewVirtualWan()
    host1 := wan.NewVirtualHost("a")
    _, err := host1.Listen("tcp", ":123")
    if err != nil {
        t.Error(err)
    }

    host2 := wan.NewVirtualHost("b")
    _, err = host2.Dial("tcp", "a:321")
    if err != ListenerNotFoundErr {
        t.Error(err)
    }
}

func TestDialWrongHost(t *testing.T) {
    wan := NewVirtualWan()
    host1 := wan.NewVirtualHost("a")
    _, err := host1.Listen("tcp", ":123")
    if err != nil {
        t.Error(err)
    }

    host2 := wan.NewVirtualHost("b")
    _, err = host2.Dial("tcp", "c:123")
    if err != HostNotFoundErr {
        t.Error(err)
    }
}
