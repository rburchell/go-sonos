package sonos

import (
	"encoding/xml"

	"github.com/ianr0bkny/go-sonos/didl"
)

// Lite embeds didl.Lite struct.
type Lite struct {
	*didl.Lite
}

// NewDIDL returns a new Lite instance.
func NewDIDL() *Lite {
	return &Lite{Lite: &didl.Lite{}}
}

// ParseDIDL converts given raw string into Lite struct or otherwise returns an error.
func ParseDIDL(raw string) (*Lite, error) {
	didl := NewDIDL()
	if err := xml.Unmarshal([]byte(raw), &didl); err != nil {
		return nil, err
	}
	return didl, nil

}
