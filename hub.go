package main

import (
	"crypto/rand"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

//
type Hub struct {
	AttachNode       chan *Node
	ReleaseNode      chan *Node
	BroadcastMessage chan broadcastrequest
	StopHub          chan bool

	nodes []Node
}

//
type Node struct {
	Id string
	Ws *websocket.Conn
}

type broadcastrequest struct {
	Sender  *Node
	Message []byte
}

// Hub messaging type representation
type Message struct {
	Type string   `json:"type"`
	Data UserList `json:"data"`
}

type UserList struct {
	Users []User `json:"users"`
}

type User struct {
	Id string `json:"id"`
}

// construct a new Hub instance
func NewHub() Hub {
	var result Hub = Hub{
		BroadcastMessage: make(chan broadcastrequest),
		AttachNode:       make(chan *Node),
		ReleaseNode:      make(chan *Node),
		StopHub:          make(chan bool),

		nodes: make([]Node, 0), // init empty for json serializing
	}
	return result
}

// will listen to hub channels
// and process them
func (h *Hub) Start() {

	for {
		select {

		//
		case n := <-h.AttachNode:
			n.Id = uuid()
			h.nodes = append(h.nodes, *n)
			log.Println("HUB_ATTACH_NODE " + n.Id + " " + n.Ws.RemoteAddr().String())
			h.Describe()

		// send
		case b := <-h.BroadcastMessage:
			sender := b.Sender
			targets := excludeNode(h.nodes, sender)
			for _, target := range targets {
				err := target.Ws.WriteMessage(websocket.TextMessage, b.Message)
				if err != nil {
					log.Println("Unable to write message")
				}
			}

			log.Println("HUB_BROADCAST_MESSAGE " + string(b.Message))

		case n := <-h.ReleaseNode:
			log.Println("HUB_RELEASE_NODE " + n.Id + " " + n.Ws.RemoteAddr().String())
			h.nodes = excludeNode(h.nodes, n)

		case <-h.StopHub:
			log.Println("Stop the hub")
			break

		}
	}
}

func (h *Hub) Stop() {
	h.StopHub <- true
}

func (h *Hub) Describe() {
	for _, node := range h.nodes {
		if node.Ws != nil {
			m := Message{
				Type: "DESCRIBE_ROOM",
				Data: UserList{
					Users: make([]User, 0),
				},
			}

			users := createUserList(excludeNode(h.nodes, &node))

			if users != nil {
				m.Data.Users = users
			}

			node.Ws.WriteJSON(m)
		}
	}
}

// creates a list of users for client hub representation
func createUserList(nodes []Node) []User {
	var users []User
	for _, node := range nodes {
		if node.Id != "" {
			users = append(users, User{Id: node.Id})
		}
	}
	return users
}

// return a slice without the node specified as target
func excludeNode(nodes []Node, target *Node) []Node {
	var result []Node
	for _, node := range nodes {
		if node.Id != target.Id && node.Id != "" {
			result = append(result, node)
		}
	}
	return result
}

func uuid() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
