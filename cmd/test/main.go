package main

import (
	"context"
	"log"
	"os"

	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
	iam "google.golang.org/api/iam/v1"
)

// THIS IS THE CODE USED BEHIND SEVERAL LAYERS OF ABSTRACTION IN HASHICORP'S GCPUTIL
func main() {
	ctx := context.Background()

	httpClient, err := google.DefaultClient(ctx, iam.CloudPlatformScope,
		compute.ComputeReadonlyScope)
	if err != nil {
		log.Println("unable to init client: ", err)
		os.Exit(1)
	}

	iamClient, err := iam.New(httpClient)
	if err != nil {
		log.Println("unable to init iam client: ", err)
		os.Exit(1)
	}

	const (
		projectID  = "games-puzzles-sandbox"
		svcAccount = "games-puzzles-sandbox@appspot.gserviceaccount.com"
		// keyID = "ca3f09762ee48b47c076360d2e71b78b93fba8f1"
		keyID = "ae94606d8f0c3e6a143b259a8f1a08e65f676ade"
	)

	resourceName := "projects/" + projectID + "/serviceAccounts/" + svcAccount + "/keys/" + keyID
	key, err := iamClient.Projects.ServiceAccounts.Keys.Get(resourceName).
		PublicKeyType("TYPE_X509_PEM_FILE").Do()
	if err != nil {
		log.Println("unable to get key: ", err)
		os.Exit(1)
	}

	log.Printf("success!!! %#v", key)
}
