package netserver

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"

	"ledctl3/pkg/codec"
)

type Server[E any] struct {
	mux               sync.Mutex
	codec             codec.Codec[E]
	ln                net.Listener
	port              int
	messageHandler    func(string, E)
	requestHandler    func(string, E, func(E) error)
	conns             map[connId]net.Conn
	connectHandler    func(string)
	disconnectHandler func(string)
	state             int

	requests map[uuid.UUID]func(E)
}

type connId struct {
	netw string
	addr string
}

func New[E any](port int, codec codec.Codec[E]) *Server[E] {
	s := &Server[E]{
		port:     port,
		codec:    codec,
		conns:    map[connId]net.Conn{},
		requests: map[uuid.UUID]func(E){},
	}

	return s
}

func (s *Server[E]) Connect(addr net.Addr) (net.Conn, error) {
	c, err := net.DialTimeout(addr.Network(), addr.String(), 1*time.Second)
	if err != nil {
		//fmt.Println("error during dial: ", err)
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

			go s.HandleConnection(c.RemoteAddr(), c)
		}
	}()

	return nil
}

func (s *Server[E]) Stop() {
	_ = s.ln.Close()
}

const (
	StateType uint8 = iota
	StateReqId
	StateResId
	StateLen
	StateData
)

const (
	TypDefault uint8 = iota
	TypReq
	TypRes
)

type connState struct {
	state uint8

	typbuf [1]byte
	idbuf  uuid.UUID
	lenbuf [4]byte
	msgbuf []byte

	typ   uint8
	reqId uuid.UUID
	len   uint32

	//reqs map[[16]byte]
}

func newconnState() connState {
	return connState{
		state: StateType,
		//idbuf:  [16]byte{},
		//lenbuf: [4]byte{},
		//typbuf: [1]byte{},
		msgbuf: make([]byte, 0),

		typ: 0,
	}
}

func (s *Server[E]) HandleConnection(addr net.Addr, conn net.Conn) error {
	//fmt.Println("HandleConnection")

	if s.connectHandler != nil {
		s.connectHandler(addr.String())
	}

	defer s.cleanupConn(addr)

	st := newconnState()

	//var foundLength bool
	//var msglen uint32
	//sizeBuf := make([]byte, 4)
	var tmplen int
	//var tmpbuf []byte

	for {
		//fmt.Println("loop")

		switch st.state {

		case StateType:
			n, err := conn.Read(st.typbuf[:])
			if err != nil {
				_ = conn.Close()
				return err
			}

			if n != 1 {
				continue
			}

			//fmt.Printf("reading: t %d\n", st.typ)

			//fmt.Printf("state %d -> %d\n", st.state, st.typbuf[0])

			switch st.typbuf[0] {
			case TypDefault:
				st.state = StateLen
			case TypReq:
				st.state = StateReqId
			case TypRes:
				st.state = StateResId
			}

			st.typ = st.typbuf[0]

		case StateReqId:

			n, err := conn.Read(st.idbuf[tmplen:])
			if err != nil {
				_ = conn.Close()
				return err
			}

			tmplen += n

			if tmplen != 16 {
				continue
			}

			st.reqId = st.idbuf
			st.state = StateLen
			tmplen = 0

			//fmt.Printf("reading: req id %d\n", st.reqId)

		case StateResId:

			n, err := conn.Read(st.idbuf[tmplen:])
			if err != nil {
				_ = conn.Close()
				return err
			}

			tmplen += n

			if tmplen != 16 {
				continue
			}

			st.reqId = st.idbuf
			st.state = StateLen
			tmplen = 0

			//fmt.Printf("reading: res id %d\n", st.reqId)

		case StateLen:

			n, err := conn.Read(st.lenbuf[tmplen:])
			if err != nil {
				_ = conn.Close()
				return err
			}

			tmplen += n

			if tmplen != 4 {
				continue
			}

			st.len = binary.LittleEndian.Uint32(st.lenbuf[:])
			st.state = StateData
			tmplen = 0
			if int(st.len) > cap(st.msgbuf) {
				st.msgbuf = make([]byte, st.len)
			}

			//fmt.Printf("reading: l %d\n", st.len)

		case StateData:

			n, err := conn.Read(st.msgbuf[tmplen:st.len])
			if err != nil {
				_ = conn.Close()
				return err
			}

			tmplen += n

			if tmplen != int(st.len) {
				continue
			}

			//fmt.Printf("reading: m %d\n", st.len)

			if st.typ == TypDefault {
				if s.messageHandler != nil {
					e, err := s.codec.UnmarshalBinary(st.msgbuf[:st.len])
					if err != nil {
						fmt.Println("error during unmarshal: ", err)
						continue
					}

					s.messageHandler(addr.String(), e)
				}
			} else if st.typ == TypReq {
				if s.requestHandler != nil {
					e, err := s.codec.UnmarshalBinary(st.msgbuf[:st.len])
					if err != nil {
						fmt.Println("error during unmarshal: ", err)
						continue
					}

					go s.requestHandler(addr.String(), e, func(e E) error {
						return s.writeResponse(addr.String(), e, st.reqId)
					})
				}
			} else if st.typ == TypRes {
				if s.requests[st.reqId] != nil {
					e, err := s.codec.UnmarshalBinary(st.msgbuf[:st.len])
					if err != nil {
						fmt.Println("error during unmarshal: ", err)
						continue
					}

					go s.requests[st.reqId](e)
					delete(s.requests, st.reqId)
				}
			}

			st.state = StateType
			tmplen = 0

		default:
			st.state = StateType
		}
	}
}

func (s *Server[E]) cleanupConn(addr net.Addr) {
	//fmt.Println("cleanupConn")

	id := connId{
		netw: addr.Network(),
		addr: addr.String(),
	}

	s.mux.Lock()
	s.conns[id] = nil
	s.mux.Unlock()

	if s.disconnectHandler != nil {
		s.disconnectHandler(addr.String())
	}
}

func (s *Server[E]) Request(addr string, e E) (E, error) {
	//fmt.Println("Request")

	id := connId{
		netw: "tcp",
		addr: addr,
	}

	s.mux.Lock()
	conn, ok := s.conns[id]
	s.mux.Unlock()

	var zero E

	if !ok {
		fmt.Println("no connection for addr", addr)
		return zero, io.ErrClosedPipe
	}

	if conn == nil {
		fmt.Println("conn nil for addr", addr)
		return zero, io.ErrClosedPipe
	}

	evt, err := s.codec.MarshalBinary(e)
	if err != nil {
		fmt.Println("error during marshal: ", err)
		return zero, err
	}

	reqId := uuid.New()

	b := make([]byte, 0, 1+16+4+len(evt))
	b = append(b, TypReq)
	b = append(b, reqId[:]...)
	length := make([]byte, 4)
	binary.LittleEndian.PutUint32(length, uint32(len(evt)))
	b = append(b, length...)
	//fmt.Printf("writing: t %d id %d l %d m %d\n", length[0], length[1], binary.LittleEndian.Uint32(header[2:]), len(evt))
	b = append(b, evt...)

	res := make(chan E, 1)
	s.requests[reqId] = func(e E) {
		res <- e
		close(res)
	}

	n, err := conn.Write(b)
	if err != nil {
		delete(s.requests, reqId)
		close(res)

		fmt.Println("error during write: ", err)
		_ = conn.Close()
		return zero, err
	}

	//fmt.Println("wrote!")

	if n != len(b) {
		delete(s.requests, reqId)
		close(res)

		fmt.Println("short write")
		return zero, io.ErrShortWrite
	}

	fmt.Println("waiting for response for request id", reqId)
	r := <-res
	fmt.Println("got response for request id", reqId)

	return r, nil
}

func (s *Server[E]) Write(addr string, e E) error {
	//fmt.Println("Write")

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

	buf, err := s.codec.MarshalBinary(e)
	if err != nil {
		fmt.Println("error during marshal: ", err)
		return err
	}

	typlen := [5]byte{}
	typlen[0] = TypDefault
	binary.LittleEndian.PutUint32(typlen[1:], uint32(len(buf)))
	//fmt.Printf("writing: t %d l %d m %d\n", typlen[0], binary.LittleEndian.Uint32(typlen[1:]), len(buf))
	buf = append(typlen[:], buf...)

	n, err := conn.Write(buf)
	if err != nil {
		fmt.Println("error during write: ", err)
		_ = conn.Close()
		return err
	}

	//fmt.Println("wrote!")

	if n != len(buf) {
		fmt.Println("short write")
		return io.ErrShortWrite
	}

	return nil
}

func (s *Server[E]) writeResponse(addr string, e E, resId uuid.UUID) error {
	//fmt.Println("writeResponse")

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

	buf, err := s.codec.MarshalBinary(e)
	if err != nil {
		fmt.Println("error during marshal: ", err)
		return err
	}

	b := make([]byte, 0, 1+16+4+len(buf))
	b = append(b, TypRes)
	b = append(b, resId[:]...)
	length := make([]byte, 4)
	binary.LittleEndian.PutUint32(length, uint32(len(buf)))
	b = append(b, length...)
	//fmt.Printf("writing: t %d id %d l %d m %d\n", length[0], length[1], binary.LittleEndian.Uint32(header[2:]), len(buf))
	b = append(b, buf...)

	n, err := conn.Write(b)
	if err != nil {
		fmt.Println("error during write: ", err)
		_ = conn.Close()
		return err
	}

	//fmt.Println("wrote!")

	if n != len(b) {
		fmt.Println("short write")
		return io.ErrShortWrite
	}

	return nil
}

func (s *Server[E]) SetRequestHandler(h func(addr string, e E, respond func(E) error)) {
	s.requestHandler = h
}

func (s *Server[E]) SetMessageHandler(h func(addr string, e E)) {
	s.messageHandler = h
}

func (s *Server[E]) SetConnectHandler(h func(addr string)) {
	s.connectHandler = h
}

func (s *Server[E]) SetDisconnectHandler(h func(addr string)) {
	s.disconnectHandler = h
}
