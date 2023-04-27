package main

import (
	"fmt"
	"io"
	"net"
	"runtime"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	// 在线用户
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播的channel
	Message chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// 监听群发
func (s *Server) Listen() {
	for {
		msg := <-s.Message

		s.mapLock.Lock()
		for _, cli := range s.OnlineMap {
			cli.C <- msg
		}
		s.mapLock.Unlock()
	}
}

func (s *Server) BroadCast(u *User, msg string) {
	sendMsg := "[" + u.Addr + "|" + u.Name + "]" + ":" + msg
	s.Message <- sendMsg
}

func (s *Server) Handler(conn net.Conn) {
	user := NewUser(conn, s)
	user.Online()

	// 监听用户是否活跃
	isLive := make(chan bool)

	// 接受客户端发送的消息
	go func() {
		buff := make([]byte, 4096)
		for {
			n, err := conn.Read(buff)
			if err != nil && err != io.EOF {
				fmt.Println("conn read error", err)
				return
			}

			// 合法关闭
			if n == 0 {
				user.Offline()
				return
			}

			// 提取用户消息，去除'\n'
			msg := string(buff[:n-1])
			user.DoMessage(msg)

			isLive <- true
		}
	}()

	// 当前Handler阻塞
	for {
		select {
		case <-isLive:
		// 当前用户是活跃的，重置定时器
		case <-time.After(1800 * time.Second):
			// 超时踢下线
			user.Send("你被踢了")
			close(user.C)
			conn.Close()
			runtime.Goexit()
		}
	}
}

func (s *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("net listen error", err)
		return
	}
	// close listen socket
	defer listener.Close()
	go s.Listen()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept error", err)
			continue
		}

		// do handler
		go s.Handler(conn)
	}

}
