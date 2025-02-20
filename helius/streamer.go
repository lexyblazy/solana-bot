package helius

import (
	"fmt"
	"log"
	"os"
	"solana-bot/config"
	"strings"

	"github.com/gorilla/websocket"
)

func createConnection(c *config.HeliusConfig) *websocket.Conn {
	wsUrl := fmt.Sprintf("%s?api-key=%s", c.WebSocketUrl, c.ApiKey)

	conn, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)

	if err != nil {
		log.Fatal("Failed to establish websocket connect")
	}

	log.Println("Connection established")

	return conn
}

type Streamer struct {
	conn   *websocket.Conn
	msgCh  chan []byte
	config *config.HeliusConfig
}

func (s *Streamer) Close() {
	s.conn.Close()
}

func (s *Streamer) Reconnect() {

	log.Println("Reconnecting...")

	// create new connection
	s.conn = createConnection(s.config)
	// subscribe to logs again
	s.SubscribeToLogs()
	// start reading messages into the channel
	s.ReadMessages()

}

func (s *Streamer) GetMessageChannel() chan []byte {
	return s.msgCh
}

func (s *Streamer) ReadMessages() {
	for {

		_, message, err := s.conn.ReadMessage()

		if err != nil {
			log.Println("ReadMessages:", err)

			if strings.Contains(err.Error(), "connection reset") {
				log.Println("Connection reset: Handle reconnection")
			} else {
				s.Close()
			}

			s.Reconnect()

			return
		}

		s.msgCh <- message

	}
}

func (s *Streamer) SubscribeToLogs() {

	contents, err := os.ReadFile("./logSubscribe.json")

	if err != nil {
		log.Fatal("failed to open logSubscribe file")
	}

	err = s.conn.WriteMessage(websocket.TextMessage, contents)

	if err != nil {
		log.Println("write:", err)
		return
	}

}

func NewStreamer(c *config.HeliusConfig) *Streamer {
	return &Streamer{
		conn:   createConnection(c),
		msgCh:  make(chan []byte, 1024),
		config: c,
	}
}
