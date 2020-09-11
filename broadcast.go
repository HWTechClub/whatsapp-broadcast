 
package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"encoding/json"
	"io/ioutil"
	"strconv"

	qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/Rhymen/go-whatsapp"
)

//--------------------------------------------------------------------- Struct for broadcast.json -------------------------------------------------------------------------------------
type Broadcast struct {
	Broadcast []User_info `json:"broadcast_field"`
}

type User_info struct {
	Name string `json: "name"`
	Mobile_no   int    `json: "mobile_no"`
}

//---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

type waHandler struct {
	c *whatsapp.Conn
}

//HandleError needs to be implemented to be a valid WhatsApp handler
func (h *waHandler) HandleError(err error) {

	if e, ok := err.(*whatsapp.ErrConnectionFailed); ok {
		log.Printf("Connection failed, underlying error: %v", e.Err)
		log.Println("Waiting 30sec...")
		<-time.After(30 * time.Second)
		log.Println("Reconnecting...")
		err := h.c.Restore()
		if err != nil {
			log.Fatalf("Restore failed: %v", err)
		}
	} else {
		log.Printf("error occoured: %v\n", err)
	}
}

//Optional to be implemented. Implement HandleXXXMessage for the types you need.
func (waHandler *waHandler) HandleTextMessage(message whatsapp.TextMessage) {
	fmt.Printf("%v %v %v %v\n\t%v\n", message.Info.Timestamp, message.Info.Id, message.Info.RemoteJid, message.ContextInfo.QuotedMessageID, message.Text)

		fmt.Println("Message Sent -> UID" + message.Info.RemoteJid)

		if message.Text == "send message"{
			file, _ := ioutil.ReadFile("broadcast.json")

			fmt.Println("here")
 
			data := Broadcast{}
		 
			_ = json.Unmarshal([]byte(file), &data)
		 
			for i := 0; i < len(data.Broadcast); i++ {
				fmt.Println("name: ", data.Broadcast[i].Name)
				fmt.Println("mobile: ", data.Broadcast[i].Mobile_no)
		
				msg := whatsapp.TextMessage{
					Info: whatsapp.MessageInfo{
						RemoteJid: strconv.Itoa(data.Broadcast[i].Mobile_no)+"@s.whatsapp.net",
					},
					Text:        "targeted spam",
				}
				waHandler.c.Send(msg)

				msgId, err := waHandler.c.Send(msg)
	    if err != nil {
		  fmt.Fprintf(os.Stderr, "error sending message: %v", err)
		  //os.Exit(1)
	    } else {
		  fmt.Println("Message Sent -> ID : " + msgId)
		}
			
			}
		}
       /* for{

		msg := whatsapp.TextMessage{
			Info: whatsapp.MessageInfo{
				RemoteJid: "916355536074@s.whatsapp.net",
			},
			Text:        "targeted spam",
		}

	    msgId, err := waHandler.c.Send(msg)
	    if err != nil {
		  fmt.Fprintf(os.Stderr, "error sending message: %v", err)
		  //os.Exit(1)
	    } else {
		  fmt.Println("Message Sent -> ID : " + msgId)
		}
		
	} */

}


func main() {
	//create new WhatsApp connection
	wac, err := whatsapp.NewConn(5 * time.Second)
	if err != nil {
		log.Fatalf("error creating connection: %v\n", err)
	}

	//Add handler
	wac.AddHandler(&waHandler{wac})

	//login or restore
	if err := login(wac); err != nil {
		log.Fatalf("error logging in: %v\n", err)
	}

	//verifies phone connectivity
	pong, err := wac.AdminTest()

	if !pong || err != nil {
		log.Fatalf("error pinging in: %v\n", err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c


	//Disconnect safe
	fmt.Println("Shutting down now.")
	session, err := wac.Disconnect()
	if err != nil {
		log.Fatalf("error disconnecting: %v\n", err)
	}
	if err := writeSession(session); err != nil {
		log.Fatalf("error saving session: %v", err)
	}
}

func login(wac *whatsapp.Conn) error {
	//load saved session
	session, err := readSession()
	if err == nil {
		//restore session
		session, err = wac.RestoreWithSession(session)
		if err != nil {
			return fmt.Errorf("restoring failed: %v\n", err)
		}
	} else {
		//no saved session -> regular login
		qr := make(chan string)
		go func() {
			terminal := qrcodeTerminal.New()
			terminal.Get(<-qr).Print()
		}()
		session, err = wac.Login(qr)
		if err != nil {
			return fmt.Errorf("error during login: %v\n", err)
		}
	}

	//save session
	err = writeSession(session)
	if err != nil {
		return fmt.Errorf("error saving session: %v\n", err)
	}
	return nil
}

func readSession() (whatsapp.Session, error) {
	session := whatsapp.Session{}
	file, err := os.Open(os.TempDir() + "/whatsappSession.gob")
	if err != nil {
		return session, err
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&session)
	if err != nil {
		return session, err
	}
	return session, nil
}

func writeSession(session whatsapp.Session) error {
	file, err := os.Create(os.TempDir() + "/whatsappSession.gob")
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(session)
	if err != nil {
		return err
	}
	return nil
}