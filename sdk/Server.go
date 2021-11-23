package sdk

var ServerRuntime = NewServer()

type Server struct {
	Clients           map[string]*ClientHolder //客户端链接对象，key=用户昵称，value=客户端链接的对象
	RegisterEventChan chan *ClientHolder       //用户登入渠道，用于处理用户登入后的处理逻辑
	LogoutEventChan   chan []byte              //用户登出渠道，用于处理用户退出后的后置处理逻辑
	RegisterMsgChan   chan []byte              //通知登入消息渠道，用于发送给其他用户有谁登录了
	LogoutMsgChan     chan []byte              //退出登录消息渠道，用于发送消息给其他用户有谁退出了
}

func NewServer() (s *Server) {
	s = &Server{
		Clients:           make(map[string]*ClientHolder, 100),
		RegisterEventChan: make(chan *ClientHolder),
		RegisterMsgChan:   make(chan []byte, 10),
		LogoutEventChan:   make(chan []byte, 1),
		LogoutMsgChan:     make(chan []byte, 10),
	}
	return s
}

/*
GetOnlineNum 获取在线用户数
 */
func (s *Server) GetOnlineNum() (num int) {
	return len(s.Clients)
}

/*
publishAll 向所有在线用户发送消息
 */
func (s *Server) publishAll(msg []byte) {
	for _, holder := range s.Clients {
		holder.output <- msg
	}
}

/*
Run
启动服务渠道监听
 */
func (s *Server) Run() {
	go func() {
		for {
			select {
			case userName := <-s.LogoutEventChan: //登出事件渠道监听
				//删除用户记录
				delete(s.Clients, string(userName))
				//向其他用户发送下线通知
				s.LogoutMsgChan <- []byte((string(userName)) + "下线了")
			case mag := <-s.LogoutMsgChan: //登出消息渠道监听
				s.publishAll(mag)
			case client := <-s.RegisterEventChan: //收到客户端登录的事件了，注册到map集合中
				s.ReceiverRegisterEvent(client)
			case msg := <-s.RegisterMsgChan: //收到登入消息，发送给其他客户端
				s.publishAll(msg)
			}
		}
	}()
}

/*
ReceiverRegisterEvent 接收到登入事件处理
 */
func (s *Server) ReceiverRegisterEvent(c *ClientHolder) {
	s.Clients[c.UserName] = c

	go s.Clients[c.UserName].ReadLoop()
	go s.Clients[c.UserName].WriteLoop()

	//发送登入消息
	s.RegisterMsgChan <- []byte(c.UserName + "上线了")
}
