package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
	"time"
)

// This file heavily borrows from https://github.com/gorilla/websocket/tree/master/examples/chat

type Client struct {
	conn *websocket.Conn

	responses chan []byte
}

type JobResponse struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type ClientResponse struct {
	Job     *JobResponse `json:"job"`
	Status  string       `json:"status"`
	Message string       `json:"message"`
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	// todo remove?
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (c *Client) queueResponse(job *Job, status JobStatusType, message string) {
	response := ClientResponse{Status: status.String(), Message: message}

	if job != nil {
		var name string
		if job.settings.Pdf {
			name = job.name
		} else {
			name = job.dir
		}
		response.Job = &JobResponse{Id: job.id, Name: name}
	}

	responseJson, err := json.Marshal(response)

	if err != nil {
		log.Printf("queueResponse error: %v", err)
		return
	}

	c.responses <- responseJson
}

func (c *Client) readRequests() {
	defer func() { _ = c.conn.Close() }()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(
		func(string) error {
			_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		},
	)

	for {
		_, request, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		log.Printf("job request: %v", request)

		// todo: make this json or something?
		if values, err := url.ParseQuery(string(request)); err == nil {
			job := service.parseJob(values)
			job.client = c
			service.ScanJobs <- job
			job.report(StatusInProgress, "queued for scanning")
		} else {
			log.Printf("job request error: %v", err)
			c.queueResponse(nil, StatusFailed, fmt.Sprintf("failed: %v", err))
		}
	}
}

func (c *Client) writeResponses() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		log.Println("client disconnected")
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case response, ok := <-c.responses:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			// channel was closed
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			_, _ = w.Write(response)
			if err = w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{conn: conn, responses: make(chan []byte, 256)}

	go client.readRequests()
	go client.writeResponses()
}
