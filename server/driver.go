package server

import (
	"crypto/tls"
	"fmt"
)

type ServerDriver interface {
	// Function that executes and message to return whenever a user
	// connects to the FTP server
	Welcome(ctx *UserContext) (string, error)
	// Logic to execute and message to return whenever a user
	// disconnects to the FTP server
	Bye(ctx *UserContext) (string, error)
	// Logic authenticating the user
	AuthUser(ctx *UserContext, user string, pass string) (bool, error)
	// GetSettings returns a set of updated server configuration parameters
	GetSettings() (*ServerSettings, error)
	// GetAccessControlSettings returns a set of access control rules enforced on
	// most FTP service request
	GetAccessControlSettings() (map[string][]string, error)
	// GetTLSConfig returns a tls config with atleast one tls certificate
	GetTLSConfig() (*tls.Config, error)
}

type ServerSettings struct {
	// Public FTP directory whose access is unrestricted to authenticated users
	PublicDirectory string
	// Port listening to control connections
	ListeningPort int
	// Public Address of FTP host
	PublicIP string
	// Port range on which data connections will be established (PASV)
	DataPortRange *PortRange
}

type PortRange struct {
	start int
	end   int
}

func NewPortRange(start, end int) (*PortRange, error) {
	if start <= 1024 {
		return nil, fmt.Errorf("start port %d is a privledged port", start)
	} else if end < start {
		return nil, fmt.Errorf("end port %d, smaller than start port %d", end, start)
	} else {
		return &PortRange{
			start: start,
			end:   end,
		}, nil
	}
}
