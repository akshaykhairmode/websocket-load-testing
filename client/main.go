package main

import (
	"flag"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"websocket-test/utils"

	"github.com/gorilla/websocket"
)

var addr string
var numClients, batchSize int
var totalConnections int //Keeping int for both 32 and 64 bit support
var batcher chan struct{}

func init() {

	flag.IntVar(&numClients, "c", 10, "the number of connections to create")
	flag.IntVar(&batchSize, "b", 100, "the number of connections to initiate at the same time")
	flag.StringVar(&addr, "h", "ws://localhost:9000", "http service address")

	utils.Init()

	_, err := url.Parse(addr)
	if err != nil {
		panic(err)
	}

	batcher = make(chan struct{}, batchSize)

}

func main() {

	wg := &sync.WaitGroup{}
	utils.LogWriter.Info().Msgf("connecting to %s", addr)

	go utils.PrintCurrentActiveConnections(&utils.LogWriter, &totalConnections)

	for i := 0; i < numClients; i++ {
		batcher <- struct{}{}
		wg.Add(1)
		go connect(wg, i, addr)
	}

	wg.Wait()

}

func connect(wg *sync.WaitGroup, clientid int, host string) {
	defer wg.Done()

	header := http.Header{}
	header.Add(utils.ClientID, strconv.Itoa(clientid))

	sublogger := utils.LogWriter.With().Int("clientid", clientid).Logger()

	c, _, err := websocket.DefaultDialer.Dial(host, header)
	if err != nil {
		sublogger.Err(err).Msgf("error while dialing")
		return
	}
	defer c.Close()

	decr := utils.Incr(&totalConnections)
	defer decr()

	sublogger.Debug().Msgf("Connection Successful for client : %v", clientid)
	<-batcher

	for {
		_, message, err := c.ReadMessage() //This is a blocking call
		if err != nil {
			sublogger.Err(err).Msgf("read message error")
			return
		}

		sublogger.Debug().Str("body", string(message)).Msgf("Read Message Successful")
	}
}
