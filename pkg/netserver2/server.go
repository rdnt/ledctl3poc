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
	mux               sync.Mutex
	codec             codec.Codec[E]
	ln                net.Listener
	port              int
	handler           func(string, E)
	conns             map[connId]net.Conn
	connectHandler    func(string)
	disconnectHandler func(string)
}

type connId struct {
	netw string
	addr string
}

func New[E any](port int, codec codec.Codec[E]) *Server[E] {
	s := &Server[E]{
		port:  port,
		codec: codec,
		conns: map[connId]net.Conn{},
	}

	return s
}

func (s *Server[E]) Connect(addr net.Addr) (net.Conn, error) {
	c, err := net.DialTimeout(addr.Network(), addr.String(), 1*time.Second)
	if err != nil {
		fmt.Println("error during dial: ", err)
		return nil, err
	}

	id := connId{
		netw: addr.Network(),
		addr: addr.String(),
	}

	s.mux.Lock()
	s.conns[id] = c
	s.mux.Unlock()

	return c, nil
}

func (s *Server[E]) Start() error {
	if s.port == -1 {
		return errors.New("server disabled")
	}

	ln, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", s.port))
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

	if s.disconnectHandler != nil {
		defer func() {
			s.disconnectHandler(addr.String())
		}()
	}

	if s.connectHandler != nil {
		s.connectHandler(addr.String())
	}

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
				fmt.Println("error during unmarshal: ", err)
				continue
			}

			//fmt.Println("received msg")

			if s.handler != nil {
				s.handler(addr.String(), e)
			}

			foundLength = false
		}
	}
}

func (s *Server[E]) Write(addr string, e E) error {
	id := connId{
		netw: "tcp",
		addr: addr,
	}

	s.mux.Lock()
	conn, ok := s.conns[id]
	s.mux.Unlock()

	if !ok {
		fmt.Println("no connection for addr", addr)
		return io.ErrClosedPipe
	}

	if conn == nil {
		fmt.Println("conn nil for addr", addr)
		return io.ErrClosedPipe
	}

	buf, err := s.codec.MarshalEvent(e)
	if err != nil {
		fmt.Println("error during marshal: ", err)
		return err
	}

	length := make([]byte, 4)
	binary.LittleEndian.PutUint32(length, uint32(len(buf)))
	buf = append(length, buf...)

	n, err := conn.Write(buf)
	if err != nil {
		fmt.Println("error during write: ", err)
		_ = conn.Close()
		return err
	}

	if n != len(buf) {
		fmt.Println("short write")
		return io.ErrShortWrite
	}

	return nil
}

func (s *Server[E]) SetMessageHandler(h func(addr string, e E)) {
	s.handler = h
}

func (s *Server[E]) SetConnectHandler(h func(addr string)) {
	s.connectHandler = h
}

func (s *Server[E]) SetDisconnectHandler(h func(addr string)) {
	s.disconnectHandler = h
}
