package handlers

import (
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
)

var wsChan = make(chan WsPayload)
var clients = make(map[WebsocketConnection]string)

var views = jet.NewSet(
	jet.NewOSFileSystemLoader("./html"),
	jet.InDevelopmentMode(),
)

var connUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

/* Home renders the home page */
func Home(w http.ResponseWriter, r *http.Request) {
	err := renderPage(w, "home.jet", nil)
	if err != nil {
		log.Println(err)
	}
}

type WebsocketConnection struct {
	*websocket.Conn
}

/* WsJsonResponse defines the response sent back from websocket */
type WsJsonResponse struct {
	Action         string   `json:"action"`
	Message        string   `json:"message"`
	MessageType    string   `json:"message_type"`
	ConnectedUSers []string `json:"connected_users"`
}

type WsPayload struct {
	Action   string              `json:"action"`
	Username string              `json:"username"`
	Message  string              `json:"message"`
	Conn     WebsocketConnection `json:"_"`
}

/* WsEndpoint upgrades connection to websocket */
func WsEndpoint(w http.ResponseWriter, r *http.Request) {
	ws, err := connUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	log.Println("Client connected to endpoint")

	var response WsJsonResponse
	response.Action = "welcome"
	response.Message = `<em><small>Welcome to the chat</small></em>`
	ws.WriteJSON(response)

	conn := WebsocketConnection{Conn: ws}
	clients[conn] = ""

	err = ws.WriteJSON(response)
	if err != nil {
		log.Println(err)
	}

	users := getUserList()
	response.Action = "list_users"
	response.ConnectedUSers = users
	broadcastToAll(response)

	go ListenForWs(&conn)
}

/* Listen to Ws Channel in server & send message to client */
func ListenToWsChannel() {
	var response WsJsonResponse

	for {
		e := <-wsChan

		switch e.Action {
		case "username":
			// get a list of all users and send it back via broadcast'
			clients[e.Conn] = e.Username
			users := getUserList()
			response.Action = "list_users"
			response.ConnectedUSers = users
			broadcastToAll(response)

		case "left":
			delete(clients, e.Conn)

			users := getUserList()
			response.Action = "list_users"
			response.ConnectedUSers = users
			broadcastToAll(response)
		case "broadcast":
			response.Action = "broadcast"
			response.Message = fmt.Sprintf("<strong>%s</strong>: %s", e.Username, e.Message)
			broadcastToAll(response)
		}

	}
}

/* listen for websocket message from client */
func ListenForWs(conn *WebsocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("\n[Error]", fmt.Sprintf("%v", r))
		}
	}()

	var payload WsPayload

	for {
		err := conn.ReadJSON(&payload)
		if err != nil {
			break
		} else {
			payload.Conn = *conn
			wsChan <- payload
		}
	}
}

/* get a list of connected users */
func getUserList() []string {
	var userList []string
	for _, x := range clients {
		if x != "" {
			userList = append(userList, x)
		}
	}
	sort.Strings(userList)
	return userList
}

/* send message to all connected clients */
func broadcastToAll(response WsJsonResponse) {
	for client := range clients {
		err := client.WriteJSON(response)
		if err != nil {
			log.Println("websocket error")
			_ = client.Close()
			delete(clients, client)
		}
	}
}

/* renderPage renders a jet template */
func renderPage(w http.ResponseWriter, tmpl string, data jet.VarMap) error {
	view, err := views.GetTemplate(tmpl)
	if err != nil {
		log.Println(err)
		return err
	}

	err = view.Execute(w, data, nil)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
