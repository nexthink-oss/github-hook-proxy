[![CodeQL](https://github.com/nexthink-oss/github-hook-proxy/actions/workflows/codeql.yml/badge.svg)](https://github.com/nexthink-oss/github-hook-proxy/actions/workflows/codeql.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/nexthink-oss/github-hook-proxy)](https://goreportcard.com/report/github.com/nexthink-oss/github-hook-proxy)

# GitHub Hook Proxy

A validating proxy to facilitate secure delivery of GitHub webhook payloads to multiple targets behind a firewall.

## Features

* Support for multiple backend webhook targets
* GitHub payload validation based on HMAC-SHA256
* Optional secret storage in HashiCorp Vault K/V store

## Configuration

By default, the daemon will look for a Viper-style configuration file with the prefix "config" (i.e. `config.yaml` for YAML configuration, `config.toml` if you prefer TOML, etc.) in `/etc/github-hook-proxy` followed by the current working directory.

### Supported keys

The following root configuration keys are supported:

* `listener`: (optional) listener configuration (default: `{address: 127.0.0.1, port: 8080, tls: {}}`)
* `targets`: (required) list of targets (default: `[]`)
* `vault`: (optional) vault configuration (default: `{}`)

#### `listener`

* `address`: (optional) listener bind address (default: `127.0.0.1`)
* `port`: (optional) listener bind port (default: `8080`)
* `tls`: (optional) listener TLS certificate configuration, see below (default: `{}`)

The `tls` key if specified should contain two keys which, when set, will cause the proxy to listen for HTTPS rather than HTTP requests:

* `private-key`: (required) path to PEM format TLS private key
* `public-key`: (required) path to PEM format TLS public key

#### `targets`

Each target object takes the form `"<targetName>": {}`, with the following keys supported:

* `url`: (required) full URL to which payloads for this target should be forwarded
* `secret`: (optional) shared secret for validation of payloads associated with this target (default: load from vault); an explicitly blank secret (`secret: ""`) will disable payload validation
* `events`: (optional) list of events to accept for this target (default: `[ping, push, pull_request]`)
* `jenkins-validation`: (optional) boolean controlling whether to accept [Jenkins GitHub plugin](https://plugins.jenkins.io/github/) validation requests (default: `false`); *not* required for Jenkins to receive externally configured webhook payloads

#### `vault`

If any target does not specify a static secret, then Vault must be configured:

* `address`: (optional) full URL of your Vault instance (default: `https://127.0.0.1:8080`); may also be specified via VAULT_ADDR.
* `token-file`: (optional) path to Vault token file; the VAULT_TOKEN environment variable will take priority.
* `mount`: (optional) mountpoint of the Vault K/V v2 store holding target webhook secrets (default: `secret`)
* `secret`: (optional) template string for the path within the Vault K/V v2 store holding a specific target's webhook secret; must contain a single `%s` which will be filled in with each target's name. (default: `github-webhooks/%s`)
* `field`: (optional) field within the target's K/V v2 secret holding the GitHub webhook secret (default: `secret`)

### Example Configuration

See [`example-config.yaml`](example-config.yaml) for example configuration.
