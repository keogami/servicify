package main

import (
	"errors"
	"strings"

	"golang.org/x/sys/windows/svc/mgr"
)

var (
	// ErrUnknownStartType is returned when the user supplied start type is not recognized
	ErrUnknownStartType = errors.New("Servicify: The provided start type is not recognized")
	// ErrUnknownServiceType is returned when the user supplied service type is not recognized
	ErrUnknownServiceType = errors.New("Servicify: The provided service type is not recognized")
)

// ServiceType is the type of service
// Possible values: { own | share | userown | usershare | driver | filesys | interact }
type ServiceType string

// Value returns the uint32 for the given string alias of service type, the bool is false if the type is not recognized
func (sv ServiceType) Value() (value uint32, valid bool) {
	types := map[string]uint32{
		"own":       0x10,
		"share":     0x20,
		"userown":   0x50,
		"usershare": 0x60,
		"driver":    0x01,
		"filesys":   0x02,
		"interact":  0x100,
	}

	value, valid = types[strings.ToLower(string(sv))]
	return
}

// StartType defines how (and when) the service is started
// Possible values: { manual | auto | disabled | delayed-auto | boot | system }
type StartType string

// Value returns the uint32 for the given string alias of start type, the bool is false if the type is not recognized
func (st StartType) Value() (value uint32, valid bool) {
	types := map[string]uint32{
		"boot":         0x00,
		"system":       0x01,
		"auto":         0x02,
		"delayed-auto": 0x02,
		"manual":       0x03,
		"disabled":     0x04,
	}

	value, valid = types[strings.ToLower(string(st))]
	return
}

// IsDelayed returns whether the service starts delayed (start type must be delayed-auto)
func (st StartType) IsDelayed() bool {
	return strings.ToLower(string(st)) == "delayed-auto"
}

// Config to be used by the running servicing entity
type Config struct {
	Name        string      `json:"Name"`                  // Name of the service, not to be confused with `DisplayName`
	Image       string      `json:"Image"`                 // Image is the path to the binary to be run inside the service
	DependsOn   []string    `json:"DependsOn,omitempty"`   // DependsOn is a list of the Names (not display name) of services it depends on
	Description string      `json:"Description,omitempty"` // Description of the service
	AccountName string      `json:"AccountName,omitempty"` // AccountName under which the service is installed
	Password    string      `json:"Password,omitempty"`    // Password to the account under which the service is installed
	Options     []string    `json:"Options,omitempty"`     // Options to be passed to the binary that runs inside the service
	DisplayName string      `json:"DisplayName,omitempty"` // DisplayName is the name shown in the service.msc
	ServiceType ServiceType `json:"ServiceType,omitempty"`
	StartType   StartType   `json:"StartType,omitempty"`
}

// Mold converts user supplied config to SCM accepted config
func (c Config) Mold() (mc mgr.Config, err error) {
	sv, valid := c.ServiceType.Value()
	if !valid {
		err = ErrUnknownServiceType
		return
	}

	st, valid := c.StartType.Value()
	if !valid {
		err = ErrUnknownStartType
		return
	}

	mc = mgr.Config{
		ServiceType:      sv,
		StartType:        st,
		DelayedAutoStart: c.StartType.IsDelayed(),
		DisplayName:      c.DisplayName,
		Description:      c.Description,
		Dependencies:     c.DependsOn,
		ServiceStartName: c.AccountName,
		Password:         c.Password,
	}
	return
}
