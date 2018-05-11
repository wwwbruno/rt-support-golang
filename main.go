package main

import (
	"log"
	"net/http"

	r "gopkg.in/gorethink/gorethink.v4"
)

type Channel struct {
	Id   string `json:"id" gorethink:"id,omitempty"`
	Name string `json:"name" gorethink:"name"`
}

type User struct {
	Id   string `gorethink:"id,omitempty"`
	Name string `gorethink:"name"`
}

type ChannelMessage struct {
	Id        string `json:"id" gorethink:"id,omitempty"`
	Body      string `json:"body" gorethink:"body"`
	User      string `json:"user" gorethink:"user"`
	ChannelId string `json:"channelId" gorethink:"channelId"`
}

func main() {
	session, err := r.Connect(r.ConnectOpts{
		Address:  "localhost:28015",
		Database: "rtsupport",
	})
	if err != nil {
		log.Panic(err.Error())
	}

	router := NewRouter(session)

	// channel handlers
	router.Handle("channel add", addChannel)
	router.Handle("channel subscribe", subscribeChannel)
	router.Handle("channel unsubscribe", unsubscribeChannel)

	// user handlers
	router.Handle("user edit", editUser)
	router.Handle("user subscribe", subscribeUser)
	router.Handle("user unsubscribe", unsubscribeUser)

	// message handlers
	router.Handle("message add", addMessage)
	router.Handle("message subscribe", subscribeMessage)
	router.Handle("message unsubscribe", unsubscribeMessage)

	http.Handle("/", router)
	http.ListenAndServe(":4000", nil)
}
