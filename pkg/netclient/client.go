package netclient

import (
	"encoding/binary"
	"fmt"
	"io"
	"ledctl3/pkg/codec"
	"net"
	"sync"
	"time"
)

type Client[E any] struct {
	mux     sync.Mutex
	codec   codec.Codec[E]
	handler func(net.Addr, E)
	conn    net.Conn
}

func New[E any](codec codec.Codec[E], eventHandler func(net.Addr, E)) *Client[E] {
	c := &Client[E]{
		codec:   codec,
		handler: eventHandler,
	}

	return c
}

func (c *Client[E]) Connect(addr net.Addr) {
	connected := make(chan bool)

	go c.connect(addr, connected)

	//<-connected
}

func (c *Client[E]) connect(addr net.Addr, connected chan bool) {
	var first bool
	go func() {
		for {
			netConn, err := net.DialTimeout(addr.Network(), addr.String(), 1*time.Second)
			if err != nil {
				fmt.Println("error during dial: ", err)
				continue
			}

			c.mux.Lock()
			c.conn = netConn
			c.mux.Unlock()

			if !first {
				first = true
				//connected <- true
			}

			c.processEvents(addr, netConn)

			c.mux.Lock()
			c.conn = nil
			c.mux.Unlock()
		}
	}()
}

func (c *Client[E]) processEvents(addr net.Addr, conn net.Conn) {
	defer fmt.Println("HANDLE CONN DONE")

	var foundLength bool
	var msglen uint32
	sizeBuf := make([]byte, 4)

	for {
		if !foundLength {
			n, err := conn.Read(sizeBuf)
			if err != nil {
				_ = conn.Close()
				fmt.Println("error during read: ", err)
				return
			}

			if n != 4 {
				fmt.Println("invalid header")
				continue
			}

			msglen = binary.LittleEndian.Uint32(sizeBuf)
			if msglen > 0 {
				foundLength = true
			} else {
				fmt.Println("ACK")
			}
		} else {
			readBuf := make([]byte, msglen)
			n, err := conn.Read(readBuf)
			if err != nil {
				_ = conn.Close()
				fmt.Println("error during read: ", err)
				return
			}

			if n != int(msglen) {
				fmt.Println("invalid message")
				continue
			}

			var e E
			err = c.codec.UnmarshalEvent(readBuf, &e)
			if err != nil {
				continue
			}

			fmt.Println("received msg")

			c.handler(addr, e)

			foundLength = false
		}
	}
}

func (c *Client[E]) Send(e E) error {
	c.mux.Lock()
	conn := c.conn
	c.mux.Unlock()

	if conn == nil {
		return io.ErrClosedPipe
	}

	buf, err := c.codec.MarshalEvent(e)
	if err != nil {
		return err
	}

	length := make([]byte, 4)
	binary.LittleEndian.PutUint32(length, uint32(len(buf)))
	buf = append(length, buf...)

	n, err := conn.Write(buf)
	if err != nil {
		_ = conn.Close()
		return err
	}

	if n != len(buf) {
		return io.ErrShortWrite
	}

	return nil
}
