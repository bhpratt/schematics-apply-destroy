// Program that takes user input to either create or delete resources using the IBM Cloud Schematics service.
// Requires a pre-configured Schematics workspace.
// Expected input: `program <ibmcloud apikey> <schematics-workspace-id> <`apply` or `destroy`>`
// Apply sends a post call to IBM Cloud schematics to apply the configured workspace. Destroy sends a post call to tear down all resources in the workspace.
package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// struct for holding IAM response
type Iam struct {
	AccessToken       string `json:"access_token"`
	RefreshToken      string `json:"refresh_token"`
	UserID            int    `json:"ims_user_id"`
	TokenType         string `json:"token_type"`
	Expires           int    `json:"expires_in"`
	Expiration        int    `json:"expiration"`
	RefreshExpiration int    `json:"refresh_token_expiration"`
	Scope             string `json:"scope"`
}

// Main function. Parses commandline and sends request for tokens and the desired post call to IBM Cloud Schematics.
// Expected input: `main <ibmcloud apikey> <schematics-workspace-id> <`apply` or `destroy`>`
func main() {

	schematicsWorkspaceID := os.Args[2]
	action := os.Args[3]
	accessToken, refreshToken := getTokens(os.Args[1])
	clusterCreateOrDestroy(accessToken, refreshToken, action, schematicsWorkspaceID)
}

// The call to IAM that this command translates into GoLang:
//
//	curl --header "Content-Type: application/x-www-form-urlencoded" \
//	    --header "Accept: application/json" \
//	    --header --header "Authorization: Basic Yng6Yng=" \ ## this equals the url encoded authorization for username bx and password bx
//	    --data "grant_type=urn:ibm:params:oauth:grant-type:apikey" \
//	    --data "apikey=<apikey>" \
//		https://iam.cloud.ibm.com/identity/token
//
// Required input is an IBM Cloud API Key
// Output is loaded into the Iam struct and returns two strings holding the Access Token and Refresh Token
func getTokens(apiKey string) (string, string) {
	endpoint := "https://iam.cloud.ibm.com/identity/token"
	data := url.Values{}
	data.Set("grant_type", "urn:ibm:params:oauth:grant-type:apikey")
	data.Set("apikey", apiKey)
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		panic(err.Error())
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Basic Yng6Yng=")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}

	//print out status and response
	log.Println("IAM response:")
	log.Println(resp.Status)

	var iam Iam
	json.Unmarshal([]byte(body), &iam)
	//debugging
	// log.Println(iam)

	return iam.AccessToken, iam.RefreshToken
}

// The calls to IBM Cloud Schematics that this function translates to golang:
// apply: curl -X PUT https://schematics.cloud.ibm.com/v1/workspaces/{workspace-id}}/apply -H "Authorization: Bearer $IAM" -H "refresh_token: $REFRESH"
// destroy: curl -X PUT https://schematics.cloud.ibm.com/v1/workspaces/{workspace-id}/destroy -H "Authorization: Bearer <iam_token>" -H "refresh_token: <refresh_token>"
// Requires Access Token, Refresh Token, the action (either `apply` or `destroy`), and the IBM Cloud Schematics workspace ID
func clusterCreateOrDestroy(accessToken string, refreshToken string, action string, schematicsWorkspaceID string) {

	endpoint := "https://schematics.cloud.ibm.com/v1/workspaces/" + schematicsWorkspaceID + "/" + action
	log.Println("endpoint to target:")
	log.Println(endpoint)

	reqSchematics, err := http.NewRequest("PUT", endpoint, nil)
	if err != nil {
		panic(err.Error())
	}

	reqSchematics.Header.Set("Authorization", accessToken)
	reqSchematics.Header.Set("Refresh_token", refreshToken)

	// send requesting to schematics to apply or destroy resources in Schematics
	respClusterCreate, err := http.DefaultClient.Do(reqSchematics)
	if err != nil {
		panic(err.Error())
	}

	defer respClusterCreate.Body.Close()

	bodyClusterCreate, err := ioutil.ReadAll(respClusterCreate.Body)
	if err != nil {
		panic(err.Error())
	}

	log.Println("Schematics response:")
	log.Println(respClusterCreate.Status)
	log.Println(string(bodyClusterCreate))
}
