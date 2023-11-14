package netserver2

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"ledctl3/pkg/codec"
)

type Server[E any] struct {
	mux       sync.Mutex
	codec     codec.Codec[E]
	ln        net.Listener
	port      int
	onConnect func(net.Addr)
	handler   func(net.Addr, E)
	conns     map[connId]net.Conn
}

type connId struct {
	netw string
	addr string
}

func New[E any](port int, codec codec.Codec[E], onConnect func(net.Addr), handler func(net.Addr, E)) *Server[E] {
	s := &Server[E]{
		port:      port,
		codec:     codec,
		conns:     map[connId]net.Conn{},
		onConnect: onConnect,
		handler:   handler,
	}

	return s
}

func (s *Server[E]) Connect(addr net.Addr) (conn net.Conn, dispose func()) {
	for {
		c, err := net.DialTimeout(addr.Network(), addr.String(), 1*time.Second)
		if err != nil {
			fmt.Println("error during dial: ", err)
			continue
		}

		id := connId{
			netw: addr.Network(),
			addr: addr.String(),
		}

		s.mux.Lock()
		s.conns[id] = c
		s.mux.Unlock()

		return c, func() {
			_ = c.Close()
		}
	}
}

func (s *Server[E]) Start() error {
	if s.port == -1 {
		return errors.New("server disabled")
	}

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return err
	}

	s.ln = ln

	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				fmt.Println(err)
				continue
			}

			id := connId{
				netw: c.RemoteAddr().Network(),
				addr: c.RemoteAddr().String(),
			}

			s.mux.Lock()
			s.conns[id] = c
			s.mux.Unlock()

			go s.ProcessEvents(c.RemoteAddr(), c)
		}
	}()

	return nil
}

func (s *Server[E]) Stop() {
	_ = s.ln.Close()
}

func (s *Server[E]) ProcessEvents(addr net.Addr, conn net.Conn) {
	//fmt.Println("PROCESSING EVENTS FROM", addr)
	//defer fmt.Println("HANDLE CONN DONE")

	defer func() {
		id := connId{
			netw: addr.Network(),
			addr: addr.String(),
		}

		s.mux.Lock()
		s.conns[id] = nil
		s.mux.Unlock()
	}()

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
			err = s.codec.UnmarshalEvent(readBuf, &e)
			if err != nil {
				continue
			}

			//fmt.Println("received msg")

			s.handler(addr, e)

			foundLength = false
		}
	}
}

func (s *Server[E]) Write(addr net.Addr, e E) error {
	id := connId{
		netw: addr.Network(),
		addr: addr.String(),
	}

	s.mux.Lock()
	conn, ok := s.conns[id]
	s.mux.Unlock()
	if !ok {
		return io.ErrClosedPipe
	}

	if conn == nil {
		return io.ErrClosedPipe
	}

	buf, err := s.codec.MarshalEvent(e)
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
