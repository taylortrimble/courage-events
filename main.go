package main

import (
	"code.google.com/p/go-uuid/uuid"
	"github.com/thenewtricks/courage-service/binmsg"
	"github.com/thenewtricks/courage-service/tcp"
	"log"
)

type RealtimeClient struct {
	username   string
	password   string
	providerID uuid.UUID
	channelID  uuid.UUID
	deviceID   uuid.UUID
}

func (rc *RealtimeClient) ServeTCPConnect(mwch chan binmsg.MessageWriter) {
	// Create the subscribe request payload.
	subscribeRequest := &TCPSubscribeRequest{rc.username, rc.password, rc.providerID, rc.channelID, rc.deviceID, 0}

	// Get the message writer when it's ready.
	mw := <-mwch

	// Write the subscribe request header.
	err := mw.WriteHeader(TCPProtocolIDSubscribe, TCPMessageTypeSubscribeRequest)
	if err != nil {
		log.Println("failed to write message header:", err)
		return
	}

	// Write the subscribe request payload.
	err = WriteTCPSubscribeRequest(mw, subscribeRequest)
	if err != nil {
		log.Println("failed to write subscribe request:", err)
		return
	}

	// Flush the message writer to send the buffered message to the underlying connection.
	err = mw.Flush()
	if err != nil {
		log.Println("failed to flush message writer to connection:", err)
		return
	}

	log.Println("[OK] subscribed to channel:", rc.providerID, rc.channelID, rc.deviceID)

	mwch <- mw
}

func (rc *RealtimeClient) ServeTCP(mwch chan binmsg.MessageWriter, mr binmsg.MessageReader) {
	_, messageType := mr.Header()

	switch messageType {
	case TCPMessageTypeSubscribeOKResponse:
		ok, err := ReadTCPSubscribeOKResponse(mr)
		if err != nil {
			log.Println("failed to read ok response:", err)
			return
		}

		log.Println("[OK] ok response:", ok.ChannelID)

	case TCPMessageTypeSubscribeErrorResponse:
		e, err := ReadTCPSubscribeErrorResponse(mr)
		if err != nil {
			log.Println("failed to read error response:", err)
			return
		}

		log.Println("[ER] error response:", e.ChannelID, e.ErrorCode)

	case TCPMessageTypeSubscribeStreamingResponse:
		resp, err := ReadTCPSubscribeStreamingResponse(mr)
		if err != nil {
			log.Println("failed to read streaming reponse:", err)
			return
		}

		log.Println("[OK] streaming reponse:", resp.ChannelID, resp.EventData)

	default:
		log.Println("invalid message type:", messageType)
	}
}

func main() {
	providerID := uuid.Parse("791f011e-7d16-4614-9e2f-1b28db45b7b3")
	channelID := uuid.Parse("C0B83377-9F23-46D5-8F25-6E66F2B523CA")
	deviceID := uuid.Parse("B1E0C2CA-2A0A-4AAA-B489-BD8658FF89D4")

	rc := &RealtimeClient{"tylrtrmbl", "furball-calico", providerID, channelID, deviceID}

	peer := binmsg.NewPeer()
	peer.HandleConnect(rc)
	peer.Handle(TCPProtocolIDSubscribe, rc)

	err := tcp.Dial("localhost:2874", peer)
	if err != nil {
		log.Fatalln("failed to connect to localhost:2874")
		return
	}

	stayinAlive := make(chan interface{})
	<-stayinAlive
}
