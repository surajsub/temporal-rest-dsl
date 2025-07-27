package executors

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	"github.com/surajsub/temporal-rest-dsl/models"
)

// This executor is responsible to execute the vault commands to perform the required operation.

type VaultExecutor struct {
	*ExecutorBase
	Submitter string
	Project   string
}

// Constructor for VaulExecutor
func NewVaultExecutor(customer, workspace, provider, resource, provisioner, action, operation, submitter, project string) *VaultExecutor {
	return &VaultExecutor{
		ExecutorBase: NewExecutorBase(customer, workspace, provider, resource, provisioner, action, operation),
		Submitter:    submitter,
		Project:      project,
	}
}

func (v *VaultExecutor) Execute(step models.Step, executor string, payload map[string]any) (map[string]interface{}, error) {
	v.Logger.Infof("Executing Vault Executor with action %s and operation %s", v.Operation)

	if step.ID == "getcreds" {
		data, err := v.GetCredentialsFromVault(payload["url"].(string), payload["mount_path"].(string), payload["secret_path"].(string), step.SecretId, step.RoleID)
		if err != nil {
			return nil, err
		}

		return data, nil
	}

	return nil, errors.New("unsupported operation for Vault Executor")
}

func (v *VaultExecutor) ValidateOperation(step models.Step) error {

	return nil
}

func (v *VaultExecutor) GetCredentialsFromVault(url, path, secretpath, secretid, roleid string) (map[string]interface{}, error) {

	ctx := context.Background()
	// prepare a client

	tls := vault.TLSConfiguration{}
	tls.ServerCertificate.FromFile = "./vaultwork/vault-tls/ca.crt"

	client, err := vault.New(
		vault.WithAddress(url),
		vault.WithRequestTimeout(30*time.Second),
		vault.WithTLS(tls),
	)
	if err != nil {
		return nil, nil
	}

	log.Println("Getting credentials from Vault")

	// In order to get this secret , we need to do the following

	resp, err := client.Auth.AppRoleLogin(
		ctx,
		schema.AppRoleLoginRequest{
			RoleId:   roleid,
			SecretId: secretid,
		},
		vault.WithMountPath("approle"), // optional, defaults to "approle"
	)
	if err != nil {
		log.Println(err)
		log.Fatal(err)
	}

	log.Println("Printing this otoken for validation ********", resp.Auth.ClientToken)

	if err := client.SetToken(resp.Auth.ClientToken); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully authenticated using AppRole")

	fmt.Printf("This is the token %s", client.SetToken(resp.Auth.ClientToken))

	path = path // string | Location of the secret.
	resp1, err := client.Secrets.KvV2Read(
		context.Background(),
		secretpath,
		vault.WithMountPath(path),
	)
	if err != nil {
		log.Println(err.Error())
		log.Fatal(err)
	}

	log.Println(resp1.Data)

	return resp1.Data.Data, nil
}
