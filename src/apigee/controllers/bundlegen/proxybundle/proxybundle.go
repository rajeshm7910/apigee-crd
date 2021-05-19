// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package proxybundle

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-logr/logr"

	genapi "apigee.com/m/controllers/bundlegen"
	apiproxy "apigee.com/m/controllers/bundlegen/apiproxydef"
	policies "apigee.com/m/controllers/bundlegen/policies"
	proxies "apigee.com/m/controllers/bundlegen/proxies"
	target "apigee.com/m/controllers/bundlegen/targetendpoint"
)

func GenerateAPIProxyBundle(name string, content string, fileName string, skipPolicy bool, log logr.Logger) (err error) {
	const rootDir = "apiproxy"
	var apiProxyData, proxyEndpointData, targetEndpointData string

	if err = os.Mkdir(rootDir, os.ModePerm); err != nil {
		return err
	}

	// write API Proxy file
	if apiProxyData, err = apiproxy.GetAPIProxy(); err != nil {
		return err
	}

	err = writeXMLData(rootDir+string(os.PathSeparator)+name+".xml", apiProxyData)
	if err != nil {
		return err
	}

	proxiesDirPath := rootDir + string(os.PathSeparator) + "proxies"
	policiesDirPath := rootDir + string(os.PathSeparator) + "policies"
	targetDirPath := rootDir + string(os.PathSeparator) + "targets"
	oasDirPath := rootDir + string(os.PathSeparator) + "resources" + string(os.PathSeparator) + "oas"

	if err = os.Mkdir(proxiesDirPath, os.ModePerm); err != nil {
		return err
	}

	if proxyEndpointData, err = proxies.GetProxyEndpoint(); err != nil {
		return err
	}

	err = writeXMLData(proxiesDirPath+string(os.PathSeparator)+"default.xml", proxyEndpointData)
	if err != nil {
		return err
	}

	if err = os.Mkdir(targetDirPath, os.ModePerm); err != nil {
		return err
	}

	if targetEndpointData, err = target.GetTargetEndpoint(); err != nil {
		return err
	}

	if err = writeXMLData(targetDirPath+string(os.PathSeparator)+"default.xml", targetEndpointData); err != nil {
		return err
	}

	if !skipPolicy {
		if err = os.MkdirAll(oasDirPath, os.ModePerm); err != nil {
			return err
		}
		if err = writeXMLData(oasDirPath+string(os.PathSeparator)+fileName, content); err != nil {
			return err
		}
	}

	if err = os.Mkdir(policiesDirPath, os.ModePerm); err != nil {
		return err
	}

	//add oauth policy
	if genapi.GenerateOAuthPolicy() {
		if err = writeXMLData(policiesDirPath+string(os.PathSeparator)+"OAuth-v20-1.xml", policies.AddOAuth2Policy()); err != nil {
			return err
		}
	}

	//add api key policy
	if genapi.GenerateAPIKeyPolicy() {
		if err = writeXMLData(policiesDirPath+string(os.PathSeparator)+"Verify-API-Key-1.xml", policies.AddVerifyApiKeyPolicy()); err != nil {
			return err
		}
	}

	if !skipPolicy {
		//add oas policy
		if err = writeXMLData(policiesDirPath+string(os.PathSeparator)+"OpenAPI-Spec-Validation-1.xml", policies.AddOpenAPIValidatePolicy(fileName)); err != nil {
			return err
		}
	}

	if err = archiveBundle(rootDir, "/tmp/"+name+".zip"); err != nil {
		return err
	}

	defer os.RemoveAll(rootDir) // clean up
	return nil
}

func writeXMLData(fileName string, data string) error {
	fileWriter, err := os.Create(fileName)
	if err != nil {
		return err
	}
	_, err = fileWriter.WriteString(data)
	if err != nil {
		return err
	}

	fileWriter.Close()
	return nil
}

func archiveBundle(pathToZip, destinationPath string) error {
	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	myZip := zip.NewWriter(destinationFile)
	err = filepath.Walk(pathToZip, func(filePath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if err != nil {
			return err
		}
		relPath := strings.TrimPrefix(filePath, filepath.Dir(pathToZip))
		zipFile, err := myZip.Create(relPath)
		if err != nil {
			return err
		}
		fsFile, err := os.Open(filePath)
		if err != nil {
			return err
		}
		_, err = io.Copy(zipFile, fsFile)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = myZip.Close()
	if err != nil {
		return err
	}
	return nil
}
