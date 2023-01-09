package target

import (
	"bytes"
	"net/http"
	"net/url"

	"github.com/nexthink-oss/github-hook-proxy/internal/util"

	"github.com/google/go-github/v48/github"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

type Target struct {
	Events  []string `default:"[ping,push,pull_request]" mapstructure:"events" yaml:",flow"`
	Jenkins bool     `mapstructure:"jenkins-validation" yaml:"jenkins-validation,omitempty"`
	Secret  *string  `mapstructure:"secret"`
	URL     string   `mapstructure:"url"`
	host    string
}

func (t *Target) CheckPayload(r *http.Request) (payload []byte, err error) {
	return github.ValidatePayload(r, []byte(*t.Secret))
}

func (t *Target) AcceptsEvent(r *http.Request) bool {
	if t.Events == nil {
		return true
	}
	return slices.Contains(t.Events, github.WebHookType(r))
}

func (t *Target) FillHost() (err error) {
	url, err := url.Parse(t.URL)
	if err == nil && url.Host != "" {
		t.host = url.Host
	}
	return
}

func (t *Target) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := zap.L().With(zap.String("target", t.host), zap.String("method", req.Method))

	source, err := util.GetReqSource(req)
	if err != nil {
		logger.With(zap.String("source", req.RemoteAddr)).Warn("received invalid X-Forwarded-For header")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	logger = logger.With(zap.Strings("source", source))

	if req.Method == "POST" {
		if t.Jenkins && util.IsJenkinsValidationRequest(req) {
			logger.Info("X-Jenkins-Validation request")
			rw.WriteHeader(http.StatusOK)
			return
		}
		payload, err := t.CheckPayload(req)
		if err != nil {
			logger.Debug("received invalid payload", zap.Error(err))
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		logger = logger.With(zap.String("event_type", github.WebHookType(req)), zap.String("event_id", github.DeliveryID(req)))
		logger.Debug("payload ok")
		if !t.AcceptsEvent(req) {
			logger.Debug("reject event")
			rw.WriteHeader(http.StatusNotImplemented)
			return
		}
		resp, err := t.ProxyRequest(req, payload)
		if err != nil {
			logger.Warn("proxy error", zap.Error(err))
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		logger.Info("complete", zap.Int("status", resp.StatusCode))
		rw.WriteHeader(resp.StatusCode)
	} else {
		logger.Debug("forbidden")
		rw.WriteHeader(http.StatusForbidden)
	}
}

func (t *Target) ProxyRequest(r *http.Request, payload []byte) (resp *http.Response, err error) {
	client := http.DefaultClient
	body := bytes.NewReader(payload)
	request, err := http.NewRequestWithContext(r.Context(), r.Method, t.URL, body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to construct proxy request")
	}

	request.Header = r.Header
	request.Header.Set("Host", t.host)

	return client.Do(request)
}
