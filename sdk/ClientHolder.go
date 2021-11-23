package sdk

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

/*
ClientHolder
由于websocket的close()方法是可重入的，会多次调用，但是关闭channel的close()是不可重入的，因此需要isClosed进行判断。
isClosed可能会发生资源竞争，因此需要互斥锁。
关闭websocket链接后，也要自动关闭输入和输出流，因此通过signalCloseLoopChan实现
*/
type ClientHolder struct {
	UserName            string          //用户名称
	conn                *websocket.Conn //链接对象
	input               chan []byte     // 输入渠道
	output              chan []byte     //输出渠道
	signalCloseLoopChan chan byte       //关闭信号
	isClosed            bool            //是否调用过close()方法
	lock                sync.Mutex      //锁
}

// Create 创建连接对象
func Create(c *websocket.Conn) (conn *ClientHolder, err error) {
	conn = &ClientHolder{
		conn:                c,
		input:               make(chan []byte, 1000),
		output:              make(chan []byte, 1000),
		signalCloseLoopChan: make(chan byte, 1),
		isClosed:            false,
	}
	return
}

func (c *ClientHolder) Close() {
	_ = c.conn.Close()
	c.lock.Lock()
	defer func() {
		c.lock.Unlock()
	}()
	if !c.isClosed {
		close(c.signalCloseLoopChan)
		close(c.input)
		close(c.output)
		c.isClosed = true
	}
	return
}

func (c *ClientHolder) ReadLoop() {
	for {
		//读取消息
		_, bytes, err := c.conn.ReadMessage()
		if err != nil {
			c.signalCloseLoopChan <- 1
			break
		}

		message := NewMessage(bytes)
		to := message.To
		msg := fmt.Sprintf("%v", message.Msg)

		client := ServerRuntime.Clients[to]
		if client == nil {
			log.Println(to + " 不在线，发送失败")

			c.output <- []byte(to + " 不在线，发送失败")
			continue
		}
		client.output <- []byte(msg)
	}
}

func (c *ClientHolder) WriteLoop() {
	var data []byte
	for {
		select {
		case data = <-c.output: //写出消息渠道
			if len(data) > 0 {
				if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
					c.signalCloseLoopChan <- 1
				}
			}
		case <-c.signalCloseLoopChan: //关闭链接消息渠道
			ServerRuntime.LogoutEventChan <- []byte(c.UserName)
			c.Close()
		}

	}
}
