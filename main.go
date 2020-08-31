package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
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
	MemUsage   int `json:"memUsage"`
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

func makeSh(cmd string) error {
	f, err := os.Create("execCmd.sh")
	if err != nil {
		return err
	}

	f.WriteString("#!/bin/bash\n")
	f.WriteString(cmd)

	f.Close()

	os.Chmod("execCmd.sh", 0777)

	return nil
}

func execCmd(request requestJSON) cmdResultJSON {
	var cmdResult cmdResultJSON
	cmdResult.SessionID = request.SessionID

	if err := makeSh(request.Cmd); err != nil {
		log.Println(err)
		return cmdResult
	}

	cmd := exec.Command("sh", "-c", "/usr/bin/time -v ./execCmd.sh 2>&1 | grep -E 'Maximum' | awk '{ print $6}' > mem_usage.txt")

	start := time.Now()
	timeout := time.After(2 * time.Second)

	if err := cmd.Start(); err != nil {
		cmdResult.ErrMessage = err.Error()
	}

	done := make(chan error)
	go func() { done <- cmd.Wait() }()

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

	memUsage, err := getMemUsage()
	if err != nil {
		log.Println(err)
	}
	cmdResult.MemUsage = memUsage

	cmdResult.Result = cmd.ProcessState.ExitCode() == 0

	return cmdResult
}

func getMemUsage() (int, error) {
	fp, err := os.Open("mem_usage.txt")
	if err != nil {
		return 0, err
	}
	defer fp.Close()

	buf, err := ioutil.ReadAll(fp)

	tmp := strings.Replace(string(buf), "\n", "", -1)

	mem, err := strconv.Atoi(tmp)
	if err != nil {
		return 0, err
	}

	return mem, err
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
