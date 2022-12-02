package util

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func PrettyPrint(obj interface{}) {
	bytes, _ := yaml.Marshal(obj)
	fmt.Println(string(bytes))
}

func IsTTY() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

func GetReqSource(req *http.Request) (source []string) {
	source = strings.FieldsFunc(req.Header.Get("X-Forwarded-For"),
		func(r rune) bool { return r == ',' || r == ' ' })
	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		source = append(source, clientIP)
	}
	return
}

func IsJenkinsValidationRequest(r *http.Request) bool {
	return r.Header.Get("X-Jenkins-Validation") == "true" && r.Header.Get("Content-Length") == "0" && strings.HasPrefix(r.Header.Get("User-Agent"), "Java")
}
