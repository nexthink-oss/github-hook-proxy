package config

import (
	"github.com/mcuadros/go-defaults"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/nexthink-oss/github-hook-proxy/internal/target"
	"github.com/nexthink-oss/github-hook-proxy/internal/tls"
	"github.com/nexthink-oss/github-hook-proxy/internal/vault"
)

type Config struct {
	Listener Listener                  `mapstructure:"listener"`
	Targets  map[string]*target.Target `mapstructure:"targets"`
	Vault    vault.Vault               `mapstructure:"vault"`
	Verbose  bool                      `mapstructure:"verbose"`
}

type Listener struct {
	Address string  `mapstructure:"address" yaml:"address"`
	Port    uint16  `mapstructure:"port" yaml:"port"`
	TLS     tls.TLS `mapstructure:"tls" yaml:",omitempty"`
}

func (c *Config) LoadConfig(configName string) (err error) {
	viper.AddConfigPath("/etc/github-hook-proxy/")
	viper.AddConfigPath(".")
	viper.SetConfigName(configName)

	viper.SetDefault("listener.address", "127.0.0.1")
	viper.SetDefault("listener.port", "8080")
	viper.SetDefault("vault.address", "http://127.0.0.1:8200")
	viper.BindEnv("vault.address", "GHP_VAULT_ADDRESS", "VAULT_ADDR")
	viper.SetDefault("vault.mount", "secret")
	viper.SetDefault("vault.secret", "github-hooks/%s")
	viper.SetDefault("vault.field", "secret")
	viper.SetDefault("verbose", false)

	err = viper.ReadInConfig()
	if err != nil {
		return errors.Wrap(err, "problem reading config")
	}

	err = viper.UnmarshalExact(c)
	if err != nil {
		return errors.Wrap(err, "problem unmarshalling config")
	}

	return c.ProcessConfig()
}

func (c *Config) HasMissingSecrets() bool {
	for _, target := range c.Targets {
		if target.Secret == nil {
			return true
		}
	}
	return false
}

func (c *Config) ProcessConfig() (err error) {
	if c.HasMissingSecrets() && !c.Vault.IsInitialized() {
		err = c.Vault.Initialize()
		if err != nil {
			return
		}
	}

	for instance, target := range c.Targets {
		defaults.SetDefaults(target)
		err = target.FillHost()
		if err != nil {
			return
		}
		// load missing secrets from Vault
		if target.Secret == nil && c.Vault.IsInitialized() {
			secret, err := c.Vault.GetSecret(instance)
			if err != nil {
				// zap.L().Error("getting secret", zap.Error(err))
				return err
			}
			target.Secret = &secret
		}
	}
	return
}
