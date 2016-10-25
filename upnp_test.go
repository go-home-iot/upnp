package upnp_test

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-home-iot/upnp"
	"github.com/stretchr/testify/require"
)

type MockDevice struct {
	SubscribeCount   int
	SubscribeSID     string
	CallbackURL      string
	Timeout          string
	UnsubscribeCount int
	UnsubscribeSID   string
}

func (d *MockDevice) Start(addr string) {
	var sid int = 1
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		fmt.Printf("%+v\n", r)
		switch r.Method {
		case "SUBSCRIBE":
			d.SubscribeCount++

			if len(r.Header["Sid"]) > 0 {
				d.SubscribeSID = r.Header["Sid"][0]
			}
			d.Timeout = r.Header["Timeout"][0]
			d.CallbackURL = r.Header["Callback"][0]

			// If refresh SID is passed to the request
			w.Header().Set("SID", strconv.Itoa(sid))
			sid++
			w.Header().Set("TIMEOUT", r.Header["Timeout"][0])
			w.WriteHeader(http.StatusOK)

		case "UNSUBSCRIBE":
			d.UnsubscribeCount++
		}

		fmt.Println("mock device got request")
		fmt.Printf("%+v\n", r)

		/*
			HTTP/1.1 200 OK
			DATE: Sun, 11 Jan 2015 18:27:05 GMT
			SERVER: Unspecified, UPnP/1.0, Unspecified
			CONTENT-LENGTH: 0
			X-User-Agent: redsonic
			SID: uuid:7206f5ac-1dd2-11b2-80f3-e76de858414e
			TIMEOUT: Second-600*/
	})
	http.ListenAndServe(addr, mux)
}

func (d *MockDevice) RaiseEvent(e upnp.NotifyEvent) error {

	url := strings.TrimSuffix(strings.TrimPrefix(d.CallbackURL, "<"), ">")
	req, err := http.NewRequest("NOTIFY", url, strings.NewReader(e.Body))
	if err != nil {
		return err
	}

	req.Header.Add("NT", "upnp:event")
	req.Header.Add("NTS", "upnp:propchange")
	req.Header.Add("SID", e.SID)
	req.Header.Add("SEQ", "0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Non 200 return code %d", resp.StatusCode)
	}
	return nil
}

type MockSubscriber struct {
	NotifyCount int
	Event       upnp.NotifyEvent
}

func (s *MockSubscriber) UPNPNotify(e upnp.NotifyEvent) {
	s.NotifyCount++
	s.Event = e
}

func TestStart(t *testing.T) {
	s := upnp.NewSubServer()

	// Start the server, this will listen for upnp events
	go func() {
		err := s.Start("127.0.0.1:9001")
		fmt.Println("server stopped")
		fmt.Println(err)
	}()

	// Create a mock device that will simulate sending updates and allow
	// subscriptions
	d := &MockDevice{}
	go d.Start("127.0.0.1:9002")

	// Subscriber who will get the notification events
	sub := &MockSubscriber{}

	// Subscribe to the devices events
	sid, err := s.Subscribe(
		"http://127.0.0.1:9002/upnp/event/basicevent1",
		"",
		50,
		false,
		sub,
	)

	require.Nil(t, err)
	require.Equal(t, "1", sid)

	// Shouldn't have have a SID pass to subscribe the first time
	require.Equal(t, 1, d.SubscribeCount)
	require.Equal(t, "", d.SubscribeSID)
	require.Equal(t, "Second-50", d.Timeout)
	require.Equal(t, "<http://127.0.0.1:9001>", d.CallbackURL)
}

func TestSubscriberGetsNotifyEvents(t *testing.T) {
	s := upnp.NewSubServer()

	go func() {
		err := s.Start("127.0.0.1:9004")
		fmt.Println("server stopped")
		fmt.Println(err)
	}()

	d := &MockDevice{}
	go d.Start("127.0.0.1:9005")

	sub := &MockSubscriber{}

	// Subscribe to the devices events
	sid, err := s.Subscribe(
		"http://127.0.0.1:9005/upnp/event/basicevent1",
		"",
		50,
		false,
		sub,
	)
	require.Nil(t, err)
	require.Equal(t, "<http://127.0.0.1:9004>", d.CallbackURL)

	// Simulate the device raising a notify event, make sure that the
	// subscriber gets it
	time.Sleep(100 * time.Millisecond)
	evt := upnp.NotifyEvent{
		SID:  sid,
		Body: "this is a test",
	}
	err = d.RaiseEvent(evt)
	require.Nil(t, err)
	require.Equal(t, 1, sub.NotifyCount)
	require.Equal(t, evt, sub.Event)
}

func TestUnsubscribe(t *testing.T) {
	//TODO:
}

func TestRefreshSubscription(t *testing.T) {
	//TODO:
}

func TestAutoRefresh(t *testing.T) {
	//TODO:
}

//wemo
//		"http://192.168.0.34:49154/upnp/event/basicevent1",
