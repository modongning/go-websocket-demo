package main

import (
	"github.com/gorilla/websocket"
	"net/http"
	"websocket_demo/sdk"
)

/*
main 主程序入口
*/
func main() {
	//
	sdk.ServerRuntime.Run()

	//注册websocket路由
	http.HandleFunc("/ws", MessageHandler)
	//监听8900端口，这里没有实现服务器异常处理器，所以第二个参数是nil
	http.ListenAndServe(":8900", nil)
}

/*
MessageHandler
websocket链接处理器
*/
func MessageHandler(res http.ResponseWriter, req *http.Request) {

	//创建socket链接
	var upgrader = websocket.Upgrader{
		//允许跨域
		CheckOrigin: func(req *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(res, req, nil)
	if err != nil {
		panic(err.Error())
	}
	client, err := sdk.Create(conn)
	if err != nil {
		panic(err.Error())
	}

	req.ParseForm() //对表单参数进行转换，转换后才可以获取到参数
	name := req.Form.Get("name")
	client.UserName = name

	//发送客户端登入事件
	sdk.ServerRuntime.RegisterEventChan <- client
}
