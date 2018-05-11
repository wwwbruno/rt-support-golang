package main

import (
	"fmt"

	"github.com/gorilla/websocket"
	r "gopkg.in/gorethink/gorethink.v4"
)

type Message struct {
	Name string      `json:"name"`
	Data interface{} `json:"data"`
}

type FindHandler func(string) (Handler, bool)

type Client struct {
	send         chan Message
	socket       *websocket.Conn
	findHandler  FindHandler
	session      *r.Session
	stopChannels map[int]chan bool
	id           string
	userName     string
}

func (client *Client) NewStopChannel(stopKey int) chan bool {
	client.StopForKey(stopKey)

	stop := make(chan bool)
	client.stopChannels[stopKey] = stop
	return stop
}

func (client *Client) StopForKey(key int) {
	if channel, found := client.stopChannels[key]; found {
		channel <- true
		delete(client.stopChannels, key)
	}
}

func (client *Client) Close() {
	for _, channel := range client.stopChannels {
		channel <- true
	}
	close(client.send)

	r.Table("user").Get(client.id).Delete().Exec(client.session)
}

func (client *Client) Read() {
	var message Message
	for {
		if err := client.socket.ReadJSON(&message); err != nil {
			break
		}
		if handler, found := client.findHandler(message.Name); found {
			handler(client, message.Data)
		}
	}

	client.socket.Close()
}

func (client *Client) Write() {
	for msg := range client.send {
		if err := client.socket.WriteJSON(msg); err != nil {
			break
		}
	}

	client.socket.Close()
}

func NewClient(socket *websocket.Conn, findHandler FindHandler, session *r.Session) *Client {
	var user User
	var id string

	user.Name = "anonymous"

	response, err := r.Table("user").Insert(user).RunWrite(session)
	if err != nil {
		fmt.Println(err.Error())
	}

	if len(response.GeneratedKeys) > 0 {
		id = response.GeneratedKeys[0]
	}

	return &Client{
		send:         make(chan Message),
		socket:       socket,
		findHandler:  findHandler,
		session:      session,
		stopChannels: make(map[int]chan bool),
		id:           id,
		userName:     user.Name,
	}
}
