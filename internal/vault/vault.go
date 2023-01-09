package vault

import (
	"context"
	"fmt"
	"os"
	"strings"

	vault "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Vault struct {
	client    *vault.Client
	Address   string `mapstructure:"address"`
	TokenFile string `mapstructure:"token-file" yaml:"token-file,omitempty"`
	Mount     string `mapstructure:"mount"`
	Secret    string `default:"%s/github-webhook" mapstructure:"secret" yaml:"secret"`
	Field     string `default:"secret" mapstructure:"field" yaml:"field"`
}

func (v *Vault) IsInitialized() bool {
	return v.client != nil
}

func (v *Vault) Initialize() (err error) {
	config := vault.DefaultConfig()
	if err = config.ReadEnvironment(); err != nil {
		return
	}

	if v.Address != "" {
		config.Address = v.Address
	}

	client, err := vault.NewClient(config)
	if err != nil {
		return
	}

	if client.Token() == "" {
		if v.TokenFile != "" {
			tokenBytes, err := os.ReadFile(v.TokenFile)
			if err != nil {
				return err
			}
			token := strings.TrimSpace(string(tokenBytes))
			client.SetToken(token)
		} else {
			return fmt.Errorf("no Vault token found")
		}
	}

	v.client = client
	return
}

// GetSecret retrieves the secret for a specific target instance
func (v *Vault) GetSecret(targetName string) (secret string, err error) {
	zap.L().Info("read secret from Vault", zap.String("target", targetName))

	mountPath := v.client.KVv2(v.Mount)
	if mountPath == nil {
		return "", fmt.Errorf("unable to open path %q", v.Mount)
	}

	var secretPath string
	switch len(strings.Split(v.Secret, "%")) {
	case 1:
		// secret path contains no format verbs
		secretPath = v.Secret
	case 2:
		secretPath = fmt.Sprintf(v.Secret, targetName)
	default:
		return "", fmt.Errorf("invalid secret_path: %q", v.Secret)
	}

	secretData, err := mountPath.Get(context.Background(), secretPath)
	if err != nil {
		return "", errors.Wrap(err, "unable to read secret")
	}

	secretKey, ok := secretData.Data[v.Field]
	if !ok {
		return "", fmt.Errorf("the specified secret is missing the %q field", v.Field)
	}

	secret, ok = secretKey.(string)
	if !ok {
		return "", fmt.Errorf("unexpected secret key type for %q field", v.Field)
	}

	return
}
