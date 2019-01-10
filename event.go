package protoslib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	util "github.com/protosio/protos/util"
)

var eventHandlers = make(map[string]func(interface{}))

func processEvent(msg []byte) error {
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

// AddEventHandler registers a function that acts as an event handler for a specific msg type
func (p Protos) AddEventHandler(msgType string, handler func(interface{})) error {
	switch msgType {
	case util.WSMsgTypeUpdate:
		eventHandlers[util.WSMsgTypeUpdate] = handler
	default:
		return fmt.Errorf("Failed to add event handler. Message type %s is not supported", msgType)
	}
	return nil
}

// StartWSLoop opens a websocket connection to Protos and listens for any updates
func (p Protos) StartWSLoop(interval int64) error {
	wsURL := "ws://" + p.URL + "ws"
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
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// create a ticker that can be used to check the providers' resources once in a while
	// in case updates are missed for some reasons the provider can still find out about new resources
	ticker := time.NewTicker(time.Second * time.Duration(interval))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			return nil
		case <-interrupt:
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				return errors.Wrap(err, "Failed to close ws connection")
			}
			return nil
		default:
			// read message from Protos and process it
			_, message, err := c.ReadMessage()
			if err != nil {
				return errors.Wrap(err, "Failed to read ws message")
			}
			err = processEvent(message)
			if err != nil {
				return errors.Wrap(err, "Failed to process event")
			}
		}
	}
}
