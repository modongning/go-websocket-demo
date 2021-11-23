package sdk

import (
	"encoding/json"
	"log"
)

type Message struct {
	To  string `json:"to"`
	Msg interface{} `json:"msg"`
}

func NewMessage(msg []byte) *Message {
	log.Println("收到消息："+string(msg))

	message := &Message{}
	if err := json.Unmarshal(msg, message);err !=nil{
		log.Println("error:"+err.Error())
	}

	log.Println(message.Msg)

	return message
}
