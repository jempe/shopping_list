package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/gorilla/websocket"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/qor/admin"
	"github.com/qor/qor"

	"github.com/jempe/shopping_list/models/lists"
)

type Configuration struct {
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

var indexTemplate *template.Template
var config Configuration

type remote struct {
	forward  chan message
	join     chan *client
	leave    chan *client
	clients  map[*client]bool
	database *gorm.DB
}

type client struct {
	socket   *websocket.Conn
	send     chan []byte
	remote   *remote
	isClient bool
	//remoteId string
	//clientID string
}

type message struct {
	msg []byte
	//senderId string
}

func NewRemote(db *gorm.DB) *remote {
	return &remote{
		forward:  make(chan message),
		join:     make(chan *client),
		leave:    make(chan *client),
		clients:  make(map[*client]bool),
		database: db,
	}
}

func (r *remote) Run() {
	for {
		select {
		case client := <-r.join:
			r.clients[client] = true
			log.Println("New client joined")
		case client := <-r.leave:
			delete(r.clients, client)
			close(client.send)
			log.Println("Client left")
		case message := <-r.forward:
			msg := message.msg
			log.Println("Message received: ", string(msg))

			for client := range r.clients {
				client.send <- msg
				log.Println("message sent to Clients")
			}
		}
	}
}

var upgrader = &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

func (r *remote) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)

	if err != nil {
		log.Fatal("ServeHTTP error:", err)
		return
	}

	for _, status := range r.clients {
		log.Println("client status", status)
	}

	client := &client{
		socket: socket,
		send:   make(chan []byte, 256),
		remote: r,
	}
	r.join <- client
	defer func() {
		r.leave <- client
	}()
	go client.write()
	client.read()
}

func (c *client) read() {
	defer c.socket.Close()
	for {
		_, msg, err := c.socket.ReadMessage()
		if err != nil {
			return
		}
		c.remote.forward <- message{msg: msg}
	}
}

func (c *client) write() {
	defer c.socket.Close()
	for msg := range c.send {
		err := c.socket.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return
		}
	}
}

func main() {
	file, err := ioutil.ReadFile("./config/config.json")
	panicError(err)

	json.Unmarshal(file, &config)

	gormDB, dbError := gorm.Open("sqlite3", "shopping_lists.db")

	if dbError != nil {
		log.Println(dbError)
		os.Exit(1)
	}

	gormDB.AutoMigrate(&lists.List{}, &lists.Item{})

	Admin := admin.New(&admin.AdminConfig{
		DB:       gormDB,
		SiteName: "Shopping Lists",
	})

	Admin.AddResource(&lists.List{})
	Admin.AddResource(&lists.Item{})

	API := admin.New(&qor.Config{DB: gormDB})

	API.AddResource(&lists.List{})
	item := API.AddResource(&lists.Item{})

	item.Action(&admin.Action{
		Name: "check",
		Handler: func(actionArgument *admin.ActionArgument) error {
			for _, record := range actionArgument.FindSelectedRecords() {
				actionArgument.Context.DB.Model(record.(*lists.Item)).Update("Checked", true)
			}

			return nil
		},
	})

	item.Action(&admin.Action{
		Name: "uncheck",
		Handler: func(actionArgument *admin.ActionArgument) error {
			for _, record := range actionArgument.FindSelectedRecords() {
				actionArgument.Context.DB.Model(record.(*lists.Item)).Update("Checked", false)
			}

			return nil
		},
	})

	//setup homepage template
	paths := []string{"tmpl/index.html"}
	indexTemplate = template.Must(template.ParseFiles(paths...))

	// setup http handlers
	mux := http.NewServeMux()

	Admin.MountTo("/admin", mux)
	API.MountTo("/api", mux)

	mux.Handle("/system/", http.FileServer(http.Dir("public")))

	mux.HandleFunc("/", pageHandler)

	remote := NewRemote(gormDB)

	mux.Handle("/remote", remote)

	go remote.Run()

	log.Println("Serve running on port:", config.Port)

	// If config file has a username and password, use basic Auth middleware
	if config.Username != "" && config.Password != "" {
		panic(http.ListenAndServe(":"+config.Port, basicAuth(mux)))
	} else {
		panic(http.ListenAndServe(":"+config.Port, mux))
	}
}

func basicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		username, password, ok := r.BasicAuth()
		if ok {
			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))
			expectedUsernameHash := sha256.Sum256([]byte(config.Username))
			expectedPasswordHash := sha256.Sum256([]byte(config.Password))

			usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
			passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)

			if usernameMatch && passwordMatch {
				next.ServeHTTP(w, r)
				return
			} else {
				log.Println("login error")
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		indexTemplate.Execute(w, nil)
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Page Not Found"))
	}
}
func logAndExit(message string) {
	log.Println(message)
	os.Exit(1)
}
func panicError(err error) {
	if err != nil {
		logAndExit(err.Error())
	}
}
