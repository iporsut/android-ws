package main

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

var namedChannels map[string]chan string

func init() {
	namedChannels = make(map[string]chan string)
}

func echoHandler(ws *websocket.Conn) {
	var req map[string]string
	var timeoutChan chan bool = make(chan bool)
	var timeout = false
	msg := make([]byte, 6000)
	n, err := ws.Read(msg)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(msg[:n],&req)

	if id , ok := req["id"]; ok {
		if namedChannel, ok := namedChannels[id]; ok {
			go func(timeoutChan chan bool) {
				var timeout = false
				for !timeout {
					select {
						case <- time.After(30 * time.Second):
							timeout = true
							timeoutChan <- true
					}
				}
			}(timeoutChan)
			for !timeout {
				select {
					case m := <- namedChannel:
						_, err := ws.Write([]byte(m))
						if err != nil {
							log.Fatal(err)
						}
					case <- timeoutChan:
						_, err := ws.Write([]byte("Time Out"))
						if err != nil {
							log.Fatal(err)
						}
						delete(namedChannels, id)
						timeout = true
				}
			}
			fmt.Println("Break!")		
			err = ws.Close()
			if err != nil {
				log.Fatal(err)
			}
		} else {
			_, err := ws.Write([]byte("Don't Create Channel"))
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func createChannel (w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}
	id := r.FormValue("id")
	if _ , ok := namedChannels[id]; !ok {
		namedChannels[id] = make(chan string)
	}
}

func sendData (w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}
	id := r.FormValue("id")
	data := r.FormValue("data")
	if namedChannel , ok := namedChannels[id]; ok {
		namedChannel <- data
	}
}

func main() {
	http.Handle("/echo", websocket.Handler(echoHandler))
	http.Handle("/createChannel", http.HandlerFunc(createChannel))
	http.Handle("/sendData", http.HandlerFunc(sendData))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
