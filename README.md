# apigee-crd Apigee Custom Resource Definitions

Its goal is to manage Apigee  resources like apiproxies, apiproducts, developes etc as Kubernetes Objects.

This projects publishes bunch of Custom Resource Definitions for Apigee resources that can be deployed in a Kubernetes Cluster. At this time Apigee offers gcloud (imperative) or APIs to create Apigee resources. This offers declarative way of defining Apigee resources as yaml files. Finally kubectl can used to apply and manage these configurations.


### CRD Supported

- ApiProduct
- ApiProxy
- Developer
- DeveloperApp

### Getting Started

 1. For OPDK, Edit the config file as given in samples(samples/config/opdk.yaml) and apply Configuration.
 ```kubectl apply -f samples/config/opdk.yaml```

Please note that profile is legacy and auth is base64. In case you want to use SAML based token, please follow the steps mentioned in SAAS below.

 2. For SAAS, Edit the config file as given in samples(samples/config/saas.yaml) and apply Configuration.
  ```kubectl apply -f samples/config/saas.yaml```

Please note that profile is legacy and auth is token.The username and password are the machine users for automated token generation. In case you want to authenticate with mfa_token you can also put those values along with regular user credentials.
In case you want to use base64 credentials which is discouraged, you can proide auth value as base64.
 
 3. For Hybrid or ApigeeX

 - Create a secret from the Service Account json file.

  ```
 kubectl create secret generic amer-cs-hybrid-demo32-org-admin --from-file=service_account=./amer-cs-hybrid-demo32-org-admin.json --namespace apigee-config
 ```
 
 - Edit the config file as given in samples(samples/config/hybrid.yaml) and apply Hybrid specific configuration with the service_account_secret set to the secret name created above.

```
---
apiVersion: v1
kind: Namespace
metadata:
  labels:
    control-plane: controller-manager
  name: apigee-config
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: apigee-hybrid-config
  namespace: apigee-config
data:
  mgmt_api: https://apigee.googleapis.com/v1
  service_account_secret : amer-cs-hybrid-demo32-org-admin
  org_name: amer-cs-hybrid-demo32
  env_name: test
  profile: apigeex
  auth: token

```

```kubectl apply -f samples/config/hybrid.yaml```


4.  Applying CRD's

  Apply Apigee CRD 
 ```kubectl apply -f crds/```


5. To Test

 ```kubectl apply -f samples/apigee_v1_apiproduct.yaml```

Check the metadata section of the sample. You can specify the env in the metadata which will override the default environment provided in config above.  The config and config-namespace section should map to the config-map created above.

 ```
apiVersion: apigee.google.com/v1
kind: ApiProduct
metadata:
  name: my-samples
  env : test
  config: apigee-hybrid-config
  config-namespace: apigee-config
 ```

```
kubectl get apiproducts
NAME AGE
my-samples 55s
```

6. Check through Edge UI if APIProduct are created.
7. Delete the API Product 

	```
	kubectl delete apiproduct my-samples
	apiproduct.apigee.google.com "my-samples" deleted
	```
8. Check through Edge UI to see if the APIProduct is deleted
