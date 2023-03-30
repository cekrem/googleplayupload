package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

const (
	credentialsEnv = "GOOGLE_PLAY_CREDENTIALS_PATH"
	nameEnv        = "GOOGLE_PLAY_PACKAGE_NAME"
	pathEnv        = "GOOGLE_PLAY_PACKAGE_PATH"
	commitEnv      = "GOOGLE_PLAY_COMMIT"
)

func main() {
	var failed bool

	credentialsPath, ok := os.LookupEnv(credentialsEnv)
	if !ok {
		failed = true
		fmt.Fprintf(os.Stderr, "missing %s\n", credentialsEnv)
	}
	packageName, ok := os.LookupEnv(nameEnv)
	if !ok {
		failed = true
		fmt.Fprintf(os.Stderr, "missing %s\n", nameEnv)
	}
	packagePath, ok := os.LookupEnv(pathEnv)
	if !ok {
		failed = true
		fmt.Fprintf(os.Stderr, "missing %s\n", pathEnv)
	}

	if failed {
		os.Exit(1)
	}

	packageFile, err := os.Open(packagePath)
	if err != nil {
		log.Fatalf("error opening %q: %v\n", packagePath, err)
	}

	ctx := context.Background()
	service, err := androidpublisher.NewService(ctx, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		log.Fatalln(err)
	}

	idRes, err := service.Edits.Insert(packageName, &androidpublisher.AppEdit{}).Do()
	if err != nil {
		log.Fatalf("error creating edit, %v\n", err)
		os.Exit(1)
	}

	uploadRes, err := service.Edits.Bundles.Upload(packageName, idRes.Id).Media(packageFile, googleapi.ContentType("application/octet-stream")).Do()
	if err != nil {
		log.Fatalf("error uploading package, %v\n", err)
		os.Exit(1)
	}

	_, err = service.Edits.Tracks.Update(packageName, idRes.Id, "alpha", &androidpublisher.Track{
		Releases: []*androidpublisher.TrackRelease{
			{VersionCodes: []int64{uploadRes.VersionCode}},
		},
	}).Do()
	if err != nil {
		log.Fatalf("error updating tracks, %v\n", err)
		os.Exit(1)
	}

	if os.Getenv(commitEnv) == "true" {
		commitRes, err := service.Edits.Commit(packageName, idRes.Id).Do()
		if err != nil {
			log.Fatalf("error commiting edit, %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("commited edit successfully, %v\n", commitRes)
	} else {
		err = service.Edits.Delete(packageName, idRes.Id).Do()
		fmt.Printf("edit aborted with err: %v\n", err)
	}
}
