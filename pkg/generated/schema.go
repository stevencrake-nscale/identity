// Package generated provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.12.4 DO NOT EDIT.
package generated

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	var res = make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	var resolvePath = PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		var pathToFile = url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
