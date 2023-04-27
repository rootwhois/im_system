package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	user := &User{
		Name:   conn.RemoteAddr().String(),
		Addr:   conn.RemoteAddr().String(),
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	// 启动消息监听
	go user.Listen()
	return user
}

func (u *User) Online() {
	// 用户上线，将用户加入到在线表中
	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.mapLock.Unlock()
	// 广播当前用户上线消息
	u.server.BroadCast(u, "已上线")
}

func (u *User) Offline() {
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.mapLock.Unlock()
	u.server.BroadCast(u, "已下线")
}

func (u *User) DoMessage(msg string) {
	if strings.HasPrefix(msg, "/online") {
		// 查询当前在线用户
		var singleMsg string
		u.server.mapLock.Lock()
		for _, user := range u.server.OnlineMap {
			singleMsg += "[" + user.Addr + "|" + user.Name + "]:" + "在线\n"
		}
		u.server.mapLock.Unlock()
		u.Send(singleMsg)
	} else if strings.HasPrefix(msg, "rename|") {
		// /rename newname
		s := strings.Split(msg, "|")
		if len(s) <= 1 || strings.Trim(s[1], " ") == "" {
			u.Send("修改失败！")
			return
		}
		newName := strings.Trim(s[1], " ")
		if strings.Contains(newName, "|") {
			u.Send("含有敏感字符，修改失败！")
			return
		}
		// 判断name是否存在
		if _, ok := u.server.OnlineMap[newName]; ok {
			u.Send("当前用户名已被使用！")
		} else {
			u.server.mapLock.Lock()
			delete(u.server.OnlineMap, u.Name)
			u.server.OnlineMap[newName] = u
			u.server.mapLock.Unlock()
			oldName := u.Name
			u.Name = newName
			u.server.BroadCast(u, fmt.Sprintf("用户%s改名为%s", oldName, newName))
		}

	} else if strings.HasPrefix(msg, "to|") {
		s := strings.Split(msg, "|")
		if len(s) != 3 {
			u.Send("消息格式错误！参考格式: to|张三|你好啊")
			return
		}

		distName := s[1]
		content := s[2]
		// u.server.mapLock.Lock()
		if user, ok := u.server.OnlineMap[distName]; !ok {
			u.Send(fmt.Sprintf("找不到用户%s，发送失败！", distName))
		} else {
			user.Send(fmt.Sprintf("%s对你私聊，消息内容：%s", u.Name, content))
		}
		// u.server.mapLock.Unlock()
	} else {
		// 广播
		u.server.BroadCast(u, msg)
	}
}

// 监听chan的方法，一旦有消息直接发送给客户端
func (u *User) Listen() {
	for msg := range u.C {
		u.Send(msg)
	}
}

func (u *User) Send(msg string) {
	u.conn.Write([]byte(msg + "\n"))
}
