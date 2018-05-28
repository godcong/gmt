package gmt

import (
	"testing"
	"log"
	"time"
)

var nameList = []string{
	"1",
	"2",
	"3",
	"4",
	"5",
	"6",
	"7",
	"8",
	"9",
	"10",
	"11",
	"12",
}

type server struct {
	ServerName string
}

func (s *server) Run(obj *ThreadMain) error {
	log.Println("running: ", s.Name())
	//if s.Name() == "1" {
		obj.SendTo(s, "2", "hello world")
	//}
	return nil
}

func (s *server) Name() string {
	return s.ServerName
}

func (s *server) ReceiveFrom(name string, val interface{}) error {
	log.Println(s.Name(),"receive from ", name, val)
	return nil
}

func TestThreadMain_Start(t *testing.T) {
	//register you thread
	for _, v := range nameList {
		Register(&server{
			ServerName:v ,
		})
	}

	//run thread
	go Start()

	time.Sleep(time.Second * 3)

	// stop
	Stop()

}
