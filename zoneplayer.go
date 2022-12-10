package sonos

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	avt "github.com/caglar10ur/sonos/services/AVTransport"
	clk "github.com/caglar10ur/sonos/services/AlarmClock"
	ain "github.com/caglar10ur/sonos/services/AudioIn"
	con "github.com/caglar10ur/sonos/services/ConnectionManager"
	dir "github.com/caglar10ur/sonos/services/ContentDirectory"
	dev "github.com/caglar10ur/sonos/services/DeviceProperties"
	gmn "github.com/caglar10ur/sonos/services/GroupManagement"
	rcg "github.com/caglar10ur/sonos/services/GroupRenderingControl"
	mus "github.com/caglar10ur/sonos/services/MusicServices"
	ply "github.com/caglar10ur/sonos/services/QPlay"
	que "github.com/caglar10ur/sonos/services/Queue"
	ren "github.com/caglar10ur/sonos/services/RenderingControl"
	sys "github.com/caglar10ur/sonos/services/SystemProperties"
	vli "github.com/caglar10ur/sonos/services/VirtualLineIn"
	zgt "github.com/caglar10ur/sonos/services/ZoneGroupTopology"
)

type SonosService interface {
	ControlEndpoint() *url.URL
	EventEndpoint() *url.URL
	ParseEvent([]byte) []interface{}
}

type SpecVersion struct {
	XMLName xml.Name `xml:"specVersion"`
	Major   int      `xml:"major"`
	Minor   int      `xml:"minor"`
}

type Service struct {
	XMLName     xml.Name `xml:"service"`
	ServiceType string   `xml:"serviceType"`
	ServiceID   string   `xml:"serviceId"`
	ControlURL  string   `xml:"controlURL"`
	EventSubURL string   `xml:"eventSubURL"`
	SCPDURL     string   `xml:"SCPDURL"`
}

type Icon struct {
	XMLName  xml.Name `xml:"icon"`
	ID       string   `xml:"id"`
	Mimetype string   `xml:"mimetype"`
	Width    int      `xml:"width"`
	Height   int      `xml:"height"`
	Depth    int      `xml:"depth"`
	URL      url.URL  `xml:"url"`
}

type Device struct {
	XMLName                 xml.Name  `xml:"device"`
	DeviceType              string    `xml:"deviceType"`
	FriendlyName            string    `xml:"friendlyName"`
	Manufacturer            string    `xml:"manufacturer"`
	ManufacturerURL         string    `xml:"manufacturerURL"`
	ModelNumber             string    `xml:"modelNumber"`
	ModelDescription        string    `xml:"modelDescription"`
	ModelName               string    `xml:"modelName"`
	ModelURL                string    `xml:"modelURL"`
	SoftwareVersion         string    `xml:"softwareVersion"`
	SwGen                   string    `xml:"swGen"`
	HardwareVersion         string    `xml:"hardwareVersion"`
	SerialNum               string    `xml:"serialNum"`
	MACAddress              string    `xml:"MACAddress"`
	UDN                     string    `xml:"UDN"`
	Icons                   []Icon    `xml:"iconList>icon"`
	MinCompatibleVersion    string    `xml:"minCompatibleVersion"`
	LegacyCompatibleVersion string    `xml:"legacyCompatibleVersion"`
	APIVersion              string    `xml:"apiVersion"`
	MinAPIVersion           string    `xml:"minApiVersion"`
	DisplayVersion          string    `xml:"displayVersion"`
	ExtraVersion            string    `xml:"extraVersion"`
	RoomName                string    `xml:"roomName"`
	DisplayName             string    `xml:"displayName"`
	ZoneType                int       `xml:"zoneType"`
	Feature1                string    `xml:"feature1"`
	Feature2                string    `xml:"feature2"`
	Feature3                string    `xml:"feature3"`
	Seriesid                string    `xml:"seriesid"`
	Variant                 int       `xml:"variant"`
	InternalSpeakerSize     float32   `xml:"internalSpeakerSize"`
	BassExtension           float32   `xml:"bassExtension"`
	SatGainOffset           float32   `xml:"satGainOffset"`
	Memory                  int       `xml:"memory"`
	Flash                   int       `xml:"flash"`
	FlashRepartitioned      int       `xml:"flashRepartitioned"`
	AmpOnTime               int       `xml:"ampOnTime"`
	RetailMode              int       `xml:"retailMode"`
	Services                []Service `xml:"serviceList>service"`
	Devices                 []Device  `xml:"deviceList>device"`
}

type Root struct {
	XMLName     xml.Name    `xml:"root"`
	Xmlns       string      `xml:"xmlns,attr"`
	SpecVersion SpecVersion `xml:"specVersion"`
	Device      Device      `xml:"device"`
}

type ZonePlayerOption func(*ZonePlayer)

func WithClient(c *http.Client) ZonePlayerOption {
	return func(z *ZonePlayer) {
		z.client = c
	}
}

func WithLocation(u *url.URL) ZonePlayerOption {
	return func(z *ZonePlayer) {
		z.location = u
	}
}

func FromEndpoint(endpoint string) (*url.URL, error) {
	return url.Parse(fmt.Sprintf("http://%s:1400/xml/device_description.xml", endpoint))
}

type ZonePlayer struct {
	Root *Root

	client *http.Client
	// A URL that can be queried for device capabilities
	location *url.URL

	*Services
}

type Services struct {
	// services
	AlarmClock            *clk.Service
	AudioIn               *ain.Service
	AVTransport           *avt.Service
	ConnectionManager     *con.Service
	ContentDirectory      *dir.Service
	DeviceProperties      *dev.Service
	GroupManagement       *gmn.Service
	GroupRenderingControl *rcg.Service
	MusicServices         *mus.Service
	QPlay                 *ply.Service
	Queue                 *que.Service
	RenderingControl      *ren.Service
	SystemProperties      *sys.Service
	VirtualLineIn         *vli.Service
	ZoneGroupTopology     *zgt.Service
}

// NewZonePlayer returns a new ZonePlayer instance.
func NewZonePlayer(opts ...ZonePlayerOption) (*ZonePlayer, error) {
	zp := &ZonePlayer{
		Root: &Root{},
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	// Loop through each option
	for _, opt := range opts {
		// Call the option giving the instantiated *ZonePlayer as the argument
		opt(zp)
	}

	if zp.location == nil {
		return nil, fmt.Errorf("Empty location")
	}

	resp, err := zp.client.Get(zp.location.String())
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = xml.Unmarshal(body, zp.Root)
	if err != nil {
		return nil, err
	}

	zp.Services = &Services{
		AlarmClock: clk.NewService(
			clk.WithLocation(zp.location),
			clk.WithClient(zp.client),
		),
		AVTransport: avt.NewService(
			avt.WithLocation(zp.location),
			avt.WithClient(zp.client),
		),
		AudioIn: ain.NewService(
			ain.WithLocation(zp.location),
			ain.WithClient(zp.client),
		),
		ConnectionManager: con.NewService(
			con.WithLocation(zp.location),
			con.WithClient(zp.client),
		),
		ContentDirectory: dir.NewService(
			dir.WithLocation(zp.location),
			dir.WithClient(zp.client),
		),
		DeviceProperties: dev.NewService(
			dev.WithLocation(zp.location),
			dev.WithClient(zp.client),
		),
		GroupManagement: gmn.NewService(
			gmn.WithLocation(zp.location),
			gmn.WithClient(zp.client),
		),
		GroupRenderingControl: rcg.NewService(
			rcg.WithLocation(zp.location),
			rcg.WithClient(zp.client),
		),
		MusicServices: mus.NewService(
			mus.WithLocation(zp.location),
			mus.WithClient(zp.client),
		),
		QPlay: ply.NewService(
			ply.WithLocation(zp.location),
			ply.WithClient(zp.client),
		),
		Queue: que.NewService(
			que.WithLocation(zp.location),
			que.WithClient(zp.client),
		),
		RenderingControl: ren.NewService(
			ren.WithLocation(zp.location),
			ren.WithClient(zp.client),
		),
		SystemProperties: sys.NewService(
			sys.WithLocation(zp.location),
			sys.WithClient(zp.client),
		),
		VirtualLineIn: vli.NewService(
			vli.WithLocation(zp.location),
			vli.WithClient(zp.client),
		),
		ZoneGroupTopology: zgt.NewService(
			zgt.WithLocation(zp.location),
			zgt.WithClient(zp.client),
		),
	}

	return zp, nil
}

// Client returns the underlying http client.
func (z *ZonePlayer) Client() *http.Client {
	return z.client
}

func (z *ZonePlayer) Location() *url.URL {
	return z.location
}

func (z *ZonePlayer) RoomName() string {
	return z.Root.Device.RoomName
}

func (z *ZonePlayer) ModelName() string {
	return z.Root.Device.ModelName
}

func (z *ZonePlayer) HardwareVersion() string {
	return z.Root.Device.HardwareVersion
}

func (z *ZonePlayer) SerialNum() string {
	return z.Root.Device.SerialNum
}

func (z *ZonePlayer) IsCoordinator() bool {
	zoneGroupState, err := z.GetZoneGroupState()
	if err != nil {
		return false
	}
	for _, group := range zoneGroupState.ZoneGroups {
		if "uuid:"+group.Coordinator == z.Root.Device.UDN {
			return true
		}
	}

	return false
}

func (z *ZonePlayer) GetZoneGroupState() (*ZoneGroupState, error) {
	zoneGroupStateResponse, err := z.ZoneGroupTopology.GetZoneGroupState(&zgt.GetZoneGroupStateArgs{})
	if err != nil {
		return nil, err
	}
	var zoneGroupState ZoneGroupState
	err = xml.Unmarshal([]byte(zoneGroupStateResponse.ZoneGroupState), &zoneGroupState)
	if err != nil {
		return nil, err
	}

	return &zoneGroupState, nil
}

func (z *ZonePlayer) GetVolume() (int, error) {
	res, err := z.RenderingControl.GetVolume(&ren.GetVolumeArgs{Channel: "Master"})
	if err != nil {
		return 0, err
	}

	return int(res.CurrentVolume), err
}

func (z *ZonePlayer) SetVolume(desiredVolume int) error {
	_, err := z.RenderingControl.SetVolume(&ren.SetVolumeArgs{
		Channel:       "Master",
		DesiredVolume: uint16(desiredVolume),
	})
	return err
}

func (z *ZonePlayer) Play() error {
	_, err := z.AVTransport.Play(&avt.PlayArgs{
		Speed: "1.0",
	})
	return err
}

func (z *ZonePlayer) Stop() error {
	_, err := z.AVTransport.Stop(&avt.StopArgs{})
	return err
}

func (z *ZonePlayer) SetAVTransportURI(url string) error {
	_, err := z.AVTransport.SetAVTransportURI(&avt.SetAVTransportURIArgs{
		CurrentURI: url,
	})
	return err
}

func (zp *ZonePlayer) Event(evt interface{}) {
	switch e := evt.(type) {
	default:
		fmt.Printf("Unhandeld event %T: %q\n", e, e)
	}
}
