package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func NewClient(serverIp string, serverPort int) *Client {
	// 创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}

	// 连接server
	c, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net dial err", err)
		return nil
	}
	client.conn = c

	// 返回对象
	return client
}

func (c *Client) menu() bool {
	fmt.Println("1. 公聊模式")
	fmt.Println("2. 私聊模式")
	fmt.Println("3. 更新用户名")
	fmt.Println("0. 退出")

	var flag int
	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		c.flag = flag
		return true
	} else {
		fmt.Println(">>>请输入合法范围内的数字！")
		return false
	}
}

func (c *Client) DealResponse() {
	io.Copy(os.Stdout, c.conn)
}

func (c *Client) SendMsg(msg string) bool {
	_, err := c.conn.Write([]byte(msg))
	if err != nil {
		fmt.Println(">>>发送失败！")
		return false
	}
	return true
}

func (c *Client) Rename() bool {
	fmt.Print(">>>请输入用户名：")
	fmt.Scanln(&c.Name)
	msg := "rename|" + c.Name + "\n"
	return c.SendMsg(msg)
}

func (c *Client) PublicChat() {
	var msg string
	fmt.Print(">>>请输入内容，(exit退出)：")
	fmt.Scanln(&msg)
	for msg != "exit" {
		if len(msg) == 0 {
			fmt.Println("内容不可为空！")
		} else if !c.SendMsg(msg + "\n") {
			break
		}
		msg = ""
		fmt.Print(">>>请输入内容，(exit退出)：")
		fmt.Scanln(&msg)
	}
}

func (c *Client) ShowUsers() {
	msg := "/online\n"
	c.SendMsg(msg)
}

func (c *Client) PrivateChat() {
	c.ShowUsers()

	var distName string
	var content string
	fmt.Print(">>>请输入私聊对象的用户名，(exit退出)：")
	fmt.Scanln(&distName)

	for distName != "exit" {
		fmt.Print(">>>请输入消息内容，(exit退出)：")
		fmt.Scanln(&content)

		for content != "exit" {
			if len(content) == 0 {
				fmt.Println("内容不可为空！")
			} else if !c.SendMsg("to|" + distName + "|" + content + "\n") {
				break
			}
			content = ""
			fmt.Print(">>>请输入内容，(exit退出)：")
			fmt.Scanln(&content)
		}

		fmt.Print(">>>请输入私聊对象的用户名，(exit退出)：")
		fmt.Scanln(&distName)
	}

}

func (c *Client) Run() {
	for c.flag != 0 {
		for !c.menu() {
		}

		switch c.flag {
		case 1:
			c.PublicChat()
			break
		case 2:
			c.PrivateChat()
			break
		case 3:
			c.Rename()
			break
		}
	}
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "IP地址")
	flag.IntVar(&serverPort, "port", 8888, "端口号")
}

func main() {
	flag.Parse()
	c := NewClient(serverIp, serverPort)
	if c == nil {
		fmt.Println(">>>连接失败！")
		return
	}

	fmt.Println(">>>连接成功！")
	go c.DealResponse()

	c.Run()
}
