package gmt

import (
	"sync"
	"context"
	"time"
	"runtime"
)

type ThreadMain struct {
	DelayedSeconds int
	stop           bool
	threads        sync.Map
	transport      sync.Pool
}

type NameAble interface {
	Name() string
}

type RunAble interface {
	Run(obj *ThreadMain) error
}

type ProcessAble interface {
	ReceiveFrom(name string, val interface{}) error
}

type ThreadAble interface {
	NameAble
	RunAble
	ProcessAble
}

type Thread interface {
	NameAble
	ThreadRun(ctx context.Context)
	Stop()
}

type delayedTransmission struct {
	Name NameAble
	To   string
	Val  interface{}
}

var DefaultThread = NewThreadMain()

func NewThreadMain() *ThreadMain {
	return &ThreadMain{
		DelayedSeconds: 5,
	}
}

func Register(t ThreadAble) () {
	DefaultThread.Register(t)
}

func RegisterThread(t Thread) () {
	DefaultThread.RegisterThread(t)
}

func SendTo(self NameAble, name string, val interface{}) {
	DefaultThread.SendTo(self, name, val)
}

func DelayedSendTo(self NameAble, name string, val interface{}) {
	DefaultThread.DelayedSendTo(self, name, val)
}

//func Find(name string) (t ThreadAble) {
//	return defaultThread.Find(name)
//}

func Start() {
	go DefaultThread.Start()
}

func Stop() {
	DefaultThread.Stop()
}

func (obj *ThreadMain) Register(inter interface{}) {
	if nm, b := inter.(NameAble); b {
		obj.threads.Store(nm.Name(), inter)
		return
	}
	obj.threads.Store(GenerateRandomString(16), inter)
}

func (obj *ThreadMain) RegisterRun(t ThreadAble) {
	obj.Register(t)
}

func (obj *ThreadMain) RegisterThread(t Thread) {
	obj.Register(t)
}

func (obj *ThreadMain) DelayedSendTo(self NameAble, name string, val interface{}) {
	obj.transport.Put(&delayedTransmission{
		Name: self,
		To:   name,
		Val:  val,
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
	// config cpu number
	runtime.GOMAXPROCS(runtime.NumCPU())

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
				go v.ThreadRun(ctx)
				return true
			}
			//fmt.Println("not succeeded with ", key)
			return false
		})
	for {
		if obj.stop {
			break
		}
		for delayedSend(obj) {

		}
		time.Sleep(time.Second * time.Duration(obj.DelayedSeconds))

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
