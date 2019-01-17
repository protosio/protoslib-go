package protoslib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	util "github.com/protosio/protos/util"
)

const (
	// EventTerminate triggers when the WS connection is closed
	EventTerminate = "terminate"
	// EventTimer triggers periodically, depending on the timer settings in the WS event loop
	EventTimer = "timer"
	// EventNewMessage triggers whenever there is a new message from the WS connection
	EventNewMessage = "newmessage"
)

var eventHandlers = make(map[string]func(...interface{}))

//
// Various event handlers
//

func handleNewMessage(msg []byte) error {
	wsmsg := util.WSMessage{}
	err := json.Unmarshal(msg, &wsmsg)
	if err != nil {
		return errors.Wrap(err, "Failed to decode ws message")
	}

	if wsmsg.MsgType != util.WSMsgTypeUpdate {
		return fmt.Errorf("Failed to process Protos event. Message type %s is not supported", wsmsg.MsgType)
	}

	updateHandler, found := eventHandlers[util.WSMsgTypeUpdate]
	if found != true {
		return fmt.Errorf("Failed to process Protos event. Message type %s has no registered handler", wsmsg.MsgType)
	}

	updateHandler(wsmsg)
	return nil
}

func handleTimer() error {
	timerHandler, found := eventHandlers[EventTimer]
	if found != true {
		return errors.New("Failed to process timer event. No handler registered")
	}

	timerHandler()
	return nil
}

func handleTermination(c *websocket.Conn) {

	// sending a close message to the other peer. Ignoring any error message in case the connection is already closed
	c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "terminating"))
	c.Close()

	// if there is an event handler registered for the termination sequence, call it
	terminationHandler, found := eventHandlers[EventTerminate]
	if found == false {
		return
	}
	terminationHandler()
}

//
// General methods
//

// AddEventHandler registers a function that acts as an event handler for a specific msg type
func (p Protos) AddEventHandler(msgType string, handler func(...interface{})) error {
	switch msgType {
	case EventNewMessage:
		eventHandlers[EventNewMessage] = handler
	case EventTimer:
		eventHandlers[EventTimer] = handler
	case EventTerminate:
		eventHandlers[EventTerminate] = handler
	default:
		return fmt.Errorf("Failed to add event handler. Message type %s is not supported", msgType)
	}
	return nil
}

func wsMessageReader(c *websocket.Conn, messageChan chan []byte, errChan chan error) {
	// read message from Protos and process it
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			errChan <- errors.Wrap(err, "Failed to read ws message")
			return
		}
		messageChan <- message
	}
}

// StartWSLoop opens a websocket connection to Protos and listens for any updates
func (p Protos) StartWSLoop(interval int64) error {
	wsURL := "ws://" + p.Host + "/" + p.PathPrefix + "/ws"
	header := http.Header{"Appid": []string{p.AppID}}

	c, response, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		errMsg, err := decodeError(response)
		if err != nil {
			return errors.Wrap(err, "Failed to establish ws connection")
		}
		return fmt.Errorf("Failed to establish ws connection %s", errMsg)
	}
	defer c.Close()

	// listeting for an interrupt from the OS
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, syscall.SIGINT, syscall.SIGTERM)

	// create a ticker that can be used to check the providers' resources once in a while
	// in case updates are missed for some reasons the provider can still find out about new resources
	ticker := time.NewTicker(time.Second * time.Duration(interval))
	defer ticker.Stop()

	// channel used for receiving messages from the message reader routine
	messageChan := make(chan []byte, 1)
	errChan := make(chan error, 1)
	go wsMessageReader(c, messageChan, errChan)

	// triggering a timer event in the very beginning so that resources providers can do an initial check
	// of all the resources
	err = handleTimer()
	if err != nil {
		handleTermination(c)
		return err
	}

	for {
		select {
		case <-ticker.C:
			err = handleTimer()
			if err != nil {
				handleTermination(c)
				return err
			}
		case <-interruptChan:
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			handleTermination(c)
			return nil
		case err := <-errChan:
			handleTermination(c)
			return err
		case msg := <-messageChan:
			err = handleNewMessage(msg)
			if err != nil {
				handleTermination(c)
				return errors.Wrap(err, "Failed to process event")
			}
		}
	}
}
