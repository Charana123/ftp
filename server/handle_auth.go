package server

import (
	"strings"
)

// user handles a user 'USER' control command
func (conn *ftpConnection) user(args []string) {
	if len(args) == 0 {
		conn.sendReply(500, "Syntax error, command unrecognized.")
		return
	}
	conn.ctx.User = args[0]
	conn.sendReply(331, "User name okay, need password.")
}

// pass handles a user 'PASS' control command
func (conn *ftpConnection) pass(args []string) {
	if len(args) == 0 {
		conn.sendReply(500, "Syntax error, command unrecognized.")
		return
	}
	// Handling when the password is sent before the username
	if conn.ctx.User == "" {
		conn.sendReply(503, "Bad sequence of commands.")
	}
	success, err := globalDriver.AuthUser(conn.ctx, conn.ctx.User, args[0])
	if err != nil || !success {
		conn.ctx.User = ""
		conn.sendReply(530, "Not logged in.")
	} else {
		conn.loggedIn = true
		conn.sendReply(230, "User logged in, proceed.")
	}
}

// checkAccessControl checks global and per user access control permissions
// to allow an FTP service command
func (conn *ftpConnection) checkAccessControl(path string) bool {
	if all, ok := globalAccessControlSettings["all"]; ok {
		for _, p := range all {
			// If `p` is a path to a directory
			if p[len(p)-1] == '/' {
				// Check if `path` IS the directory or is a file (or directory) IN that directory
				if strings.HasPrefix(path, p) || strings.Compare(path, p[:len(p)-1]) == 0 {
					return true
				}
			} else {
				// Check if `path` IS that file
				if strings.Compare(path, p) == 0 {
					return true
				}
			}
		}
	}
	return false
}
