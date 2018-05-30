package main

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	r "gopkg.in/gorethink/gorethink.v4"
)

const (
	ChannelStop = iota
	UserStop
	MessageStop
)

func addChannel(client *Client, data interface{}) {
	var channel Channel
	adder(client, "channel", data, &channel)
}

func subscribeChannel(client *Client, data interface{}) {
	cursor, err := r.Table("channel").
		Changes(r.ChangesOpts{IncludeInitial: true}).
		Run(client.session)
	if err != nil {
		client.send <- Message{"Error", err.Error()}
		return
	}

	subscriber(client, cursor, ChannelStop, "channel")
}

func unsubscribeChannel(client *Client, data interface{}) {
	client.StopForKey(ChannelStop)
}

func editUser(client *Client, data interface{}) {
	var user User

	err := mapstructure.Decode(data, &user)
	if err != nil {
		client.send <- Message{"error", err.Error()}
		return
	}

	client.userName = user.Name

	go func() {
		err = r.Table("user").Get(client.id).Update(user).Exec(client.session)
		if err != nil {
			client.send <- Message{"error", err.Error()}
			fmt.Println(err.Error())
		}
	}()
}

func subscribeUser(client *Client, data interface{}) {
	cursor, err := r.Table("user").
		Changes(r.ChangesOpts{IncludeInitial: true}).
		Run(client.session)
	if err != nil {
		client.send <- Message{"Error", err.Error()}
		return
	}

	subscriber(client, cursor, UserStop, "user")
}

func unsubscribeUser(client *Client, data interface{}) {
	client.StopForKey(UserStop)
}

func addMessage(client *Client, data interface{}) {
	var channelMessage ChannelMessage
	channelMessage.User = client.userName

	adder(client, "message", data, &channelMessage)
}

func subscribeMessage(client *Client, data interface{}) {
	eventData := data.(map[string]interface{})
	val, _ := eventData["channelId"]
	channelID, _ := val.(string)

	cursor, err := r.Table("message").
		Filter(r.Row.Field("channelId").Eq(channelID)).
		Changes(r.ChangesOpts{IncludeInitial: true}).
		Run(client.session)
	if err != nil {
		client.send <- Message{"Error", err.Error()}
		return
	}

	subscriber(client, cursor, MessageStop, "message")
}

func unsubscribeMessage(client *Client, data interface{}) {
	client.StopForKey(MessageStop)
}

func adder(client *Client, channel string, data interface{}, message interface{}) {
	err := mapstructure.Decode(data, message)
	if err != nil {
		client.send <- Message{"error", err.Error()}
		return
	}

	go func() {
		err = r.Table(channel).Insert(message).Exec(client.session)
		if err != nil {
			client.send <- Message{"error", err.Error()}
			fmt.Println(err.Error())
		}
	}()
}

func subscriber(client *Client, cursor *r.Cursor, stopKey int, channel string) {
	result := make(chan r.ChangeResponse)
	stop := client.NewStopChannel(stopKey)

	go func() {
		var change r.ChangeResponse
		for cursor.Next(&change) {
			result <- change
		}
	}()

	go func() {
		for {
			select {
			case <-stop:
				cursor.Close()
				return
			case change := <-result:
				if change.NewValue != nil && change.OldValue == nil {
					client.send <- Message{fmt.Sprintf("%s add", channel), change.NewValue}
				} else if change.NewValue != nil && change.OldValue != nil {
					client.send <- Message{fmt.Sprintf("%s edit", channel), change.NewValue}
				} else if change.NewValue == nil && change.OldValue != nil {
					client.send <- Message{fmt.Sprintf("%s remove", channel), change.OldValue}
				}
			}
		}
	}()
}
