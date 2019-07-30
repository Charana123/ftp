package driver

import (
	"crypto/tls"
	"fmt"
	"sync/atomic"

	"github.com/Charana123/ftp/server"
	"github.com/Charana123/ftp/utils"
)

type ExampleDriver struct {
	MaxConnections int32
	numClients     int32
}

func (d *ExampleDriver) Welcome(ctx *server.UserContext) (string, error) {
	if atomic.AddInt32(&d.numClients, 1) > d.MaxConnections {
		return "", fmt.Errorf("Maximum number of connections %d reached", d.MaxConnections)
	}
	return fmt.Sprintf("Welcome"), nil
}

func (d *ExampleDriver) Bye(ctx *server.UserContext) (string, error) {
	atomic.AddInt32(&d.numClients, -1)
	return fmt.Sprintf("Bye"), nil
}

func (d *ExampleDriver) AuthUser(ctx *server.UserContext, user string, pass string) (bool, error) {
	// No authentication.
	// You would most likely parse a local configuration file here that contains a set of
	// privledged users and match the username and password provided here against them.
	return true, nil
}

func (d *ExampleDriver) GetAccessControlSettings() (map[string][]string, error) {
	accessControlSettings := make(map[string][]string)
	// Defines an access control rule to the PWD of the FTP server
	// for all authenticating users.
	accessControlSettings["all"] = []string{
		"/Users/charana/Documents/ftp-public/",
	}
	return accessControlSettings, nil
}

func (d *ExampleDriver) GetSettings() (*server.ServerSettings, error) {
	publicIP, err := utils.GetExternalIP()
	if err != nil {
		return nil, err
	}
	dataPortRange, err := server.NewPortRange(8000, 8050)
	if err != nil {
		return nil, err
	}
	return &server.ServerSettings{
		ListeningPort:   2121,
		PublicIP:        publicIP,
		PublicDirectory: "/Users/charana/Documents/ftp-public/",
		DataPortRange:   dataPortRange,
	}, nil
}

func (d *ExampleDriver) GetTLSConfig() (*tls.Config, error) {
	// Returning nil disables FTP over TLS
	// You would want to use the 'crypto/tls' package to load a tls certificate
	// with `LoadX509KeyPair`
	return nil, nil
}
