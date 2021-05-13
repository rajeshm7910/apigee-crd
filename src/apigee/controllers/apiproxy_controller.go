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
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	apigeev1 "apigee.com/m/api/v1"

	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	types "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"encoding/base64"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"time"

	"cloud.google.com/go/storage"
)

// ApiProxyReconciler reconciles a ApiProxy object
type ApiProxyReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apigee.google.com,resources=apiproxies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apigee.google.com,resources=apiproxies/status,verbs=get;update;patch

func (r *ApiProxyReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("apiproxies", req.NamespacedName)

	log.V(1).Info("Starting the Proxies update")

	var configMap corev1.ConfigMap
	if err := r.Client.Get(context.TODO(), types.NamespacedName{Name: "apigee-config", Namespace: "apigee-config"}, &configMap); err != nil && apierrs.IsNotFound(err) {
		log.V(0).Info("Error in calling configmap")
	}

	//log.V(1).Info(fmt.Sprintf("configMap = %+v", configMap.Data["env_name"]))
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

	encoded := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	base64_auth := "Basic " + encoded
	log.V(1).Info("Auth  " + base64_auth)

	url := mgmt_api + "/organizations/" + org_name + "/apis"
	log.V(1).Info("url  " + url)

	var instance apigeev1.ApiProxy

	if err := r.Client.Get(ctx, req.NamespacedName, &instance); err != nil {
		//log.Error(err, "unable to fetch API Product")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.V(0).Info("Setting Finalizers")
	// name of our custom finalizer
	myFinalizerName := "apiproxies.finalizers.apigee.kubebuilder.io"

	// examine DeletionTimestamp to determine if object is under deletion
	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !containsString(instance.ObjectMeta.Finalizers, myFinalizerName) {
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, myFinalizerName)
			log.V(0).Info("Updating Finalizers")
			rev := createApiProxy(instance.Spec.Name, instance.Spec.ZipUrl, url, base64_auth, log)
			instance.Status.DeploymentState = ""
			instance.Status.Revision = rev
			DeployApiProxy(instance.Spec.Name, rev, true, mgmt_api, org_name, env_name, base64_auth, log)
			instance.SetLabels(map[string]string{"revision": strconv.Itoa(rev)})
			instance.Status.DeploymentState = "deployed"

			if err := r.Client.Update(context.Background(), &instance); err != nil {
				return ctrl.Result{}, err
			}
		}

	} else {
		// The object is being deleted
		if containsString(instance.ObjectMeta.Finalizers, myFinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			log.V(0).Info("Name of Spec=" + instance.Spec.Name)
			labels := instance.GetLabels()
			rev, _ := strconv.Atoi(labels["revision"])

			UnDeployApiProxy(instance.Spec.Name, rev, true, mgmt_api, org_name, env_name, base64_auth, log)
			instance.Status.DeploymentState = ""
			deleteApiProxy(instance.Spec.Name, url, base64_auth, log)
			// remove our finalizer from the list and update it.
			instance.ObjectMeta.Finalizers = removeString(instance.ObjectMeta.Finalizers, myFinalizerName)
			if err := r.Client.Update(context.Background(), &instance); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	//createApiProxy(instance.Spec.Name, instance.Spec.ZipUrl, url, base64_auth, log)

	log.V(0).Info("Finishing API Proxy Reconciliation")

	// your logic here

	return ctrl.Result{}, nil
}

func (r *ApiProxyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apigeev1.ApiProxy{}).
		Complete(r)
}

func parseProxyInput(instance apigeev1.ApiProxy, log logr.Logger) string {

	log.V(0).Info("In Parse Function")
	data, _ := json.Marshal(instance.Spec)
	pushdata := string(data)
	return pushdata

}

func deleteApiProxy(name string, url string, base64_auth string, log logr.Logger) {

	url = url + "/" + name
	req2, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		fmt.Println(err)
	}

	// Set headers
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", base64_auth)

	// Set client timeout
	client := &http.Client{Timeout: time.Second * 10}
	// Fetch Request
	resp, err := client.Do(req2)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	// Read Response Body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%s\n", respBody)
	log.V(1).Info("Deleting Product Apigee")
}

func checkApiProxy(name string, url string, base64_auth string, log logr.Logger) bool {

	url = url + "/" + name
	req1, err1 := http.NewRequest("GET", url, nil)
	if err1 != nil {
		//err1
	}

	log.V(1).Info("calling http1")

	// Set headers
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Authorization", base64_auth)

	// Set client timeout
	client := &http.Client{Timeout: time.Second * 10}

	// Send request
	resp1, err1 := client.Do(req1)
	log.V(1).Info("calling http2")
	if err1 != nil {
		log.V(0).Info(err1.Error())
	}

	defer resp1.Body.Close()

	log.V(1).Info("calling http3")

	if resp1.StatusCode == 200 {
		log.V(0).Info("returning true")
		return true
	}

	log.V(0).Info("returning false")

	return false
}

func createApiProxy(name string, zipurl string, Baseurl string, base64_auth string, log logr.Logger) (revision int) {

	log.V(1).Info("calling http")
	u, _ := url.Parse(zipurl)
	log.V(1).Info(fmt.Sprintf("proto: %q, bucket: %q, key: %q", u.Scheme, u.Host, u.Path))

	bucket := u.Host
	proxy := strings.TrimLeft(u.Path, "/")
	log.V(1).Info("proxy name=" + proxy)

	data, _ := downloadFile(bucket, proxy)
	//log.V(1).Info(fmt.Sprintf("data = %+v", data))
	WriteByteArrayToFile("/tmp/temp00.zip", false, data, log)
	rev, _ := ImportBundle(name, name, "/tmp/temp00.zip", Baseurl, base64_auth, log)
	log.V(0).Info(fmt.Sprintf("Revision = %+v", rev))

	return rev

	//deploy proxy revision

}

// downloadFile downloads an object.
func downloadFile(bucket, object string) ([]byte, error) {
	// bucket := "bucket-name"
	// object := "object-name"
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("Object(%q).NewReader: %v", object, err)
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("ioutil.ReadAll: %v", err)
	}

	return data, nil
}

//WriteByteArrayToFile accepts []bytes and writes to a file
func WriteByteArrayToFile(exportFile string, fileAppend bool, payload []byte, log logr.Logger) error {
	var fileFlags = os.O_CREATE | os.O_WRONLY

	if fileAppend {
		fileFlags |= os.O_APPEND
	} else {
		fileFlags |= os.O_TRUNC
	}

	f, err := os.OpenFile(exportFile, fileFlags, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(payload)
	if err != nil {
		log.Error(err, "Error saving file")
		return err
	}
	log.V(0).Info("File Saved Successfully")
	return nil
}

//ImportBundle imports a sharedflow or api proxy bundle
func ImportBundle(entityType string, name string, bundlePath string, BaseURL string, base64_auth string, log logr.Logger) (revision int, err1 error) {
	err := ReadBundle(bundlePath)
	if err != nil {
		log.Error(err, "Reading Bundle")
		return 0, err
	}

	//when importing from a folder, proxy name = file name
	if name == "" {
		_, fileName := filepath.Split(bundlePath)
		names := strings.Split(fileName, ".")
		name = names[0]
	}

	u, _ := url.Parse(BaseURL)
	u.Path = path.Join(u.Path, entityType)
	log.V(0).Info("Path =" + u.Path)

	q := u.Query()
	q.Set("name", name)
	q.Set("action", "import")
	u.RawQuery = q.Encode()

	respBody, err := PostHttpOctet(true, false, u.String(), bundlePath, base64_auth, log)
	PrettyPrint(respBody)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)
	rev := fmt.Sprintf("%v", result["revision"])
	revInt, err := strconv.Atoi(rev)

	log.V(0).Info(fmt.Sprintf("Revision = %+v", revInt))

	if err != nil {
		log.Error(err, "Posting http")
		return 0, err
	}

	return revInt, nil
}

//ReadBundle confirms if the file format is a zip file
func ReadBundle(filename string) error {
	if !strings.HasSuffix(filename, ".zip") {
		return errors.New("source must be a zipfile")
	}

	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return err
	}

	_, err = zip.NewReader(file, fi.Size())

	if err != nil {
		return err
	}

	return nil
}

func PostHttpOctet(print bool, update bool, url string, proxyName string, base64_auth string, log logr.Logger) (respBody []byte, err error) {
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

	req.Header.Add("Authorization", base64_auth)
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
	} else if resp.StatusCode != 201 {
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

//DeployProxy
func DeployApiProxy(name string, revision int, overrides bool, BaseURL string, org string, env string, base64_auth string, log logr.Logger) (respBody []byte, err error) {
	u, _ := url.Parse(BaseURL)
	if overrides {
		q := u.Query()
		q.Set("override", "true")
		u.RawQuery = q.Encode()
	}
	u.Path = path.Join(u.Path, "organizations", org, "environments", env,
		"apis", name, "revisions", strconv.Itoa(revision), "deployments")
	log.V(0).Info(u.String())
	respBody, err = HttpClient(true, base64_auth, log, u.String(), "")
	return respBody, err
}

func UnDeployApiProxy(name string, revision int, overrides bool, BaseURL string, org string, env string, base64_auth string, log logr.Logger) (respBody []byte, err error) {
	u, _ := url.Parse(BaseURL)
	if overrides {
		q := u.Query()
		q.Set("override", "true")
		u.RawQuery = q.Encode()
	}
	u.Path = path.Join(u.Path, "organizations", org, "environments", env,
		"apis", name, "revisions", strconv.Itoa(revision), "deployments")
	log.V(0).Info(u.String())
	respBody, err = HttpClient(true, base64_auth, log, u.String(), "", "DELETE")
	return respBody, err
}

func HttpClient(print bool, base64_auth string, log logr.Logger, params ...string) (respBody []byte, err error) {
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

	req.Header.Add("Authorization", base64_auth)
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
