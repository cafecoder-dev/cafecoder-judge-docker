package util

import (
	"encoding/base64"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// mem_usage.txt, ret_code.txt に数値が記述されているため、それを読み取ってくる
func GetFileNum(name string) (int, error) {
	fp, err := os.Open(name)
	if err != nil {
		return 0, err
	}
	defer fp.Close()

	buf, err := ioutil.ReadAll(fp)
	if err != nil {
		return 0, err
	}

	tmp := strings.Replace(string(buf), "\n", "", -1)

	mem, err := strconv.Atoi(tmp)
	if err != nil {
		return 0, err
	}

	return mem, err
}

// return base64 encoded string
func GetFileStrBase64(name string) (string, error) {
	stderrFp, err := os.Open(name)
	if err != nil {
		return "", err
	}
	defer stderrFp.Close()

	buf, err := ioutil.ReadAll(stderrFp)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf) + "\n", nil
}

// ファイル(userStdout.txt)のサイズを読む
func GetFileSize(name string) int64 {
	info, err := os.Stat(name)
	if err != nil {
		return 0
	}

	return info.Size()
}

// container のホストの IP を引っ張ってくる
func GetHostIP() string {
	r, _ := exec.Command("sh", "-c", "ip route | awk 'NR==1 {print $3}'").Output()
	return strings.TrimRight(string(r), "\n") + ":3344"
}

// time -> string
func TimeToString(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
