package main

import (
	"fmt"
	"github.com/notedit/rtmp-lib"
	"github.com/notedit/rtmp-lib/pubsub"
	"sync"
)

type Channel struct {
	que  *pubsub.Queue
}

var channels = map[string]*Channel{}

func main()  {
	l := &sync.RWMutex{}
	server := rtmp.NewServer(&rtmp.Config{ChunkSize: 1024, BufferSize: 1024})
	server.Addr = "0.0.0.0:1945"
	server.HandlePlay = func(conn *rtmp.Conn) {
		fmt.Println("play")
		l.RLock()
		ch := channels[conn.URL.Path]
		l.RUnlock()
		if ch != nil {
			cursor := ch.que.Latest()
			streams, err := cursor.Streams()
			if err != nil {
				fmt.Println(err)
				return
			}
			conn.WriteHeader(streams)
			for {
				packet, err := cursor.ReadPacket()
				if err != nil {
					fmt.Println(err)
					break
				}
				conn.WritePacket(packet)
			}
		}
	}
	server.HandlePublish = func(conn *rtmp.Conn) {
		l.Lock()
		ch := channels[conn.URL.Path]
		if ch == nil {
			ch = &Channel{}
			ch.que = pubsub.NewQueue()
			ch.que.SetMaxGopCount(1)
			channels[conn.URL.Path] = ch
		}
		l.Unlock()
		defer func() {
			fmt.Println("Connection close")
			if err := conn.Close(); err != nil {
				fmt.Println(err)
			}
		}()
		fmt.Println(conn.URL)
		fmt.Println(conn.NetConn().LocalAddr())
		fmt.Println(conn.NetConn().RemoteAddr())
		streams, err := conn.Streams()
		if err != nil {
			fmt.Println("Error 1")
			return
		}
		ch.que.WriteHeader(streams)
		for {
			pkt, err := conn.ReadPacket()
			if err != nil {
				fmt.Println("Error 2")
				break
			}
			ch.que.WritePacket(pkt)
		}
		l.Lock()
		delete(channels, conn.URL.Path)
		l.Unlock()
		fmt.Println("ssss")
	}
	err := server.ListenAndServe()
	fmt.Println(err)
}
