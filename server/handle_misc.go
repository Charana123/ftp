package server

import (
	"strings"
	"time"

	"github.com/Charana123/ftp/utils"
	"k8s.io/apimachinery/pkg/util/wait"
)

// syst handles a suer 'SYST' control command
func (conn *ftpConnection) syst() {
	conn.sendReply(215, "UNIX Type: L8")
}

// ttype handles a user 'TYPE' control command
func (conn *ftpConnection) ttype(argument string) {
	// Only support IMAGE (binary) and ASCII data representations
	if strings.Compare(argument, "I") != 0 && strings.Compare(argument, "A") != 0 {
		conn.sendReply(504, "Command not implemented for that parameter.")
	} else {
		conn.sendReply(200, "Command okay.")
	}
}

// stru handles a user 'STRU' control command
func (conn *ftpConnection) stru(argument string) {
	// Only supports FILE structure
	if strings.Compare(argument, "F") != 0 {
		conn.sendReply(504, "Command not implemented for that parameter.")
	} else {
		conn.sendReply(200, "Command okay.")
	}
}

// mode handles a user 'MODEE' control command
func (conn *ftpConnection) mode(argument string) {
	// Only supports STREAM mode
	if strings.Compare(argument, "S") != 0 {
		conn.sendReply(504, "Command not implemented for that parameter.")
	} else {
		conn.sendReply(200, "Command okay.")
	}
}

// quit handles a user 'QUIT' control command
func (conn *ftpConnection) quit() {
	// Asynchronously block until all ongoing file transfers conclude then close the connection.
	// Performing this asynchronously allows the server to reply to subsequent (invalid) commands
	// with a 421 reply code.
	go func() {
		conn.closing = true
		if conn.ongoingFileTransfer {
			interval, err := time.ParseDuration("1s")
			utils.HandleFatalError(nil, err)
			wait.PollInfinite(interval, wait.ConditionFunc(func() (bool, error) {
				return !conn.ongoingFileTransfer, nil
			}))
		}
		conn.sendReply(221, "Closing Service closing control connection. Logged out if appropriate.")
		conn.control.Close()
	}()
}
