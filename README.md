# upnp
upnp library for go

##Documentation
See [godoc](https://godoc.org/github.com/go-home-iot/upnp)

##Installation
```bash
go get github.com/go-home-iot/upnp
```

##Package
```go
import "github.com/go-home-iot/upnp"
```

##Usage
See upnp_test.go for detailed examples of how to use this library.

```go
server := upnp.NewSubServer()

// Start the server, this will listen for upnp events
go func() {
	err := s.Start("127.0.0.1:9001")
	fmt.Println(err)
}()

// Subscriber who will get the notification events
sub := &MockSubscriber{}

// Subscribe to the devices events
sid, err := s.Subscribe(
	"http://127.0.0.1:9002/upnp/event/basicevent1",  //URL of devices event service
	"", //SID - pass in to renew a subscription
	50, //Time in seconds to maintain the subscription
	false, // Autorenew subscription
	sub, // Instance who will receive notifications, implementing upnp.Subscriber interface
)

type MockSubscriber struct {
	NotifyCount int
	Event       upnp.NotifyEvent
}

func (s *MockSubscriber) UPNPNotify(e upnp.NotifyEvent) {
	s.NotifyCount++
	s.Event = e
}
```

##Version History
###0.1.0
Initial release - support SUBSCRIBE/UNSUBSCRIBE/NOTIFY for devices
