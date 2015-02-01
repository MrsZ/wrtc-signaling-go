package main

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// configure upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func SignalingHandler(writer http.ResponseWriter, req *http.Request) {

	// don't let other requests than GET get through
	if req.Method != "GET" {
		http.Error(writer, "Method not allowed", 405)
		return
	}

	// create new connection
	ws, err := upgrader.Upgrade(writer, req, nil)
	if err != nil {
		http.Error(writer, "An error occured while upgrading request", 500)
		return
	}

	// create node for hub, attach ws connection
	node := &Node{
		Ws: ws,
	}

	// Add the node
	hub.AttachNode <- node

	// finally {}
	defer func() {
		hub.ReleaseNode <- node
		node.Ws.Close()
	}()

	// block this goroutine with reading
	// messages and passing them to the hub
	for {
		_, message, err := node.Ws.ReadMessage()
		if err != nil {
			break
		}
		hub.BroadcastMessage <- broadcastrequest{
			Sender:  node,
			Message: message,
		}

	}

}
