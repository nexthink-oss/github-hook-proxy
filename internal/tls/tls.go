package tls

type TLS struct {
	PrivateKey string `mapstructure:"private-key" yaml:"private-key,omitempty"`
	PublicKey  string `mapstructure:"public-key" yaml:"public-key,omitempty"`
}

func (t *TLS) IsConfigured() bool {
	return t.PrivateKey != "" && t.PublicKey != ""
}
