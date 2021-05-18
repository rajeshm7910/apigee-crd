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

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"

	"github.com/go-logr/logr"
	types "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +kubebuilder:scaffold:imports

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

func getMetadata(annotatedData string, log logr.Logger) (config string, env string, org string) {

	var result map[string]interface{}
	json.Unmarshal([]byte(annotatedData), &result)
	metadata := result["metadata"].(map[string]interface{})
	config = fmt.Sprintf("%v", metadata["config"])
	env = fmt.Sprintf("%v", metadata["env"])
	org = fmt.Sprintf("%v", metadata["org"])

	return config, env, org

}

func getAuth(client client.Client, log logr.Logger, configName string, namespace string) (baseUrl string, authString string, org string, env string) {

	var configMap corev1.ConfigMap
	if err := client.Get(context.TODO(), types.NamespacedName{Name: configName, Namespace: namespace}, &configMap); err != nil && apierrs.IsNotFound(err) {
		log.V(0).Info("Error in calling configmap")
	}

	config_type := configMap.Data["type"]
	auth := configMap.Data["auth"]

	baseUrl = configMap.Data["mgmt_api"]
	env = configMap.Data["env_name"]
	org = configMap.Data["org_name"]

	//log.V(1).Info("config type " + config_type)
	//log.V(1).Info("auth type " + auth)

	authString = ""

	if config_type == "legacy" {
		if auth == "base64" {
			username := configMap.Data["username"]
			password := configMap.Data["password"]
			encoded := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
			authString = "Basic " + " " + encoded
		} else if auth == "token" {
			authString = "Bearer "
		}
	} else {

		//Lets get the token call
		var secret corev1.Secret

		service_account_secret := configMap.Data["service_account_secret"]
		if err := client.Get(context.TODO(), types.NamespacedName{Name: service_account_secret, Namespace: namespace}, &secret); err != nil && apierrs.IsNotFound(err) {
		}

		service_account := secret.Data["service_account"]
		access_token, _ := generateAccessTokenFromSecret(service_account)
		authString = "Bearer " + " " + access_token
	}

	//log.V(1).Info("Auth  " + authString)

	//log.V(1).Info(fmt.Sprintf("configMap = %+v", configMap.Data["env_name"]))
	/*
		mgmt_api := configMap.Data["mgmt_api"]
		env_name := configMap.Data["env_name"]
		org_name := configMap.Data["org_name"]
		username := configMap.Data["username"]
		password := configMap.Data["password"]

		log.V(1).Info("Mgmt API " + mgmt_api)
		log.V(1).Info("Env  " + env_name)
		log.V(1).Info("Org  " + org_name)
		log.V(1).Info("user  " + username)
		//log.V(1).Info("Password  " + password)

		var configHybridMap corev1.ConfigMap
		if err := r.Client.Get(context.TODO(), types.NamespacedName{Name: "apigee-hybrid-config", Namespace: "apigee-config"}, &configHybridMap); err != nil && apierrs.IsNotFound(err) {
			log.V(0).Info("Error in calling configmap")
		}

		service_account_secret := configHybridMap.Data["service_account_secret"]
		mgmt_api_hyrbid := configHybridMap.Data["mgmt_api"]
		org_name_hybrid := configHybridMap.Data["org_name"]

		log.V(1).Info("Hybrid Mgmt API " + mgmt_api_hyrbid)
		log.V(1).Info("Org Hybrid  " + org_name_hybrid)
		log.V(1).Info("Service Account  " + service_account_secret)

	*/

	return baseUrl, authString, org, env
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

func PostHttpOctet(print bool, update bool, url string, proxyName string, authString string, log logr.Logger) (respBody []byte, err error) {
	file, _ := os.Open(proxyName)
	defer file.Close()

	var req *http.Request

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("proxy", proxyName)
	if err != nil {
		log.Error(err, "Post1")
		return nil, err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		log.Error(err, "Post2")
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		log.Error(err, "Post3")
		return nil, err
	}

	client := &http.Client{Timeout: time.Second * 10}

	if err != nil {
		log.Error(err, "Post4")
		return nil, err
	}

	if !update {
		req, err = http.NewRequest("POST", url, body)
	} else {
		req, err = http.NewRequest("PUT", url, body)
	}

	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", authString)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := client.Do(req)

	if err != nil {
		log.Error(err, "After sending request")
		return nil, err
	}

	defer resp.Body.Close()
	respBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	} else if !(resp.StatusCode == 200 || resp.StatusCode == 201) {
		log.V(0).Info(fmt.Sprintf("status Code = %+v", resp.StatusCode))
		log.Error(err, "Post5")
		return nil, errors.New("error in response")
	}
	if print {
		return respBody, PrettyPrint(respBody)
	}

	return respBody, nil
}

//PrettyPrint method prints formatted json
func PrettyPrint(body []byte) error {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, body, "", "\t")
	if err != nil {
		return err
	}
	fmt.Println(prettyJSON.String())
	return nil
}

func HttpClient(print bool, authString string, log logr.Logger, params ...string) (respBody []byte, err error) {
	// The first parameter instructs whether the output should be printed
	// The second parameter is url. If only one parameter is sent, assume GET
	// The third parameter is the payload. The two parameters are sent, assume POST
	// THe fourth parameter is the method. If three parameters are sent, assume method in param
	//The fifth parameter is content type
	var req *http.Request
	contentType := "application/x-www-form-urlencoded"

	client := &http.Client{Timeout: time.Second * 10}

	switch paramLen := len(params); paramLen {
	case 1:
		log.V(0).Info("Before making request GET")
		req, err = http.NewRequest("GET", params[0], nil)
	case 2:
		log.V(0).Info("Before making request POST")
		req, err = http.NewRequest("POST", params[0], bytes.NewBuffer([]byte(params[1])))
	case 3:
		if req, err = getRequest(params); err != nil {
			return nil, err
		}
	case 4:
		if req, err = getRequest(params); err != nil {
			return nil, err
		}
		contentType = params[3]
	default:
		return nil, errors.New("unsupported method")
	}

	req.Header.Add("Authorization", authString)
	req.Header.Set("Content-Type", contentType)

	log.V(0).Info("Before making request")

	resp, err := client.Do(req)

	log.V(0).Info("After making request")

	if resp != nil {
		log.V(0).Info("resp is not null")
		defer resp.Body.Close()
	}

	respBody, err = ioutil.ReadAll(resp.Body)
	log.V(0).Info("Getting Response body")
	log.V(0).Info(fmt.Sprintf("resp status code = %+v", resp.StatusCode))
	//PrettyPrint(respBody)

	if err != nil {
		log.Error(err, "Error in reading")
		return nil, err
	} else if resp.StatusCode > 299 {
		return nil, errors.New("error in response")
	}
	if print && contentType == "application/json" {
		return respBody, PrettyPrint(respBody)
	}
	return respBody, nil
}

func getRequest(params []string) (req *http.Request, err error) {
	if params[2] == "DELETE" {
		req, err = http.NewRequest("DELETE", params[0], nil)
	} else if params[2] == "PUT" {
		req, err = http.NewRequest("PUT", params[0], bytes.NewBuffer([]byte(params[1])))
	} else if params[2] == "PATCH" {
		req, err = http.NewRequest("PATCH", params[0], bytes.NewBuffer([]byte(params[1])))
	} else if params[2] == "POST" {
		req, err = http.NewRequest("POST", params[0], bytes.NewBuffer([]byte(params[1])))
	} else {
		return nil, errors.New("unsupported method")
	}
	return req, err
}
