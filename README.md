# apigee-crd Apigee Custom Resource Definitions

Its goal is to manage Apigee  resources like apiproxies, apiproducts, developes etc as Kubernetes Objects.

This project publishes Custom Resource Definition for Apigee resources that can be deployed in a Kubernetes Cluster.This offers declarative way of defining Apigee resources as yaml files. Finally the resources can be managed by the easy-to-use kubectl commands.


### CRD Supported

- ApiProduct
- ApiProxy
- Developer
- DeveloperApp

### Getting Started

 1. For OPDK, Edit the config file as given in samples(samples/config/opdk.yaml) and apply Configuration.
 ```kubectl apply -f samples/config/opdk.yaml```

opdk sample config

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: apigee-config-opdk
  namespace: apigee-config
data:
  mgmt_api: http://xx.xx.xx.xx:8080/v1
  username: opdk@apigee.com
  password: Secret123
  org_name: demo
  env_name: test
  profile: legacy
  auth: base64
```

Please note that profile is legacy and auth is base64. In case you want to use SAML based token, please follow the steps mentioned in SAAS below.

 2. For SAAS, Edit the config file as given in samples(samples/config/saas.yaml) and apply Configuration.
  ```kubectl apply -f samples/config/saas.yaml```

saas sample config
```
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: apigee-config-saas
  namespace: apigee-config
data:
  mgmt_api: https://api.enterprise.apigee.com/v1
  username: xcxcxcxcx@apigee.com
  password: xxcxcxcxc
  token_url: https://login.apigee.com/oauth/token
  #mfa_token: "695573"
  org_name: imedusa
  env_name: test
  profile: legacy
  auth: token
```

Please note that profile is legacy and auth is token.The username and password are the machine users for automated token generation. In case you want to authenticate with mfa_token you can also put those values along with regular user credentials.
In case you want to use base64 credentials which is deprecated, you can proide auth value as base64.
 
 3. For Hybrid or ApigeeX

- Create apigee-config namespace
```
kubectl create namespace apige-config
```
- Obtain the Service Account json file with the appropriate role. 

- Create a kubernetes secret from the Service Account json file.

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

```kubectl apply -f samples/apigee_v1_apiproxy.yaml```

Check the metadata section of this sample. You can specify the env in the metadata which will override the default env provided in config above.  The config and config-namespace section should map to the config-map created above.

 ```
apiVersion: apigee.google.com/v1
kind: ApiProxy
metadata:
  name: loans-api
  env : test
  config: apigee-hybrid-config
  config-namespace: apigee-config
 ```

```
kubectl get apiproxies
NAME        DEPLOYMENT   REVISION   AGE
loans-api   deployed     1          13s
```

6. Check through Edge UI if Proxies are created.
7. Delete the API Proxy 

	```
	kubectl delete -f samples/apigee_v1_apiproxy.yaml
apiproxy.apigee.google.com "loans-api" deleted
	```
8. Check through Edge UI to see if the Api Proxy is deleted
