package main

import (
	"ledctl3/pkg/netbroker"
	"net"
)

func main() {
	br := netbroker.New[string](func(s string) ([]byte, error) {
		return []byte(s), nil
	}, func(b []byte) (string, error) {
		return string(b), nil
	})
	br.Start(":8080")

	ip := net.ParseIP("127.0.0.1")
	addr := &net.TCPAddr{
		IP:   ip,
		Port: 8080,
		Zone: "",
	}

	br.AddServer(addr)

	//time.Sleep(1 * time.Second)
	//
	//err := br.Write(&net1.TCPAddr{
	//	IP:   ip,
	//	Port: 8080,
	//	Zone: "",
	//}, "hello")
	//fmt.Print("send err: ", err, "\n")
	//
	//time.Sleep(1 * time.Second)
	//
	//br.Receive(func(addr net1.Addr, s string) {
	//	fmt.Println("RECEIVED MESSAGE from ", addr, ": ", s)
	//})
	//
	//time.Sleep(1 * time.Second)
	//
	//err = br.Write(&net1.TCPAddr{
	//	IP:   ip,
	//	Port: 8080,
	//	Zone: "",
	//}, "hello")
	//fmt.Print("send err: ", err, "\n")
	//
	//time.Sleep(1 * time.Second)
	//
	//dispose := br.Receive(func(addr net1.Addr, s string) {
	//	fmt.Println("RECEIVER 2!!!!! ", addr, ": ", s)
	//})
	//
	//time.Sleep(1 * time.Second)
	//
	//for i := 0; i < 10; i++ {
	//	go br.Write(&net1.TCPAddr{
	//		IP:   ip,
	//		Port: 8080,
	//		Zone: "",
	//	}, "hello")
	//}
	//
	//time.Sleep(1 * time.Second)
	//
	//dispose()
	//
	//time.Sleep(1 * time.Second)
	//
	//err = br.Write(&net1.TCPAddr{
	//	IP:   ip,
	//	Port: 8080,
	//	Zone: "",
	//}, "hello")
	//fmt.Print("send err: ", err, "\n")
	//conn, err := net1.Dial("tcp", ":8080")
	//if err != nil {
	//	println("ResolveTCPAddr failed:", err.Error())
	//	os.Exit(1)
	//}
	//
	//time.Sleep(1 * time.Second)
	//
	//b := []byte{5, 0, 0, 0} // LE
	//b = append(b, []byte("hello")...)
	//
	//_, err = conn.Write(b)
	//if err != nil {
	//	println("Write to server failed:", err.Error())
	//	os.Exit(1)
	//}
	//
	//time.Sleep(1 * time.Second)
	//
	//b = []byte{6, 0, 0, 0} // LE
	//b = append(b, []byte("world!")...)
	//
	//_, err = conn.Write(b)
	//if err != nil {
	//	println("Write to server failed:", err.Error())
	//	os.Exit(1)
	//}

	select {}
}
