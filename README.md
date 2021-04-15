# apigee-crd Apigee Custom Resource Definitions

Its goal is to manage Apigee  resources like apiproxies, apiproducts, developes etc as Kubernetes Objects.

This projects publishes bunch of Custom Resource Definitions for Apigee resources that can be deployed in a Kubernetes Cluster. At this time Apigee offers gcloud (imperative) or APIs to create Apigee resources. This offers declarative way of defining Apigee resources as yaml files. Finally kubectl can used to apply and manage these configurations.


### CRD Supported

- ApiProduct

### Getting Started

 1. Apply API Product CRD 
 ```kubectl apply -f crds/ApiProductCRD.yaml```
 2. Edit samples/config/opdk.yaml and set your properties
 3. Apply Configuration
 ```kubectl apply -f samples/config/opdk.yaml```
 5. To Test - 
 ```kubectl apply -f samples/apigee_v1_apiproduct.yaml```

	```
	kubectl get apiproducts
	NAME AGE
	my-samples 55s
	```
5. Check through Edge UI if APIProduct are created.
6. Delete the API Product 

	```
	kubectl delete apiproduct my-samples
	apiproduct.apigee.google.com "my-samples" deleted
	```
7. Check through Edge UI to see if the APIProduct is deleted
