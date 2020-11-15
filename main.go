package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/cafecoder-dev/cafecoder-container-client/gcplib"
	"github.com/cafecoder-dev/cafecoder-container-client/util"
	"github.com/cafecoder-dev/cafecoder-judge/src/checklib"
	"github.com/cafecoder-dev/cafecoder-judge/src/types"
)

const (
	ContainerPort = "0.0.0.0:8887"
	MaxFileSize   = 200000000 // 200MB
	MaxMemUsage   = 1024000
)

func main() {
	listen, err := net.Listen("tcp", ContainerPort) //from backend server
	if err != nil {
		os.Exit(1)
	}

	for {
		cnct, err := listen.Accept()
		if err != nil {
			continue //continue to receive request
		}

		go func() {
			var (
				request types.RequestJSON
				ctx     context.Context = context.Background()
			)

			_ = json.NewDecoder(cnct).Decode(&request)

			cmdResult := types.CmdResultJSON{
				SessionID: request.SessionID,
			}

			_ = os.Chmod("/", 0777)

			switch request.Mode {
			case "compile":
				cmdResult, err = execCmd(request)
				if err != nil {
					cmdResult.ErrMessage = base64.StdEncoding.EncodeToString([]byte(err.Error())) + "\n"
				}

			case "judge":
				cmdResult, err = tryTestcase(ctx, request)
				if err != nil {
					cmdResult.ErrMessage = base64.StdEncoding.EncodeToString([]byte(err.Error())) + "\n"
					cmdResult.Result = false
				}

			case "download":
				cmdResult = types.CmdResultJSON{SessionID: request.SessionID}
				if err = gcplib.DownloadSourceCode(ctx, request.CodePath, request.Filename); err != nil {
					cmdResult.ErrMessage = base64.StdEncoding.EncodeToString([]byte(err.Error())) + "\n"
					cmdResult.Result = false
				}
				cmdResult.Result = true

			default:
				cmdResult = types.CmdResultJSON{
					SessionID:  request.SessionID,
					Result:     false,
					ErrMessage: base64.StdEncoding.EncodeToString([]byte("invalid request")),
				}
			}

			b, err := json.Marshal(cmdResult)
			if err != nil {
				cmdResult.ErrMessage = err.Error() + "\n"
			}

			conn, err := net.Dial("tcp", util.GetHostIP())
			if err != nil {
				log.Fatal(err)
			}
			defer conn.Close()

			_, _ = conn.Write(b)

			conn.Close()
		}()
	}
}

// 実行するコマンドをシェルスクリプトに書き込む
func createSh(cmd string) error {
	f, err := os.Create("execCmd.sh")
	if err != nil {
		return err
	}

	_, _ = f.WriteString("#!/bin/bash\n")
	_, _ = f.WriteString("export PATH=$PATH:/usr/local/go/bin\n")
	_, _ = f.WriteString("export PATH=\"$HOME/.cargo/bin:$PATH\"\n")
	_, _ = f.WriteString(cmd + "\n")
	_, _ = f.WriteString("echo $? > exit_code.txt")

	f.Close()

	for {
		_, err := os.Stat("execCmd.sh")
		if err == nil {
			break
		}
	}

	_ = os.Chmod("execCmd.sh", 0777)

	return nil
}

func tryTestcase(ctx context.Context, request types.RequestJSON) (types.CmdResultJSON, error) {
	submitIDint64, _ := strconv.ParseInt(request.SessionID, 10, 64)

	testcaseInput, testcaseOutput, err := gcplib.DownloadTestcase(ctx, request.Problem.UUID, request.Testcase.Name)
	if err != nil {
		return types.CmdResultJSON{}, err
	}

	testcaseResults := types.TestcaseResultsGORM{SubmitID: submitIDint64, TestcaseID: request.Testcase.TestcaseID}

	file, _ := os.Create("./testcase.txt")
	_, _ = file.Write(testcaseInput)
	file.Close()

	res, err := execCmd(request)
	if err != nil {
		return types.CmdResultJSON{}, err
	}

	testcaseResults.Status, err = judging(request, res, string(testcaseOutput))
	if err != nil {
		return types.CmdResultJSON{}, err
	}

	testcaseResults.ExecutionTime = res.Time
	testcaseResults.ExecutionMemory = res.MemUsage
	testcaseResults.CreatedAt = util.TimeToString(time.Now())
	testcaseResults.UpdatedAt = util.TimeToString(time.Now())

	res.TestcaseResults = testcaseResults

	return res, nil
}

func judging(req types.RequestJSON, cmdres types.CmdResultJSON, testcaseOutput string) (string, error) {
	if cmdres.IsPLE {
		return "PLE", nil
	}
	if !cmdres.Result {
		return "RE", nil
	}
	if cmdres.Time > req.TimeLimit {
		return "TLE", nil
	}
	userOutput, err := ioutil.ReadFile("userStdout.txt")
	if err != nil {
		return "IE", err
	}
	if !checklib.Normal(string(userOutput), testcaseOutput) {
		return "WA", nil
	}
	if cmdres.StdoutSize > MaxFileSize {
		return "OLE", nil
	}
	if cmdres.MemUsage > MaxMemUsage {
		return "MLE", nil
	}

	return "AC", nil
}

// request.Cmd を実行する
func execCmd(request types.RequestJSON) (types.CmdResultJSON, error) {
	var (
		err       error
		cmdResult types.CmdResultJSON = types.CmdResultJSON{
			SessionID: request.SessionID,
		}
		timeout <-chan time.Time
	)

	if err := createSh(request.Cmd); err != nil {
		return cmdResult, err
	}

	cmd := exec.Command("sh", "-c", "/usr/bin/time -v ./execCmd.sh 2>&1 | grep -E 'Maximum' | awk '{ print $6 }' > mem_usage.txt")

	start := time.Now()

	// todo: ホストからタイムアウト指定するようにする
	if request.Mode == "compile" {
		timeout = time.After(20 * time.Second)
	} else {
		timeout = time.After(time.Duration(request.TimeLimit)*time.Millisecond + 200*time.Millisecond)
	}

	if err := cmd.Start(); err != nil {
		return cmdResult, err
	}

	done := make(chan error)
	go func() { done <- cmd.Wait() }()

	select {
	case <-timeout:
		// timeout からシグナルが送られてきたらプロセスをキルする
		if err := cmd.Process.Kill(); err != nil {
			return types.CmdResultJSON{}, err
		}
	case err := <-done:
		if err != nil {
			return cmdResult, err
		}
	}
	end := time.Now()

	cmdResult.Time = int((end.Sub(start)).Milliseconds())

	cmdResult.ErrMessage, err = util.GetFileStrBase64("/userStderr.txt")
	if err != nil {
		return cmdResult, err
	}

	cmdResult.MemUsage, _ = util.GetFileNum("mem_usage.txt")
	exitCode, _ := util.GetFileNum("exit_code.txt")

	cmdResult.Result = exitCode == 0

	cmdResult.StdoutSize = util.GetFileSize("userStdout.txt")

	return cmdResult, nil
}
