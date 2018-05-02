package main

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	r "gopkg.in/gorethink/gorethink.v4"
)

func addChannel(client *Client, data interface{}) {
	var channel Channel

	err := mapstructure.Decode(data, &channel)
	if err != nil {
		client.send <- Message{"error", err.Error()}
		return
	}

	go func() {
		err = r.Table("channel").Insert(channel).Exec(client.session)
		if err != nil {
			client.send <- Message{"error", err.Error()}
			fmt.Println(err.Error())
		}
	}()
}
