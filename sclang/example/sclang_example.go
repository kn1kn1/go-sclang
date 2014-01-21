package main

import (
	// "../../sclang"
	"code.google.com/p/go.net/websocket"
	"flag"
	"fmt"
	"github.com/kn1kn1/go-sclang/sclang"
	"net/http"
	// "os"
)

var port *int = flag.Int("p", 30000, "Port to listen.")

type T struct {
	Tag   string
	Value string
}

type WebSocketWriter struct {
	conn *websocket.Conn
}

func (writer *WebSocketWriter) Write(p []byte) (n int, err error) {
	fmt.Printf("Write\n")
	s := string(p)
	msg := T{"stdout", s}
	err = websocket.JSON.Send(writer.conn, msg)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	fmt.Printf("send:%#v\n", msg)
	return len(p), nil
}

const PathToSclang = "/Applications/SuperCollider/SuperCollider.app/Contents/Resources/"

var sclangObj *sclang.Sclang

// cmdServer handles a json string sent from client using websocket.JSON.
func cmdServer(ws *websocket.Conn) {
	fmt.Printf("cmdServer %#v\n", ws.Config())

	for {
		var msg T
		// Receive receives a text message serialized T as JSON.
		err := websocket.JSON.Receive(ws, &msg)
		if err != nil {
			fmt.Println(err)
			break
		}

		fmt.Printf("recv:%#v\n", msg)
		switch msg.Tag {
		case "init":
			sclangObj, err = sclang.Start(PathToSclang, &WebSocketWriter{ws})
			//sclangObj, err = sclang.Start(PathToSclang, os.Stdout)
		case "start_server":
			err = sclangObj.StartServer()
		case "stop_server":
			err = sclangObj.StopServer()
		case "evaluate":
			err = sclangObj.Evaluate(msg.Value, false)
		case "stop_sound":
			err = sclangObj.StopSound()
		case "toggle_recording":
			err = sclangObj.ToggleRecording()
		case "restart_interpreter":
			err = sclangObj.Restart()
		}
		if err != nil {
			fmt.Println(err)
			break
		}

		// Send send a text message serialized T as JSON.
		err = websocket.JSON.Send(ws, msg)
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Printf("send:%#v\n", msg)
	}
}

func main() {
	flag.Parse()
	http.Handle("/cmd", websocket.Handler(cmdServer))
	http.Handle("/", http.FileServer(http.Dir(".")))
	fmt.Printf("http://localhost:%d/\n", *port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		panic("ListenANdServe: " + err.Error())
	}
}
