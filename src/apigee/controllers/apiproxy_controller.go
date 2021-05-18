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
	"context"
	"encoding/json"
	"errors"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	apigeev1 "apigee.com/m/api/v1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"fmt"
	"io/ioutil"
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

	var instance apigeev1.ApiProxy

	if err := r.Client.Get(ctx, req.NamespacedName, &instance); err != nil {
		//log.Error(err, "unable to fetch API Product")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	annotatedData := instance.GetObjectMeta().GetAnnotations()["kubectl.kubernetes.io/last-applied-configuration"]
	config, _, _ := getMetadata(annotatedData, log)
	baseUrl, authString, org, env := getAuth(r.Client, log, config, "apigee-config")
	log.V(1).Info("Env " + env)

	url := baseUrl + "/organizations/" + org + "/apis"

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
			rev := createApiProxy(instance.Spec.Name, instance.Spec.ZipUrl, url, authString, log)
			instance.Status.DeploymentState = ""
			instance.Status.Revision = rev
			DeployApiProxy(instance.Spec.Name, rev, true, baseUrl, org, env, authString, log)
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
			log.V(0).Info("Calling undeploy")
			UnDeployApiProxy(instance.Spec.Name, rev, false, baseUrl, org, env, authString, log)

			instance.Status.DeploymentState = ""
			log.V(0).Info("Calling delete")
			deleteApiProxy(instance.Spec.Name, url, authString, log)
			// remove our finalizer from the list and update it.
			instance.ObjectMeta.Finalizers = removeString(instance.ObjectMeta.Finalizers, myFinalizerName)
			if err := r.Client.Update(context.Background(), &instance); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	log.V(0).Info("Finishing API Proxy Reconciliation")

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

func deleteApiProxy(name string, url string, authString string, log logr.Logger) {

	url = url + "/" + name

	req2, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		fmt.Println(err)
	}

	// Set headers
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", authString)

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
	log.V(1).Info("Deleting Apigee Proxy")
}

func checkApiProxy(name string, url string, authString string, log logr.Logger) bool {

	url = url + "/" + name
	req1, err1 := http.NewRequest("GET", url, nil)
	if err1 != nil {
		//err1
	}

	log.V(1).Info("calling http1")

	// Set headers
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Authorization", authString)

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

	log.V(0).Info("Returning false")

	return false
}

func createApiProxy(name string, zipurl string, Baseurl string, authString string, log logr.Logger) (revision int) {

	log.V(1).Info("calling http")
	u, _ := url.Parse(zipurl)
	log.V(1).Info(fmt.Sprintf("proto: %q, bucket: %q, key: %q", u.Scheme, u.Host, u.Path))

	bucket := u.Host
	proxy := strings.TrimLeft(u.Path, "/")
	log.V(1).Info("proxy name=" + proxy)

	data, _ := downloadFile(bucket, proxy)
	//log.V(1).Info(fmt.Sprintf("data = %+v", data))
	fileName := "/tmp/" + proxy

	WriteByteArrayToFile(fileName, false, data, log)
	rev, _ := ImportBundle(name, name, fileName, Baseurl, authString, log)
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
func ImportBundle(entityType string, name string, bundlePath string, BaseURL string, authString string, log logr.Logger) (revision int, err1 error) {
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
	//u.Path = path.Join(u.Path, entityType)
	log.V(0).Info("Path =" + u.Path)

	q := u.Query()
	q.Set("name", name)
	q.Set("action", "import")
	u.RawQuery = q.Encode()

	respBody, err := PostHttpOctet(true, false, u.String(), bundlePath, authString, log)
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

//DeployProxy
func DeployApiProxy(name string, revision int, overrides bool, BaseURL string, org string, env string, authString string, log logr.Logger) (respBody []byte, err error) {
	u, _ := url.Parse(BaseURL)
	if overrides {
		q := u.Query()
		q.Set("override", "true")
		u.RawQuery = q.Encode()
	}
	u.Path = path.Join(u.Path, "organizations", org, "environments", env,
		"apis", name, "revisions", strconv.Itoa(revision), "deployments")
	log.V(0).Info(u.String())
	respBody, err = HttpClient(true, authString, log, u.String(), "")
	return respBody, err
}

func UnDeployApiProxy(name string, revision int, overrides bool, baseURL string, org string, env string, authString string, log logr.Logger) (respBody []byte, err error) {
	u, _ := url.Parse(baseURL)
	if overrides {
		q := u.Query()
		q.Set("override", "true")
		u.RawQuery = q.Encode()
	}
	u.Path = path.Join(u.Path, "organizations", org, "environments", env,
		"apis", name, "revisions", strconv.Itoa(revision), "deployments")
	log.V(0).Info(u.String())
	respBody, err = HttpClient(true, authString, log, u.String(), "", "DELETE")
	return respBody, err
}
