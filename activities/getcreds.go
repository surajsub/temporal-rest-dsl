package activities

//import (
//	"context"
//	"fmt"
//	vault "github.com/hashicorp/vault/api"
//	"github.com/surajsub/temporal-rest-dsl/models"
//	"time"
//)

/*

This activity should be called as and when credentials are needed by the workflow to fetch credentials from the vault system
The following is the structure of the data the needs to be passed to this activity
CustomerName
CustomerVaultUrl
CustomerUserName
CustomerPassword
Credentials of the service to be fetched ( We assume it's KV for now)
*/
//
//func GetCredsActivity(ctx context.Context, creds models.Credentials, vaultAddr, service string) (map[string]string, error) {
//	logger := GetDSLActivityLogger(ctx)
//	user := creds.UserName
//	password := creds.Password
//
//	logger.Infof("the username is %s", user, "and the password is %s", password)
//
//	return creds, nil
//
//}
