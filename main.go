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

	"github.com/cafecoder-dev/cafecoder-judge/src/types"
)

const (
	ContainerPort = "0.0.0.0:8887"
	HostPort      = "172.17.0.1:3344"
	// MaxFileSize = 1GiB
	MaxFileSize = 1073741824
)

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

		var request types.RequestJSON

		json.NewDecoder(cnct).Decode(&request)

		go func() {
			cmdResult := execCmd(request)

			getErrorDetails(&cmdResult)

			if request.Mode == "judge" {
				cmdResult.IsOLE = getFileSize("userStdout.txt") > MaxFileSize
			}

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

			os.Remove("execCmd.sh")
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

func execCmd(request types.RequestJSON) types.CmdResultJSON {
	var cmdResult types.CmdResultJSON
	cmdResult.SessionID = request.SessionID

	if err := makeSh(request.Cmd); err != nil {
		log.Println(err)
		return cmdResult
	}

	cmd := exec.Command("sh", "-c", "/usr/bin/time -v ./execCmd.sh 2>&1 | grep -E 'Maximum' | awk '{ print $6 }' > mem_usage.txt")

	start := time.Now()
	timeout := time.After(2 * time.Second)

	if err := cmd.Start(); err != nil {
		cmdResult.ErrMessage = err.Error()
	}

	done := make(chan error)
	go func() { done <- cmd.Wait() }()

	select {
	case <-timeout:
		// timeout からシグナルが送られてきたらプロセスをキルする
		cmd.Process.Kill()
	case err := <-done:
		if err != nil {
			cmdResult.ErrMessage = err.Error()
		}
	}

	end := time.Now()

	cmdResult.Time = int((end.Sub(start)).Milliseconds())

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

// return (Bytes)
func getErrorDetails(cmdResult *types.CmdResultJSON) {
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

func getFileSize(name string) int64 {
	file, _ := os.Open(name)
	info, _ := file.Stat()

	return info.Size()
}