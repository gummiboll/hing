package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	_ "github.com/gummiboll/hing/statik"
	"github.com/rakyll/statik/fs"
	"golang.org/x/net/websocket"
)

var (
	listenAddr string
	listenPort int
	devBool    bool
)

func init() {
	flag.StringVar(&listenAddr, "l", "localhost", "Which address to listen to, defaults to localhost")
	flag.IntVar(&listenPort, "p", 8080, "Which port to listen to, defaults to 8080")
	flag.BoolVar(&devBool, "dev", false, "Run in develop mode")
}

func main() {
	flag.Parse()

	// Setup channels
	rChan := make(chan *RequestTarget)
	dChan := make(chan bool)
	allResChan := make(chan RequestResult, 100)

	// Start broker
	go broker(rChan, allResChan, dChan)

	postTarget := makePostTarget(rChan, dChan)
	readData := makeReadData(allResChan)

	listen := fmt.Sprintf("%s:%d", listenAddr, listenPort)
	log.Printf("Listening on: %s\n", listen)

	if devBool == true {
		fs := http.FileServer(http.Dir("static"))
		http.Handle("/", fs)
	} else {
		fs, _ := fs.New()
		http.Handle("/", http.FileServer(fs))
	}

	http.HandleFunc("/target", postTarget)
	http.Handle("/ws", websocket.Handler(readData))
	http.ListenAndServe(listen, nil)

}

func makeReadData(allResChan chan RequestResult) func(*websocket.Conn) {
	return func(ws *websocket.Conn) {
		for {
			select {
			case res := <-allResChan:
				if err := websocket.JSON.Send(ws, res); err != nil {
					fmt.Printf("Can't send: %s", err.Error())
					fmt.Println(res)
					break
				}
			}
		}
	}
}

func makePostTarget(rChan chan *RequestTarget, dChan chan bool) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Panicf("Could not parse body %s\n", string(body))
		}

		rt := NewRequestTarget()
		err = json.Unmarshal(body, &rt)
		if err != nil {
			log.Panicf("Could not unmarshal json: %s\n", string(body))
		}

		if rt.Stop == true {
			dChan <- true
			return
		}

		rChan <- rt
	}
}

func broker(rChan chan *RequestTarget, allResChan chan RequestResult, dChan chan bool) {
	resChan := make(chan RequestResult)
	for {
		select {
		case rt := <-rChan:
			go requester(resChan, *rt, dChan)
		case res := <-resChan:
			allResChan <- res
		}
	}
}

func requester(resChan chan RequestResult, rt RequestTarget, dChan chan bool) {
	log.Printf("Doing test on %s.\n", rt.URL)
	tickChan := time.NewTicker(time.Second * time.Duration(rt.SleepReq)).C
	times := 1

	// Don't wait for the first tick, send a request right away
	go DoRequest(resChan, rt, times)

	for {
		select {
		case <-tickChan:
			times++
			go DoRequest(resChan, rt, times)
		case <-dChan:
			log.Printf("Stopping test on %s\n", rt.URL)
			return
		}
	}

}
