package server

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Charana123/ftp/utils"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	// globalLock synchronises access to all resources shared between FTP connections
	globalLock = &sync.RWMutex{}
	// freeListenerPorts is a shared list of ports available to listen for data connections
	freeListenerPorts = make([]int, 0, 0)
)

// getFreeListenerPort synchronized go routine access when obtaining a free listener port
func getFreeListenerPort() int {
	globalLock.Lock()
	freeListenerPort := freeListenerPorts[len(freeListenerPorts)-1]
	freeListenerPorts = freeListenerPorts[:len(freeListenerPorts)-1]
	globalLock.Unlock()
	return freeListenerPort
}

// putFreeListenerPort synchronized go routine access when putting back
func putFreeListenerPort(port int) {
	globalLock.Lock()
	freeListenerPorts = append(freeListenerPorts, port)
	globalLock.Unlock()
}

// pasv handles a user 'PASV' control command
func (conn *ftpConnection) pasv() {

	// Transition to a `passive` mode from `active` mode
	if !conn.isPasv {
		var pasvDataListener net.Listener
		conn.isPasv = true
		for pasvDataListener == nil {
			port := getFreeListenerPort()
			conn.pasvPort = port
			pasvDataListener, _ = net.Listen("tcp4", ":"+strconv.Itoa(port))
		}
		conn.pasvDataListener = pasvDataListener
	}
	// Asynchronously listen to incoming data connections from the user.
	// Allowing the server to handle incoming control commands without blocking here.
	go func() {
		conn.data, _ = conn.pasvDataListener.Accept()
	}()

	// Depending on wether the incoming connection came from an IP address internal/external
	// to the LAN we return either our local IP or our global IP.
	// Allows hosts communicating over LAN without our FTP server being exposed publically
	ip := net.ParseIP(strings.Split(conn.control.RemoteAddr().String(), ":")[0])
	var pasvIP string
	if utils.IsPrivateIP(ip) {
		pasvIP = utils.GetLocalIP()
	} else {
		pasvIP = globalServerSettings.PublicIP
	}

	high := strconv.Itoa(conn.pasvPort / 256)
	low := strconv.Itoa(conn.pasvPort % 256)
	pasvAddr := "(" + strings.Join(strings.Split(pasvIP, "."), ",") + "," + high + "," + low + ")"
	conn.sendReply(227, "Entering Passive Mode "+pasvAddr)
}

// port handles a user 'PORT' control command
func (conn *ftpConnection) port(argument string) {
	// Transition from `passive` mode to `active` mode
	if conn.isPasv {
		conn.isPasv = false
		conn.pasvDataListener.Close()
		// Recycles the now unused passive data connection port
		putFreeListenerPort(conn.pasvPort)
	}

	// Parse and validate user end-point (TCP address)
	byteFields := strings.Split(argument, ",")
	high, err1 := strconv.Atoi(byteFields[4])
	low, err2 := strconv.Atoi(byteFields[5])
	port := high*256 + low
	activeAddr, err3 := net.ResolveTCPAddr("tcp4", strings.Join(byteFields[0:4], ".")+":"+strconv.Itoa(port))
	if notok := utils.HandleWarning(func() {
		conn.sendReply(501, "Syntax error in parameters or arguments.")
	}, err1, err2, err3); notok {
		return
	}
	conn.activeAddr = activeAddr
	conn.sendReply(200, "Command okay.")
}

// openDataConnection handles establishing the data connection subject to `passive`
// or `active` flag set by user
func (conn *ftpConnection) openDataConnection() error {
	if conn.isPasv {
		// Block until user sends connection request to server
		interval, _ := time.ParseDuration("1s")
		timeout, _ := time.ParseDuration("5s")
		wait.Poll(interval, timeout, wait.ConditionFunc(func() (bool, error) {
			return conn.data != nil, nil
		}))
		if conn.data == nil {
			return fmt.Errorf("User did not make connection before timeout")
		}
		conn.sendReply(125, "Data connection already open. Transfer starting.")
	} else {
		// Make connection request to user
		conn.sendReply(150, "File status okay; about to open data connection.")
		data, err := net.Dial("tcp4", conn.activeAddr.String())
		if err != nil {
			return err
		}
		conn.data = data
	}
	return nil
}
