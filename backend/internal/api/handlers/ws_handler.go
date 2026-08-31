package handlers

import (
	"log"
	"net/http"

	"exchange/internal/services"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins
	},
}

type WSHandler struct {
	hub *services.WSHub
}

func NewWSHandler(hub *services.WSHub) *WSHandler {
	return &WSHandler{hub: hub}
}

func (h *WSHandler) ServeWS(c echo.Context) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Println("websocket upgrade error:", err)
		return err
	}

	client := &services.Client{
		Conn: conn,
		Send: make(chan map[string]float64, 256),
	}

	h.hub.Register <- client

	// Start the write pump for this client
	go client.WritePump()

	// Keep the read pump alive just to detect disconnects
	go func() {
		defer func() {
			h.hub.Unregister <- client
			conn.Close()
		}()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("websocket error: %v", err)
				}
				break
			}
		}
	}()

	return nil
}
