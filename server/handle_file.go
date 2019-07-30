package server

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"

	"github.com/Charana123/ftp/utils"
)

// resolvePath constructs the absolute path using the current working directory of the
// FTP connection and the arguments provided to the corresponding FTP service command
func (conn *ftpConnection) resolvePath(args []string) (string, error) {
	var filePath string
	if len(args) == 0 {
		filePath = conn.ctx.CWD
	} else {
		filePath = args[0]
		if filePath[0] != '/' {
			filePath = path.Clean(path.Join(conn.ctx.CWD, filePath))
		}
	}
	if !conn.checkAccessControl(filePath) {
		return "", fmt.Errorf("Access not allowed")
	}
	return filePath, nil
}

// retr handles a user 'RETR' control command
func (conn *ftpConnection) retr(args []string) {
	go func() {
		conn.ongoingFileTransfer = true

		filePath, err := conn.resolvePath(args)
		if notok := utils.HandleWarning(func() {
			conn.sendReply(553, "Requested action not taken. File name not allowed.")
		}, err); notok {
			return
		}

		if ok := utils.IsFile(filePath); ok {
			file, err := os.Open(filePath)
			defer file.Close()
			if notok := utils.HandleWarning(func() {
				conn.sendReply(550, "Requested action not taken. File unavailable.")
			}, err); notok {
				return
			}

			err = conn.openDataConnection()
			if notok := utils.HandleWarning(func() {
				conn.sendReply(425, "Can't open data connection.")
			}, err); notok {
				return
			}

			io.Copy(conn.data, file)
			conn.data.Close()
			conn.sendReply(226, "Closing data connection. Requested file action successful")
		} else {
			utils.HandleWarning(func() {
				conn.sendReply(550, "Requested action not taken. File unavailable.")
			}, errors.New("Argument isn't file"))
		}
		conn.ongoingFileTransfer = false
	}()
}

// stor handles a user 'STOR' control command
func (conn *ftpConnection) stor(args []string) {
	go func() {
		conn.ongoingFileTransfer = true
		filePath, err := conn.resolvePath(args)
		if notok := utils.HandleWarning(func() {
			conn.sendReply(553, "Requested action not taken. File name not allowed.")
		}, err); notok {
			return
		}

		// Creates or Overwrites the specified file
		file, err := os.Create(filePath)
		defer file.Close()
		if notok := utils.HandleWarning(func() {
			conn.sendReply(550, "Requested action not taken. File unavailable.")
		}, err); notok {
			return
		}

		err = conn.openDataConnection()
		if notok := utils.HandleWarning(func() {
			conn.sendReply(425, "Can't open data connection.")
		}, err); notok {
			return
		}

		io.Copy(file, conn.data)
		conn.data.Close()
		conn.sendReply(226, "Closing data connection. Requested file action successful")
		conn.ongoingFileTransfer = false
	}()
}

// size handles a user 'SIZE' control command
func (conn *ftpConnection) size(args []string) {
	filePath, err := conn.resolvePath(args)
	if notok := utils.HandleWarning(func() {
		conn.sendReply(553, "Requested action not taken. File name not allowed.")
	}, err); notok {
		return
	}

	// Check if the specified file or directory exists
	if exists := utils.Exists(filePath); !exists {
		conn.sendReply(550, "Requested action not taken. File unavailable.")
		return
	}
	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		cmd = exec.Command("/bin/sh", "-c", "find "+filePath+" -type f -print0 | xargs -0 stat -f%z | awk '{b+=$1} END {print b}'")
	}
	if runtime.GOOS == "linux" {
		cmd = exec.Command("/bin/sh", "-c", "du -sb "+filePath)
	}
	out, err := cmd.Output()
	if notok := utils.HandleWarning(func() {
		conn.sendReply(451, "Requested action aborted. Local error in processing.")
	}, err); notok {
		return
	}
	conn.sendReply(213, strings.Trim(string(out), "\r\n"))

}

// TODO
func (conn *ftpConnection) mdtm(args []string) {
	filePath := args[0]
	filePath = path.Join(conn.ctx.CWD, filePath)

	if exists := utils.Exists(filePath); !exists {
		conn.sendReply(550, "Requested action not taken. File unavailable (e.g., file not found, no access).")
		return
	}
	var cmd *exec.Cmd
	// TODO - Should be in GMT (not seconds since epoch)
	if runtime.GOOS == "darwin" {
		cmd = exec.Command("/bin/sh", "-c", "stat -f\"%m\" "+filePath)
	}
	if runtime.GOOS == "linux" {
		cmd = exec.Command("/bin/sh", "-c", "stat --printf \"%Y\" "+filePath)
	}
	out, err := cmd.Output()
	if notok := utils.HandleWarning(func() {
		conn.sendReply(451, "Requested action aborted. Local error in processing.")
	}, err); notok {
		return
	}
	conn.sendReply(213, strings.Trim(string(out), "\r\n"))
}
