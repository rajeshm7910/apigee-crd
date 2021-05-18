package controllers

import (
	"bytes"
	"context"
	"encoding/json"

	apigeev1 "apigee.com/m/api/v1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// ApiProductReconciler reconciles a ApiProduct object
type ApiProductReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apigee.my.domain,resources=apiproducts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apigee.my.domain,resources=apiproducts/status,verbs=get;update;patch

func (r *ApiProductReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("apiproduct", req.NamespacedName)

	log.V(1).Info("Starting the Prouct update")

	var instance apigeev1.ApiProduct

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

	url := baseUrl + "/organizations/" + org + "/apiproducts"

	log.V(0).Info("Setting Finalizers")
	// name of our custom finalizer
	myFinalizerName := "apiproducts.finalizers.apigee.kubebuilder.io"

	// examine DeletionTimestamp to determine if object is under deletion
	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !containsString(instance.ObjectMeta.Finalizers, myFinalizerName) {

			pushdata := parseInputAndCreateJSON(instance, log)
			data := []byte(pushdata)
			createApiProduct(instance.Spec.Name, data, url, authString, log)

			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, myFinalizerName)
			if err := r.Client.Update(context.Background(), &instance); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if containsString(instance.ObjectMeta.Finalizers, myFinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			log.V(0).Info("Name of Spec=" + instance.Spec.Name)
			deleteApiProduct(instance.Spec.Name, url, authString, log)
			// remove our finalizer from the list and update it.
			instance.ObjectMeta.Finalizers = removeString(instance.ObjectMeta.Finalizers, myFinalizerName)
			if err := r.Client.Update(context.Background(), &instance); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	//instance := apigeev1.ApiProduct{}
	//if err := r.Client.Get(ctx, req.NamespacedName, &instance); err != nil {
	//	return ctrl.Result{}, client.IgnoreNotFound(err)
	//}

	log.V(0).Info("Finishing the Prouct update")

	return ctrl.Result{}, nil
}

func (r *ApiProductReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apigeev1.ApiProduct{}).
		Complete(r)
}

func parseInputAndCreateJSON(instance apigeev1.ApiProduct, log logr.Logger) string {

	log.V(0).Info("In Parse Function")
	data, _ := json.Marshal(instance.Spec)
	pushdata := string(data)
	return pushdata

}

func deleteApiProduct(name string, url string, authString string, log logr.Logger) {

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
	log.V(1).Info("Deleting Product Apigee")
}

func checkApiProduct(name string, url string, authString string, log logr.Logger) bool {

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

	log.V(0).Info("returning false")

	return false
}

func createApiProduct(name string, data []byte, url string, authString string, log logr.Logger) {

	log.V(1).Info("calling http")
	method := "POST"
	if checkApiProduct(name, url, authString, log) {
		method = "PUT"
		url = url + "/" + name
	}

	req1, err1 := http.NewRequest(method, url, bytes.NewBuffer(data))
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

	fmt.Println("response Status:", resp1.Status)
	fmt.Println("response Headers:", resp1.Header)

	body, err1 := ioutil.ReadAll(resp1.Body)
	log.V(0).Info("calling http4")

	if err1 != nil {
		//log.Error("Error reading body. ", err)
	}
	fmt.Printf("%s\n", body)

}
