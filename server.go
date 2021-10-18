package main

import (
	"github.com/lorenzotinfena/chat-and-meet/proto" // Update
)

type server struct {

}
func (*server) Match(*proto.MatchRequest, proto.Service_MatchServer) error {
	return nil
}
func (*server) StartChat(proto.Service_StartChatServer) error {
	return nil
}