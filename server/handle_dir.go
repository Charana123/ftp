package server

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/Charana123/ftp/utils"
)

// cwd handles a user 'CWD' control command
func (conn *ftpConnection) cwd(args []string) {
	filePath, err := conn.resolvePath(args)
	if notok := utils.HandleWarning(func() {
		conn.sendReply(553, "Requested action not taken. File name not allowed.")
	}, err); notok {
		return
	}
	if !utils.IsDirectory(filePath) {
		conn.sendReply(550, "Requested action not taken. File unavailable.")
		return
	}
	conn.ctx.CWD = filePath
	conn.sendReply(250, "Requested file action okay, completed.")
}

// list handles a user 'LIST' control command
func (conn *ftpConnection) list() {
	go func() {
		conn.ongoingFileTransfer = true

		directoryPath := conn.ctx.CWD
		out, err := exec.Command("/bin/sh", "-c", "ls -l "+directoryPath).Output()
		if notok := utils.HandleWarning(func() {
			conn.sendReply(451, "Requested action aborted. Local error in processing.")
		}, err); notok {
			return
		}

		err = conn.openDataConnection()
		if notok := utils.HandleWarning(func() {
			conn.sendReply(425, "Can't open data connection.")
		}, err); notok {
			return
		}
		data := strings.Join(strings.Split(string(out), "\n"), "\r\n")
		fmt.Fprint(conn.data, data)
		conn.data.Close()
		conn.sendReply(226, "Closing data connection. Requested file action successful")

		conn.ongoingFileTransfer = false
	}()
}
