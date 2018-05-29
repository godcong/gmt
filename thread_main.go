package gmt

import (
	"sync"
	"context"
	"time"

	"github.com/godcong/wego/log"
)

type ThreadMain struct {
	stop      bool
	threads   sync.Map
	transport sync.Pool
}

type NameAble interface {
	Name() string
}

type ThreadAble interface {
	Run(obj *ThreadMain) error
	Name() string
	ReceiveFrom(name string, val interface{}) error
}

type Thread interface {
	RunServer(ctx context.Context)
	Name() string
	Stop()
}

type delayedTransmission struct {
	Name NameAble
	To     string
	Val    interface{}
}

var defaultThread = ThreadMain{}

func NewThreadMain() *ThreadMain {
	return &ThreadMain{}
}

func Register(t ThreadAble) () {
	defaultThread.Register(t)
}

func RegisterThread(t Thread) () {
	defaultThread.RegisterThread(t)
}

func SendTo(self NameAble, name string, val interface{}) {
	defaultThread.SendTo(self, name, val)
}

func DelayedSendTo(self NameAble, name string, val interface{}){
	defaultThread.DelayedSendTo(self,name,val)
}

//func Find(name string) (t ThreadAble) {
//	return defaultThread.Find(name)
//}

func Start() {
	go defaultThread.Start()
}

func Stop() {
	defaultThread.Stop()
}

func (obj *ThreadMain) Register(t ThreadAble) {
	obj.threads.Store(t.Name(), t)
}

func (obj *ThreadMain) DelayedSendTo(self NameAble, name string, val interface{}) {
	obj.transport.Put(&delayedTransmission{
		Name: self,
		To:     name,
		Val:    val,
	})
}

func (obj *ThreadMain) SendTo(self NameAble, name string, val interface{}) {
	obj.threads.Range(
		func(key, value interface{}) bool {
			if v, b := value.(ThreadAble); b {
				if key == name {
					v.ReceiveFrom(self.Name(), val)
					return false
				}
			}
			return true
		})
}

func (obj *ThreadMain) find(name string) (t ThreadAble) {
	if v, b := obj.threads.Load(name); b {
		if v0, b := v.(ThreadAble); b {
			return v0
		}
	}
	return nil
}

func (obj *ThreadMain) threadRun(ctx context.Context, able ThreadAble) {
	for {
		select {
		case <-ctx.Done():
		default:
			if err := able.Run(obj); err != nil {
				return
			}
		}
	}
}

func (obj *ThreadMain) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	obj.threads.Range(
		func(key, value interface{}) bool {
			//run threadable
			if v, b := value.(ThreadAble); b {
				go obj.threadRun(ctx, v)
				return true
			}
			//run thread
			if v, b := value.(Thread); b {
				go v.RunServer(ctx)
				return true
			}
			log.Error("not succeeded with ", key)
			return false
		})
	for {
		if obj.stop {
			break
		}
		for delayedSend(obj) {

		}
		time.Sleep(time.Second)

	}

	//stop all threads
	obj.threads.Range(
		func(key, value interface{}) bool {
			if v, b := value.(Thread); b {
				v.Stop()
			}
			return true
		})

}

func delayedSend(obj *ThreadMain) bool {
	if v := obj.transport.Get(); v != nil {
		if v0, b := v.(*delayedTransmission); b {
			obj.SendTo(v0.Name, v0.To, v0.Val)
			return true
		}
	}
	return false
}

func (obj *ThreadMain) Stop() {
	obj.stop = true
}

func (obj *ThreadMain) RegisterThread(thread Thread) {
	obj.threads.Store(thread.Name(), thread)
}
