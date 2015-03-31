// Copyright 2015 Aller Media AS.  All rights reserved.
// License: MIT
package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-zoo/bone"
	"github.com/gorilla/websocket"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	// Host and port to bind the server too.
	host string
	// MongoDB address
	mongoHost string
	// html/css/javascript files location
	documentRoot string = "/etc/dashboard/"
)

type wsHub struct {
	// Active connections
	clients map[*client]bool

	// Add new connection
	add chan *client

	// Remove connection
	remove chan *client

	// Send message to all clients
	broadcast chan []byte

	upgrader *websocket.Upgrader
}

func NewWSHub() *wsHub {
	return &wsHub{
		clients:   make(map[*client]bool),
		add:       make(chan *client),
		remove:    make(chan *client),
		broadcast: make(chan []byte),

		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

func (h *wsHub) Run() {
	for {
		select {
		case c := <-h.add:
			log.Println("Added client")
			h.clients[c] = true
		case c := <-h.remove:
			log.Println("Removed client")
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				close(c.send)
			}
		case m := <-h.broadcast:
			for c := range h.clients {
				select {
				case c.send <- m:
				default:
					delete(h.clients, c)
					close(c.send)
				}
			}
		}
	}
}

func (h *wsHub) Write(p []byte) (n int, err error) {
	h.broadcast <- p
	return len(p), nil
}

func (h *wsHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade", r.RemoteAddr)
		return
	}

	c := &client{
		ws:   ws,
		send: make(chan []byte, 256),
	}

	h.add <- c
	defer func() {
		h.remove <- c
	}()
	go c.write()
	c.read()
}

type client struct {
	ws *websocket.Conn

	// Outbound messages for this client
	send chan []byte
}

func (c *client) read() {
	for {
		if _, _, err := c.ws.NextReader(); err != nil {
			c.ws.Close()
			break
		}
	}
}

func (c *client) write() {
	for m := range c.send {
		if err := c.ws.WriteMessage(websocket.TextMessage, m); err != nil {
			break
		}
	}
	c.ws.Close()
}

type BuildEntry struct {
	ProjectName string `json:"name"`
	// Build details URL
	URL       string    `json:"url"`
	Author    string    `json:"author"`
	CommitID  string    `json:"commitID"`
	CommitMsg string    `json:"commitMsg"`
	Date      time.Time `json:"date"`

	Status string `json:"status"`
}

func (h *wsHub) entryBuildPost(w http.ResponseWriter, r *http.Request) {
	var t BuildEntry
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		log.Println(err)
		return
	}

	if t.Date.IsZero() {
		t.Date = time.Now().Local()
	}

	sess, err := connectToMongo()
	if err != nil {
		return
	}
	defer sess.Close()

	collection := sess.DB("dashboard").C("build")
	err = collection.Insert(t)
	if err != nil {
		log.Println("Can't insert document", err)
		return
	}

	js, _ := json.Marshal(t)
	h.broadcast <- js
}

func entryBuildGet(w http.ResponseWriter, r *http.Request) {
	sess, err := connectToMongo()
	if err != nil {
		return
	}
	defer sess.Close()

	collection := sess.DB("dashboard").C("build")
	pipe := collection.Pipe([]bson.M{{"$sort": bson.M{"projectname": 1, "date": 1}}, bson.M{"$group": bson.M{"_id": "$projectname", "docs": bson.M{"$last": bson.M{"name": "$projectname", "status": "$status", "date": "$date"}}}}})

	var results []bson.M
	if err := pipe.All(&results); err != nil {
		log.Println(err)
		// send something to client?
		return
	}
	size := len(results) - 1

	w.Header().Set("Content-Type", "application/vnd.api+json")
	fmt.Fprintln(w, "[")
	for i, v := range results {
		// @todo implement (bt *BuildType) SetBSON()
		t := BuildEntry{
			ProjectName: v["docs"].(bson.M)["name"].(string),
			Status:      v["docs"].(bson.M)["status"].(string),
			Date:        v["docs"].(bson.M)["date"].(time.Time),
		}

		js, _ := json.Marshal(t)
		fmt.Fprint(w, string(js))

		if i != size {
			fmt.Fprintln(w, ",")
		} else {
			fmt.Println("")
		}
	}
	fmt.Fprintln(w, "]")
}

func connectToMongo() (*mgo.Session, error) {
	sess, err := mgo.Dial(mongoHost)
	if err != nil {
		log.Println("Can't connect to mongo", err)
		return nil, err
	}
	sess.SetSafe(&mgo.Safe{})

	return sess, nil
}

func flagParse() {
	host = os.Getenv("DASHBOARD_HOST")
	if host == "" {
		log.Fatalln("Missing DASHBOARD_HOST")
	}

	mongoHost = os.Getenv("DASHBOARD_MONGO_ADDR")
	if mongoHost == "" {
		mongoHost = "mongo"
		log.Println("DASHBOARD_MONGO_ADDR not specified, defaulting to \"mongo\"")
	}
}

func main() {
	flagParse()

	wsh := NewWSHub()
	go wsh.Run()

	mux := bone.New()
	mux.NotFound(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("404 not found"))
	})

	mux.Get("/entry/build", http.HandlerFunc(entryBuildGet))
	mux.Post("/entry/build", http.HandlerFunc(wsh.entryBuildPost))

	mux.Handle("/receive", wsh)

	mux.Get("/", http.StripPrefix("/", http.FileServer(http.Dir(documentRoot))))
	mux.Get("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir(documentRoot+"js"))))
	mux.Get("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir(documentRoot+"css"))))

	log.Println("Server started on", host)
	log.Fatal(http.ListenAndServe(host, mux))
}
