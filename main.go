package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/struCoder/pidusage"
)

const (
	ContainerPort = "0.0.0.0:8887"
	HostPort      = "172.17.0.1:3344"
)

type cmdResultJSON struct {
	SessionID  string  `json:"sessionID"`
	Result     bool    `json:"result"`
	ErrMessage string  `json:"errMessage"`
	Time       int64   `json:"time"`
	MemUsage   float64 `json:"memUsage"`
}

type requestJSON struct {
	SessionID string `json:"sessionID"`
	DirName   string `json:"dirName"`
	Cmd       string `json:"cmd"`
	Mode      string `json:"mode"` //Mode ... "judge" or "other"
}

func main() {
	listen, err := net.Listen("tcp", ContainerPort) //from backend server
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}

	for {
		cnct, err := listen.Accept()
		if err != nil {
			continue //continue to receive request
		}
		defer cnct.Close()

		var request requestJSON

		json.NewDecoder(cnct).Decode(&request)

		go func() {
			cmdResult := execCmd(request)
			
			getErrorDetails(&cmdResult)

			conn, err := net.Dial("tcp", HostPort)
			if err != nil {
				conn.Write([]byte("tcp connect error"))
			}
			defer conn.Close()

			b, err := json.Marshal(cmdResult)
			if err != nil {
				conn.Write([]byte("marshal error"))
			}

			conn.Write(b)
		}()
	}
}

func execCmd(request requestJSON) cmdResultJSON {
	var cmdResult cmdResultJSON
	cmdResult.SessionID = request.SessionID

	cmd := exec.Command("sh", "-c", request.Cmd)

	start := time.Now()

	err := cmd.Start()

	if err != nil {
		cmdResult.ErrMessage = err.Error()
	}

	info, _ := pidusage.GetStat(cmd.Process.Pid)
	cmdResult.MemUsage = info.Memory / 1024.0

	done := make(chan error)

	go func() { done <- cmd.Wait() }()

	timeout := time.After(2 * time.Second)

	select {
	case <-timeout:
		// Timeout happened first, kill the process and print a message.
		cmd.Process.Kill()
	case err := <-done:
		if err != nil {
			cmdResult.ErrMessage = err.Error()
		}
	}

	end := time.Now()

	cmdResult.Time = (end.Sub(start)).Milliseconds()

	if cmd.ProcessState.ExitCode() != 0 {
		cmdResult.Result = false
	} else {
		cmdResult.Result = true
	}

	return cmdResult
}

func getErrorDetails(cmdResult *cmdResultJSON) {
	stderrFp, err := os.Open("/userStderr.txt")
	if err != nil {
		cmdResult.ErrMessage = err.Error()
		return
	}

	buf := make([]byte, 65536)

	buf, err = ioutil.ReadAll(stderrFp)
	if err != nil {
		cmdResult.ErrMessage = err.Error()	
		return
	}

	cmdResult.ErrMessage = base64.StdEncoding.EncodeToString(buf) + "\n"

	stderrFp.Close()
}
