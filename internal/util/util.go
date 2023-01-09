package util

import (
	"fmt"
	"net/http"
	"net/netip"
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

func GetReqSource(req *http.Request) (source []string, err error) {
	var addr netip.Addr
	for _, s := range strings.FieldsFunc(req.Header.Get("X-Forwarded-For"),
		func(r rune) bool { return r == ',' || r == ' ' }) {
		addr, err = netip.ParseAddr(s)
		if err != nil {
			return
		}
		source = append(source, addr.String())
	}
	if addrPort, err := netip.ParseAddrPort(req.RemoteAddr); err == nil {
		source = append(source, addrPort.Addr().String())
	}
	return
}

func IsJenkinsValidationRequest(r *http.Request) bool {
	return r.Header.Get("X-Jenkins-Validation") == "true" && r.Header.Get("Content-Length") == "0" && strings.HasPrefix(r.Header.Get("User-Agent"), "Java")
}
