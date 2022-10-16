package main

import (
	"flag"
	"net/http"
	"time"
	"websocket-test/utils"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  readBuffSize,
	WriteBufferSize: writeBuffSize,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

const (
	readBuffSize = 2 << 10
	writeBuffSize
)

var host string

var totalConnections int
var data = []byte("1")

func init() {
	flag.StringVar(&host, "h", "localhost:9000", "the host and port, example -h 1.1.1.1:8000")
	utils.Init()
}

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/{.*}", handler)
	r.HandleFunc("/", handler)

	go utils.PrintCurrentActiveConnections(&utils.LogWriter, &totalConnections)

	utils.LogWriter.Printf("Started Server on host : %v", host)
	http.ListenAndServe(host, r)
}

func handler(rw http.ResponseWriter, r *http.Request) {

	clientid := r.Header.Get(utils.ClientID)
	utils.LogWriter.Debug().Msgf("Got Connection request from client : %s", clientid)
	sublogger := utils.LogWriter.With().Str("clientid", clientid).Logger()

	conn, err := upgrader.Upgrade(rw, r, nil) //Upgrade the conenction
	if err != nil {
		sublogger.Err(err).Msgf("Error while upgrading connection")
		return
	}
	defer conn.Close()

	decr := utils.Incr(&totalConnections)
	defer decr()

	sublogger.Debug().Msgf("Connection Upgraded to websocket successfully")

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			sublogger.Err(err).Msgf("Writing Message Error occurred, will return from handler")
			return
		}
	}
}
