package main

import (
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/henrylee2cn/teleport/socket"

	"github.com/henrylee2cn/teleport/socket/example/pb"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:8000")
	if err != nil {
		log.Fatalf("[CLI] dial err: %v", err)
	}
	s := socket.GetSocket(conn)
	defer s.Close()

	var count sync.WaitGroup
	t := time.Now()
	loop := 100
	group := 1000
	var failNum uint32
	defer func() {
		cost := time.Since(t)
		times := time.Duration(loop * group)
		log.Printf("------------------- call times: %d ok: %d fail: %d | cost time: %v | QPS: %d -----------------", times, uint32(times)-failNum, failNum, cost, time.Second*times/cost)
		select {}
	}()

	for j := 0; j < loop; j++ {
		count.Add(group)
		for i := 0; i < group; i++ {
			go func() {
				var packet = socket.GetPacket(nil)
				defer func() {
					socket.PutPacket(packet)
					count.Done()
				}()
				// write request
				packet.Reset(nil)
				packet.HeaderCodec = "protobuf"
				packet.BodyCodec = "protobuf"
				packet.Header.Seq = 1
				packet.Header.Uri = "/a/b"
				packet.Header.Gzip = 5
				packet.Body = &pb.PbTest{A: 123, B: "pbtest"}
				err = s.WritePacket(packet)
				if err != nil {
					atomic.AddUint32(&failNum, 1)
					log.Printf("[CLI] write request err: %v", err)
					return
				}

				// read response
				packet.Reset(func(_ *socket.Header) interface{} {
					return new(pb.PbTest)
				})
				err = s.ReadPacket(packet)
				if err != nil {
					atomic.AddUint32(&failNum, 1)
					log.Printf("[CLI] read response err: %v", err)
				}
			}()
		}
		count.Wait()
	}
}
