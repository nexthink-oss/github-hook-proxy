package cmd

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nexthink-oss/github-hook-proxy/internal/config"
	"github.com/nexthink-oss/github-hook-proxy/internal/util"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	buildVersion string = "snapshot"
	buildCommit  string = "unknown"
	buildDate    string = "unknown"

	cfg    config.Config
	logger *zap.Logger
)

var rootCmd = &cobra.Command{
	Use:          "github-hook-proxy",
	Short:        "GitHub hook multiplex proxy with validation",
	SilenceUsage: true,
	Version:      fmt.Sprintf("%s-%s (built %s)", buildVersion, buildCommit, buildDate),
	RunE:         runRootCmd,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initViper)

	rootCmd.Flags().StringP("config", "c", "config", "Configuration file name without extension")
	rootCmd.Flags().Bool("dump", false, "Dump processed configuration and exit")
}

func runRootCmd(cmd *cobra.Command, args []string) (err error) {
	logger = initZapLog()
	zap.ReplaceGlobals(logger)
	defer logger.Sync()

	configName, _ := cmd.Flags().GetString("config")
	err = cfg.LoadConfig(configName)
	if err != nil {
		return errors.Wrap(err, "loading config")
	}

	if dump, _ := cmd.Flags().GetBool("dump"); dump {
		util.PrettyPrint(cfg)
		return
	}

	if len(cfg.Targets) == 0 {
		return fmt.Errorf("no targets specified")
	}

	h := http.NewServeMux()
	for instance, target := range cfg.Targets {
		h.Handle(fmt.Sprintf("/%s", instance), target)
	}
	h.HandleFunc("/", Forbidden)

	s := &http.Server{
		Handler:      h,
		Addr:         net.JoinHostPort(cfg.Listener.Address, strconv.FormatUint(uint64(cfg.Listener.Port), 10)),
		WriteTimeout: 1 * time.Second,
		ReadTimeout:  1 * time.Second,
		IdleTimeout:  1 * time.Second,
		TLSConfig:    &tls.Config{MinVersion: tls.VersionTLS13},
	}

	return ListenAndServe(s)
}

func initViper() {
	viper.SetEnvPrefix("GHP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv() // read in environment variables that match bound variables
}

func initZapLog() *zap.Logger {
	var config zap.Config
	if util.IsTTY() {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config = zap.NewProductionConfig()
	}
	logger, _ := config.Build()
	return logger
}

func ListenAndServe(s *http.Server) (err error) {
	logger := logger.With(zap.String("address", s.Addr), zap.Int("targets", len(cfg.Targets)))
	if cfg.Listener.TLS.IsConfigured() {
		logger.Info("starting", zap.String("scheme", "https"))
		err = s.ListenAndServeTLS(cfg.Listener.TLS.PublicKey, cfg.Listener.TLS.PrivateKey)
	} else {
		logger.Info("starting", zap.String("scheme", "http"))
		err = s.ListenAndServe()
	}
	return
}

func Forbidden(w http.ResponseWriter, r *http.Request) {
	logger.Debug("forbidden", zap.String("source", r.RemoteAddr))
	w.WriteHeader(http.StatusForbidden)
}
