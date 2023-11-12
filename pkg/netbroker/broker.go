package netbroker

import (
	"encoding/binary"
	"errors"
	"fmt"
	"ledctl3/pkg/uuid"
	"net"
	"sync"
	"time"
)

var deadline = 5 * time.Second
var pingTimeout = 1 * time.Second

type BrokerIface[E any] interface {
	Receive(addr net.Addr, handler func(e E)) (dispose func())
	Send(addr net.Addr, e E) error
}

type address struct {
	network string
	addr    string
}

type connection struct {
	mux  sync.Mutex
	conn net.Conn
}

type Broker[E any] struct {
	mux     sync.Mutex
	ln      net.Listener
	out     map[address]*connection
	in      map[uuid.UUID]func(net.Addr, E)
	encoder func(E) ([]byte, error)
	decoder func([]byte) (E, error)
}

func (b *Broker[E]) AddServer(addr net.Addr) {
	id := address{
		network: addr.Network(),
		addr:    addr.String(),
	}

	c := &connection{
		conn: nil,
	}

	b.mux.Lock()
	b.out[id] = c
	b.mux.Unlock()

	//conn, err := net1.DialTimeout(addr.Network(), addr.String(), 1*time.Second)
	//
	//if err == nil {
	//	c.mux.Lock()
	//	c.connected = true
	//	c.conn = conn
	//	c.mux.Unlock()
	//}

	go func() {
		//b.pingpong(c.conn)
		//
		//c.mux.Lock()
		//c.conn = nil
		//c.connected = false
		//c.mux.Unlock()

		for {
			conn, err := net.DialTimeout(addr.Network(), addr.String(), deadline)
			if err != nil {
				fmt.Println("error during dial: ", err)
				continue
			}

			//err = conn.SetDeadline(time.Now().Add(deadline))
			//if err != nil {
			//	fmt.Println("error during set deadline: ", err)
			//	continue
			//}

			c.mux.Lock()
			c.conn = conn
			c.mux.Unlock()

			b.handleConnection(c.conn)

			c.mux.Lock()
			c.conn = nil
			c.mux.Unlock()
		}
	}()

	return
}

//func (b *Broker[E]) connect(addr address) {
//	b.mux.Lock()
//	if _, ok := b.out[addr]; !ok {
//		b.out[addr].connecting = true
//	}
//	b.mux.Unlock()
//
//	for {
//		conn, err := net1.DialTimeout(addr.network, addr.addr, 1*time.Second)
//		if err != nil {
//			time.Sleep(100 * time.Millisecond) // TODO exponential
//		}
//
//		b.out[addr] = conn
//		b.mux.Unlock()
//	}
//
//	return nil
//}

func (b *Broker[E]) Receive(handler func(addr net.Addr, e E)) (dispose func()) {
	if handler == nil {
		return func() {}
	}

	id := uuid.New()

	b.mux.Lock()
	b.in[id] = handler
	b.mux.Unlock()

	return func() {
		b.mux.Lock()
		delete(b.in, id)
		b.mux.Unlock()
	}
}

func (b *Broker[E]) Send(netAddr net.Addr, e E) error {
	buf, err := b.encoder(e)
	if err != nil {
		return err
	}

	length := make([]byte, 4)
	binary.LittleEndian.PutUint32(length, uint32(len(buf)))
	buf = append(length, buf...)

	id := address{
		network: netAddr.Network(),
		addr:    netAddr.String(),
	}

	b.mux.Lock()
	c, ok := b.out[id]
	b.mux.Unlock()

	c.mux.Lock()
	defer c.mux.Unlock()
	if !ok || c.conn == nil {
		return errors.New("no connection to address")
	}

	n, err := c.conn.Write(buf)
	if n != len(buf) {
		return errors.New("failed to write all bytes")
	} else if err != nil {
		return err
	}

	return nil
}

func New[E any](encoder func(E) ([]byte, error), decoder func([]byte) (E, error)) *Broker[E] {
	return &Broker[E]{
		out:     map[address]*connection{},
		in:      map[uuid.UUID]func(net.Addr, E){},
		encoder: encoder,
		decoder: decoder,
	}
}

func (b *Broker[E]) Start(port string) {
	ln, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println(err)
		return
	}

	b.ln = ln

	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				fmt.Println(err)
				return
			}

			go b.ping(c)
			go b.handleConnection(c)
		}
	}()
}

func (b *Broker[E]) Stop() {
	_ = b.ln.Close()
}

func (b *Broker[E]) ping(conn net.Conn) {
	defer fmt.Println("PING INIT DONE")
	ping := make([]byte, 4)

	for {
		time.Sleep(pingTimeout)

		err := conn.SetWriteDeadline(time.Now().Add(deadline))
		if err != nil {
			fmt.Println("error during ping: ", err)
			return
		}

		fmt.Print("PING... ")
		n, err := conn.Write(ping)
		if err != nil || n != 4 {
			fmt.Println("error during ping: ", err)

			// cancel all blocked requests
			err = conn.SetReadDeadline(time.Now())
			if err != nil {
				fmt.Println("error during ping: ", err)
				return
			}

			return
		}

		fmt.Println("PONG!")

		//err = conn.SetWriteDeadline(time.Now().Add(deadline))
		//if err != nil {
		//	fmt.Println("error during ping: ", err)
		//	return
		//}
	}
}

func (b *Broker[E]) handleConnection(conn net.Conn) {
	defer fmt.Println("HANDLE CONN DONE")
	addr := conn.LocalAddr()

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

			e, err := b.decoder(readBuf)
			if err != nil {
				continue
			}

			fmt.Println("received msg")

			b.mux.Lock()
			for _, h := range b.in {
				go h(addr, e)
			}
			b.mux.Unlock()

			foundLength = false
		}
	}
}
