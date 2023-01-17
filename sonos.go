package sonos

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sync"
)

type Sonos struct {
	udpListener *net.UDPConn
	tcpListener net.Listener

	zonePlayers sync.Map
}

type FoundZonePlayer func(*Sonos, *ZonePlayer)

func NewSonos() (*Sonos, error) {
	// Create listener for M-SEARCH
	udpListener, err := net.ListenUDP("udp", &net.UDPAddr{IP: []byte{0, 0, 0, 0}, Port: 0, Zone: ""})
	if err != nil {
		return nil, err
	}

	// create listener for events
	tcpListener, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}

	s := &Sonos{
		udpListener: udpListener,
		tcpListener: tcpListener,
	}

	go func() {
		http.Serve(s.tcpListener, s)
	}()

	return s, nil
}

func (s *Sonos) Close() {
	s.udpListener.Close()
	s.tcpListener.Close()
}

func (s *Sonos) Search(ctx context.Context, foundFn FoundZonePlayer) error {
	go func(ctx context.Context) {
		for {
			if ctx.Err() != nil {
				break
			}
			response, err := http.ReadResponse(bufio.NewReader(s.udpListener), nil)
			if err != nil {
				continue
			}

			location, err := url.Parse(response.Header.Get("Location"))
			if err != nil {
				continue
			}
			zp, err := NewZonePlayer(WithLocation(location))
			if err != nil {
				continue
			}
			if zp.IsCoordinator() {
				zp, loaded := s.zonePlayers.LoadOrStore(zp.SerialNum(), zp)
				if !loaded {
					foundFn(s, zp.(*ZonePlayer))
				}
			}
		}
	}(ctx)

	// https://svrooij.io/sonos-api-docs/sonos-communication.html#auto-discovery
	// MX should be set to use timeout value in integer seconds
	pkt := []byte("M-SEARCH * HTTP/1.1\r\nHOST: 239.255.255.250:1900\r\nMAN: \"ssdp:discover\"\r\nMX: 1\r\nST: urn:schemas-upnp-org:device:ZonePlayer:1\r\n\r\n")
	for _, bcastaddr := range []string{"239.255.255.250:1900", "255.255.255.255:1900"} {
		bcast, err := net.ResolveUDPAddr("udp", bcastaddr)
		if err != nil {
			return err
		}
		_, err = s.udpListener.WriteTo(pkt, bcast)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Sonos) Register(zp *ZonePlayer) error {
	if zp.IsCoordinator() {
		_, loaded := s.zonePlayers.LoadOrStore(zp.SerialNum(), zp)
		if loaded {
			return fmt.Errorf("ZonePlayer already registered")
		}
		return nil
	}
	return fmt.Errorf("ZonePlayer is not coordinator")
}

func (s *Sonos) Subscribe(ctx context.Context, zp *ZonePlayer, service SonosService) (string, error) {
	conn, err := net.Dial("tcp", service.EventEndpoint().Host)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	host := fmt.Sprintf("%s:%d", conn.LocalAddr().(*net.TCPAddr).IP.String(), s.tcpListener.Addr().(*net.TCPAddr).Port)

	calbackUrl := url.URL{
		Scheme:   "http",
		Host:     host,
		RawQuery: "sn=" + zp.SerialNum(),
		Path:     service.EventEndpoint().Path,
	}

	req, err := http.NewRequestWithContext(ctx, "SUBSCRIBE", service.EventEndpoint().String(), nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("HOST", service.EventEndpoint().Host)
	req.Header.Add("CALLBACK", "<"+calbackUrl.String()+">")
	req.Header.Add("NT", "upnp:event")
	req.Header.Add("TIMEOUT", "Second-300")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", errors.New(string(body))
	}

	return res.Header.Get("sid"), nil
}

func (s *Sonos) Renew(ctx context.Context, zp *ZonePlayer, service SonosService, sid string) error {
	req, err := http.NewRequestWithContext(ctx, "SUBSCRIBE", service.EventEndpoint().String(), nil)
	if err != nil {
		return err
	}

	req.Header.Add("HOST", service.EventEndpoint().Host)
	req.Header.Add("SID", sid)
	req.Header.Add("TIMEOUT", "Second-300")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return errors.New(string(body))
	}

	return nil
}
func (s *Sonos) Unsubscribe(ctx context.Context, zp *ZonePlayer, service SonosService, sid string) error {
	req, err := http.NewRequestWithContext(ctx, "UNSUBSCRIBE", service.EventEndpoint().String(), nil)
	if err != nil {
		return err
	}

	req.Header.Add("HOST", service.EventEndpoint().Host)
	req.Header.Add("SID", sid)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return errors.New(string(body))
	}

	return nil
}

func (s *Sonos) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	query := request.URL.Query()
	sn, ok := query["sn"]
	if !ok {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	p, ok := s.zonePlayers.Load(sn[0])
	if !ok {
		response.WriteHeader(http.StatusNotFound)
		return
	}
	zonePlayer := p.(*ZonePlayer)
	data, err := ioutil.ReadAll(request.Body)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	var events []interface{}

	if request.URL.Path == zonePlayer.AlarmClock.EventEndpoint().Path {
		events = zonePlayer.AlarmClock.ParseEvent(data)
	}
	if request.URL.Path == zonePlayer.AVTransport.EventEndpoint().Path {
		events = zonePlayer.AVTransport.ParseEvent(data)
	}
	if request.URL.Path == zonePlayer.ConnectionManager.EventEndpoint().Path {
		events = zonePlayer.ConnectionManager.ParseEvent(data)
	}
	if request.URL.Path == zonePlayer.ContentDirectory.EventEndpoint().Path {
		events = zonePlayer.ContentDirectory.ParseEvent(data)
	}
	if request.URL.Path == zonePlayer.DeviceProperties.EventEndpoint().Path {
		events = zonePlayer.DeviceProperties.ParseEvent(data)
	}
	if request.URL.Path == zonePlayer.GroupManagement.EventEndpoint().Path {
		events = zonePlayer.GroupManagement.ParseEvent(data)
	}
	if request.URL.Path == zonePlayer.GroupRenderingControl.EventEndpoint().Path {
		events = zonePlayer.GroupRenderingControl.ParseEvent(data)
	}
	if request.URL.Path == zonePlayer.MusicServices.EventEndpoint().Path {
		events = zonePlayer.MusicServices.ParseEvent(data)
	}
	if request.URL.Path == zonePlayer.Queue.EventEndpoint().Path {
		events = zonePlayer.Queue.ParseEvent(data)
	}
	if request.URL.Path == zonePlayer.RenderingControl.EventEndpoint().Path {
		events = zonePlayer.RenderingControl.ParseEvent(data)
	}
	if request.URL.Path == zonePlayer.SystemProperties.EventEndpoint().Path {
		events = zonePlayer.SystemProperties.ParseEvent(data)
	}
	if request.URL.Path == zonePlayer.VirtualLineIn.EventEndpoint().Path {
		events = zonePlayer.VirtualLineIn.ParseEvent(data)
	}
	if request.URL.Path == zonePlayer.ZoneGroupTopology.EventEndpoint().Path {
		events = zonePlayer.ZoneGroupTopology.ParseEvent(data)
	}

	for _, evt := range events {
		zonePlayer.Event(evt)
	}
	response.WriteHeader(http.StatusOK)
}

func (s *Sonos) FindRoom(ctx context.Context, room string) (*ZonePlayer, error) {
	c := make(chan *ZonePlayer)
	defer close(c)

	s.Search(ctx, func(s *Sonos, zp *ZonePlayer) {
		if zp.RoomName() == room {
			c <- zp
		}
	})

	for {
		select {
		case <-ctx.Done():
			return nil, errors.New("timeout")
		case zp := <-c:
			return zp, nil
		}
	}
}
