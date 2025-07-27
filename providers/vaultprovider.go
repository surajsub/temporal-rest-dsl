package providers

import (
	"context"
	"fmt"
	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	"log"
	"time"
)

type VaultSecretsProvider struct {
	client        *vault.Client
	secretsPath   string
	mountPath     string
	vaultURL      string
	vaultCertPath string
}

func (v *VaultSecretsProvider) Init(config map[string]string) error {

	ctx := context.Background()
	// prepare a client

	tls := vault.TLSConfiguration{}
	tls.ServerCertificate.FromFile = v.vaultCertPath

	client, err := vault.New(
		vault.WithAddress(v.vaultURL),
		vault.WithRequestTimeout(30*time.Second),
		vault.WithTLS(tls),
	)
	if err != nil {
		return nil
	}

	//err = client.Auth().AppRoleLogin(context.Background(), &vault.Auth{}

	resp, err := client.Auth.AppRoleLogin(
		ctx,
		schema.AppRoleLoginRequest{
			RoleId:   config["role_id"],
			SecretId: config["secret_id"],
		},
		vault.WithMountPath("approle"), // optional, defaults to "approle"
	)

	if err != nil {
		return fmt.Errorf("vault login failed: %w", err)
	}

	log.Println("Printing this otoken for validation ********", resp.Auth.ClientToken)

	v.client = client
	return nil

}

func (v *VaultSecretsProvider) GetCredentials() (map[string]any, error) {
	secret, err := v.client.Secrets.KvV2Read(
		context.Background(),
		v.mountPath,
		vault.WithMountPath(v.mountPath),
	)
	if err != nil {
		log.Println(err.Error())
		log.Fatal(err)
	}

	log.Println(secret.Data)
	return secret.Data.Data, nil

}
