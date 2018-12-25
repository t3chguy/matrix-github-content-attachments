package main

import (
	"fmt"
	"github.com/matrix-org/gomatrix"
	"github.com/patrickmn/go-cache"
	"golang.org/x/net/context"
	"net/http"
	"time"
)

type Client struct {
	*gomatrix.Client
	RoomInfoCache *cache.Cache
}

type respStateMRoomName struct {
	Name string `json:"name"`
}
type respStateMRoomTopic struct {
	Topic string `json:"topic"`
}

type RoomInfo struct {
	Name  string
	Topic string
}

type RespRoomDirectoryAlias struct {
	RoomID  string   `json:"room_id"`
	Servers []string `json:"servers"`
}

func (c *Client) GetRoomDirectoryAlias(roomAlias string) (resp *RespRoomDirectoryAlias, err error) {
	urlPath := c.BuildURL("directory", "room", roomAlias)
	_, err = c.MakeRequest("GET", urlPath, nil, &resp)
	return
}

func (c Client) getRoomState(roomID, stateKey string, resp interface{}) (err error) {
	urlPath := c.BuildURL("rooms", roomID, "state", stateKey)
	_, err = c.MakeRequest("GET", urlPath, nil, &resp)
	return
}
func (c Client) getRoomName(roomID string, ch chan<- string) {
	var resp respStateMRoomName
	err := c.getRoomState(roomID, "m.room.name", &resp)
	if err != nil {
		fmt.Println(err)
		return
	}
	ch <- resp.Name
}
func (c Client) getRoomTopic(roomID string, ch chan<- string) {
	var resp respStateMRoomTopic
	err := c.getRoomState(roomID, "m.room.topic", &resp)
	if err != nil {
		fmt.Println(err)
		return
	}
	ch <- resp.Topic
}

func (c Client) getRoomInfo(roomID string, ch chan<- *RoomInfo) {
	nameChan := make(chan string)
	topicChan := make(chan string)
	defer close(nameChan)
	defer close(topicChan)

	go c.getRoomName(roomID, nameChan)
	go c.getRoomTopic(roomID, topicChan)
	ch <- &RoomInfo{<-nameChan, <-topicChan}
}

func (c Client) GetRoomInfo(ctx context.Context, roomIDorAlias string) *RoomInfo {
	roomID := roomIDorAlias
	if roomIDorAlias[0] == '#' {
		r, err := c.GetRoomDirectoryAlias(roomIDorAlias)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		roomID = r.RoomID
	}

	ch := make(chan *RoomInfo)
	go c.getRoomInfo(roomID, ch)

	select {
	case <-ctx.Done():
		<-ch
		fmt.Println(ctx.Err())
		return nil
	case resp := <-ch:
		return resp
	}
}

func (c Client) GetCachedRoomInfo(ctx context.Context, roomIDorAlias string) *RoomInfo {
	if item, found := c.RoomInfoCache.Get(roomIDorAlias); found {
		return item.(*RoomInfo)
	}

	roomInfo := c.GetRoomInfo(ctx, roomIDorAlias)
	c.RoomInfoCache.SetDefault(roomIDorAlias, roomInfo)
	return roomInfo
}

func NewClient(homeserverURL, userID, accessToken string) (*Client, error) {
	cli, err := gomatrix.NewClient(homeserverURL, userID, accessToken)
	if err != nil {
		return nil, err
	}

	cli.Client = &http.Client{
		Timeout: 30 * time.Second,
	}
	return &Client{
		Client:        cli,
		RoomInfoCache: cache.New(60*time.Minute, 60*time.Minute),
	}, err
}
