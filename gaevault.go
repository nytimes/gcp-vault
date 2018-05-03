package gaevault

import (
	"context"

	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"

	cloudkms "google.golang.org/api/cloudkms/v1"
	"google.golang.org/appengine/urlfetch"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type VaultInfo struct {
	RoleID            string
	EncryptedSecretID string
	LoginPath         string
	SecretPath        string
}

type KMSInfo struct {
	ProjectID   string
	Locations   string
	KeyRingID   string
	CryptoKeyID string
}

func (k KMSInfo) name() string {
	return "projects/" + k.ProjectID + "/locations/" + k.Locations + "/keyRings/#" +
		k.KeyRingID + "/cryptoKeys/#" + k.CryptoKeyID
}

func GetSecrets(ctx context.Context, kInfo KMSInfo, vInfo VaultInfo) (map[string]interface{}, error) {
	// grab KMS client
	ks, err := cloudkms.New(oauth2.NewClient(ctx,
		google.AppEngineTokenSource(ctx, "https://www.googleapis.com/auth/cloudkms")))
	if err != nil {
		return nil, errors.Wrap(err, "unable to init KMS client")
	}
	kmsClient := cloudkms.NewProjectsLocationsKeyRingsCryptoKeysService(ks)

	// decrypt our vault secret
	res, err := kmsClient.Decrypt(kInfo.name(),
		&cloudkms.DecryptRequest{Ciphertext: vInfo.EncryptedSecretID}).Context(ctx).Do()
	if err != nil {
		return nil, errors.Wrap(err, "unable to decrypt secret ID via KMS")
	}

	// init vault client
	vcfg := api.DefaultConfig()
	vcfg.HttpClient = urlfetch.Client(ctx)
	vClient, err := api.NewClient(vcfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init vault client")
	}

	// 'login' to vault
	_, err = vClient.Logical().Write(vInfo.LoginPath, map[string]interface{}{
		"role_id": vInfo.RoleID, "secret_id": res.Plaintext,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to login to vault")
	}

	// fetch secrets
	secrets, err := vClient.Logical().Read(vInfo.SecretPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get secrets")
	}

	return secrets.Data, nil
}
