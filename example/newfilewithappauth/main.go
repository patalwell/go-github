//Copyright 2021 The go-github AUTHORS. All rights reserved.
//
//Use of this source code is governed by a BSD-style
//license that can be found in the LICENSE file.

//newfilewithappauth demonstrates the functionality of GitHub's app authentication
//methods by fetching an installation access token and reauthenticating to GitHub
//with OAuth configurations.
package main

import (
	"context"
	"fmt"
	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v41/github"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/http"
	"time"
)

func main() {

	const gitHost = "https://git.api.com"

	privatePem, err := ioutil.ReadFile("path/to/pem")
	if err != nil {
		panic(err)
	}

	itr, err := ghinstallation.NewAppsTransport(http.DefaultTransport, 10, privatePem)
	if err != nil {
		fmt.Printf("faild to create app transport: %v\n", err)
	}
	itr.BaseURL = gitHost

	//create git client with app transport
	client, err := github.NewEnterpriseClient(
		gitHost,
		gitHost,
		&http.Client{
			Transport: itr,
			Timeout:   time.Second * 30,
		})
	if err != nil {
		fmt.Printf("faild to create git client for app: %v\n", err)
	}

	installations, _, err := client.Apps.ListInstallations(context.Background(), &github.ListOptions{})
	if err != nil {
		fmt.Printf("failed to list installations: %v\n", err)
	}

	//capture our installationId for our app
	//we need this for the access token
	var installID int64
	for _, val := range installations {
		installID = val.GetID()
	}

	token, _, err := client.Apps.CreateInstallationToken(
		context.Background(),
		installID,
		&github.InstallationTokenOptions{})
	if err != nil {
		fmt.Printf("failed to create installation token: %v\n", err)
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token.GetToken()},
	)
	oAuthClient := oauth2.NewClient(context.Background(), ts)

	//create new git hub client with accessToken
	apiClient, err := github.NewEnterpriseClient(gitHost, gitHost, oAuthClient)
	if err != nil {
		fmt.Printf("failed to create new git client with token: %v\n", err)
	}

	_, resp, err := apiClient.Repositories.CreateFile(
		context.Background(),
		"repoOwner",
		"sample-repo",
		"example/foo.txt",
		&github.RepositoryContentFileOptions{
			Content: []byte("foo"),
			Message: github.String("sample commit"),
			SHA:     nil,
		})
	if err != nil {
		fmt.Printf("failed to create new file: %v\n", err)
	}

	fmt.Printf("file written: %v", resp.StatusCode)
}