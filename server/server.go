// Library to build and configure your own FTP server
package server

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/Charana123/ftp/utils"
	"github.com/ziutek/telnet"
)

var (
	// driver is the server configuration object supplied by the user
	globalDriver ServerDriver
	// serverSettings updates the server parameter defaults
	globalServerSettings *ServerSettings
	// accessControlSettings is a set of user defined access control rules
	globalAccessControlSettings map[string][]string
)

type ftpConnection struct {
	control             net.Conn
	data                net.Conn
	ctx                 *UserContext
	ongoingFileTransfer bool
	loggedIn            bool
	closing             bool
	isPasv              bool
	pasvDataListener    net.Listener
	pasvPort            int
	activeAddr          net.Addr
}

// sendReply is a convnience method to corrently format and sends an FTP reply code
// followed by an appropriate human readable message.
func (conn *ftpConnection) sendReply(replyCode int, description string) {
	message := strconv.Itoa(replyCode) + " " + description + "\r\n"
	log.Print("server to (" + conn.control.RemoteAddr().String() + "): " + message)
	fmt.Fprintf(conn.control, message)
}

func (conn *ftpConnection) handleCommand(command string, arguments []string) bool {
	if command != "USER" || command != "PASS" || conn.loggedIn {
		if conn.closing {
			conn.sendReply(421, "Service not available, closing control connection.")
		} else {
			switch command {
			// Handle Authentication
			case "USER":
				conn.user(arguments)
			case "PASS":
				conn.pass(arguments)
			// Handle Connection
			case "PORT":
				conn.port(arguments[0])
			case "PASV":
				conn.pasv()
			// Handle Directory
			case "CWD":
				conn.cwd(arguments)
			case "LIST":
				conn.list()
			case "PWD":
				conn.sendReply(257, conn.ctx.CWD)
			// Handle File
			case "SIZE":
				conn.size(arguments)
			case "MDTM":
				conn.mdtm(arguments)
			case "RETR":
				conn.retr(arguments)
			case "STOR":
				conn.stor(arguments)
			// Handle Micellenous
			case "TYPE":
				conn.ttype(arguments[0])
			case "STRU":
				conn.stru(arguments[0])
			case "MODE":
				conn.mode(arguments[0])
			case "SYST":
				conn.syst()
			case "REIN":
				conn.ctx = newUserContext()
				conn.sendReply(200, "Command Okay.")
			case "NOOP":
				conn.sendReply(200, "Command okay.")
			case "QUIT":
				conn.quit()
				return true
			default:
				conn.sendReply(502, "Command not implemented.")
			}
		}
	}
	conn.sendReply(530, "Not logged in.")
	return false
}

func handleFTPConnection(conn *ftpConnection) {
	welcome, err := globalDriver.Welcome(conn.ctx)
	if err != nil {
		conn.sendReply(500, "Syntax error, command unrecognized.")
		return
	}
	conn.sendReply(220, welcome)

	r := bufio.NewReader(conn.control)
	for {
		command, err := r.ReadString('\n')
		command = strings.TrimRight(command, " \r\n")
		log.Println("user (" + conn.control.RemoteAddr().String() + "): " + command)
		if notok := utils.HandleWarning(func() {
			log.Println("Terminating connection with host: " + conn.control.RemoteAddr().String())
			conn.control.Close()
		}, err); notok {
			return
		}
		args := strings.Split(command, " ")
		comm := args[0]
		args = args[1:]
		if quit := conn.handleCommand(comm, args); quit {
			globalDriver.Bye(conn.ctx)
			return
		}
	}
}

// StartServer creates FTP server and configures it according to the supplied driver
func StartServer(driver ServerDriver) {
	globalDriver = driver

	var err error
	globalServerSettings, err = driver.GetSettings()
	utils.HandleFatalError(nil, err)
	for i := globalServerSettings.DataPortRange.start; i <= globalServerSettings.DataPortRange.end; i++ {
		freeListenerPorts = append(freeListenerPorts, i)
	}

	globalAccessControlSettings, err = driver.GetAccessControlSettings()
	utils.HandleFatalError(nil, err)

	listener, err := net.Listen("tcp4", ":"+strconv.Itoa(globalServerSettings.ListeningPort))
	log.Println("Starting server ... ")
	utils.HandleFatalError(nil, err)

	tlsConfig, err := driver.GetTLSConfig()
	utils.HandleWarning(nil, err)
	if tlsConfig != nil {
		listener = tls.NewListener(listener, tlsConfig)
	}

	for {
		con, err := listener.Accept()
		utils.HandleFatalError(nil, err)
		control, err := telnet.NewConn(con)
		utils.HandleFatalError(nil, err)
		conn := &ftpConnection{
			control: control,
			ctx:     newUserContext(),
		}
		go handleFTPConnection(conn)
	}
}
