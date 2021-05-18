/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

// +kubebuilder:scaffold:imports

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.
import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/lestrrat/go-jwx/jwa"
	"github.com/lestrrat/go-jwx/jwt"
)

type serviceAccount struct {
	Type                string `json:"type,omitempty"`
	ProjectID           string `json:"project_id,omitempty"`
	PrivateKeyID        string `json:"private_key_id,omitempty"`
	PrivateKey          string `json:"private_key,omitempty"`
	ClientEmail         string `json:"client_email,omitempty"`
	ClientID            string `json:"client_id,omitempty"`
	AuthURI             string `json:"auth_uri,omitempty"`
	TokenURI            string `json:"token_uri,omitempty"`
	AuthProviderCertURL string `json:"auth_provider_x509_cert_url,omitempty"`
	ClientCertURL       string `json:"client_x509_cert_url,omitempty"`
}

var account = serviceAccount{}

func getPrivateKey(privateKey string) (interface{}, error) {
	pemPrivateKey := fmt.Sprintf("%v", privateKey)
	block, _ := pem.Decode([]byte(pemPrivateKey))
	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return privKey, nil
}

func generateJWT(privateKey string) (string, error) {
	const aud = "https://www.googleapis.com/oauth2/v4/token"
	const scope = "https://www.googleapis.com/auth/cloud-platform"

	privKey, err := getPrivateKey(privateKey)

	if err != nil {
		return "", err
	}

	now := time.Now()
	token := jwt.New()

	_ = token.Set(jwt.AudienceKey, aud)
	_ = token.Set(jwt.IssuerKey, getServiceAccountProperty("ClientEmail"))
	_ = token.Set("scope", scope)
	_ = token.Set(jwt.IssuedAtKey, now.Unix())
	_ = token.Set(jwt.ExpirationKey, now.Unix())

	payload, err := token.Sign(jwa.RS256, privKey)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func generateAccessTokenLegacy(token_url string, username string, password string, mfa_token string, log logr.Logger) (string, error) {

	//oAuthAccessToken is a structure to hold OAuth response
	type oAuthAccessToken struct {
		AccessToken string `json:"access_token,omitempty"`
		ExpiresIn   int    `json:"expires_in,omitempty"`
		TokenType   string `json:"token_type,omitempty"`
	}

	form := url.Values{}
	form.Add("grant_type", "password")
	form.Add("username", username)
	form.Add("password", password)

	if mfa_token != "<nil>" {
		token_url = token_url + "?mfa_token=" + mfa_token
	}

	log.V(1).Info("token_url " + token_url)

	client := &http.Client{}
	req, err := http.NewRequest("POST", token_url, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	//req.Header.Add("Content-Length", strconv.Itoa(len(form.Encode())))
	req.Header.Add("Authorization", "Basic ZWRnZWNsaTplZGdlY2xpc2VjcmV0")

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	log.V(1).Info(fmt.Sprintf("resp status code = %+v", resp.StatusCode))

	if resp.StatusCode != 200 {
		_, _ = ioutil.ReadAll(resp.Body)
		return "", errors.New("error in response")
	}
	decoder := json.NewDecoder(resp.Body)
	accessToken := oAuthAccessToken{}
	if err := decoder.Decode(&accessToken); err != nil {
		return "", errors.New("error in response")
	}

	return accessToken.AccessToken, nil

}

//generateAccessToken generates a Google OAuth access token from a service account
func generateAccessToken(privateKey string) (string, error) {

	const tokenEndpoint = "https://oauth2.googleapis.com/token"
	const grantType = "urn:ietf:params:oauth:grant-type:jwt-bearer"

	//oAuthAccessToken is a structure to hold OAuth response
	type oAuthAccessToken struct {
		AccessToken string `json:"access_token,omitempty"`
		ExpiresIn   int    `json:"expires_in,omitempty"`
		TokenType   string `json:"token_type,omitempty"`
	}

	token, err := generateJWT(privateKey)

	if err != nil {
		return "", nil
	}

	form := url.Values{}
	form.Add("grant_type", grantType)
	form.Add("assertion", token)

	client := &http.Client{}
	req, err := http.NewRequest("POST", tokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(form.Encode())))

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		_, _ = ioutil.ReadAll(resp.Body)
		return "", errors.New("error in response")
	}
	decoder := json.NewDecoder(resp.Body)
	accessToken := oAuthAccessToken{}
	if err := decoder.Decode(&accessToken); err != nil {
		return "", errors.New("error in response")
	}

	//SetApigeeToken(accessToken.AccessToken)
	//_ = WriteToken(accessToken.AccessToken)
	return accessToken.AccessToken, nil
}

func readServiceAccount(serviceAccountPath string) error {
	content, err := ioutil.ReadFile(serviceAccountPath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(content, &account)
	if err != nil {
		return err
	}
	return nil
}

func generateAccessTokenFromSecret(service_account []byte) (string, error) {

	json.Unmarshal(service_account, &account)
	privateKey := getServiceAccountProperty("PrivateKey")
	return generateAccessToken(privateKey)
}

func getServiceAccountProperty(key string) (value string) {
	r := reflect.ValueOf(&account)
	field := reflect.Indirect(r).FieldByName(key)
	return field.String()
}

func checkAccessToken() bool {
	if IsSkipCheck() {
		return true
	}

	const tokenInfo = "https://oauth2.googleapis.com/tokeninfo"
	u, _ := url.Parse(tokenInfo)
	q := u.Query()
	q.Set("access_token", GetApigeeToken())
	u.RawQuery = q.Encode()

	client := &http.Client{}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	} else if resp.StatusCode != 200 {
		return false
	}
	return true
}

//SetAccessToken read from cache or if not found or expired will generate a new one
func SetAccessToken() error {
	if GetApigeeToken() == "" && GetServiceAccount() == "" {
		//SetApigeeToken(GetToken()) //read from configuration
		if GetApigeeToken() == "" {
			return fmt.Errorf("either token or service account must be provided")
		}
		if checkAccessToken() { //check if the token is still valid
			return nil
		}
		return fmt.Errorf("token expired: request a new access token or pass the service account")
	}
	if GetApigeeToken() != "" {
		//a token was passed, cache it
		if checkAccessToken() {
			//_ = WriteToken(GetApigeeToken())
			return nil
		}
	} else {
		err := readServiceAccount(GetServiceAccount())
		if err != nil { // Handle errors reading the config file
			return fmt.Errorf("error reading config file: %s", err)
		}
		privateKey := getServiceAccountProperty("PrivateKey")
		if privateKey == "" {
			return fmt.Errorf("private key missing in the service account")
		}
		if getServiceAccountProperty("ClientEmail") == "" {
			return fmt.Errorf("client email missing in the service account")
		}
		_, err = generateAccessToken(privateKey)
		if err != nil {
			return fmt.Errorf("fatal error generating access token: %s", err)
		}
		return nil
	}
	return fmt.Errorf("token expired: request a new access token or pass the service account")
}
