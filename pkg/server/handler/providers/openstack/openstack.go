/*
Copyright 2022-2023 EscherCloud.

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

package openstack

import (
	"net/http"
	"sync"
	"time"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/applicationcredentials"
	lru "github.com/hashicorp/golang-lru/v2"
	"golang.org/x/sync/errgroup"

	"github.com/eschercloudai/unikorn/pkg/providers/openstack"
	"github.com/eschercloudai/unikorn/pkg/server/authorization"
	"github.com/eschercloudai/unikorn/pkg/server/errors"
	"github.com/eschercloudai/unikorn/pkg/server/generated"
)

// Openstack provides an HTTP handler for Openstack resources.
type Openstack struct {
	endpoint string

	// Cache clients as that's quite expensive.
	identityClientCache     *lru.Cache[string, *openstack.IdentityClient]
	computeClientCache      *lru.Cache[string, *openstack.ComputeClient]
	blockStorageClientCache *lru.Cache[string, *openstack.BlockStorageClient]
	networkClientCache      *lru.Cache[string, *openstack.NetworkClient]
}

// New returns a new initialized Openstack handler.
func New(authenticator *authorization.Authenticator) (*Openstack, error) {
	identityClientCache, err := lru.New[string, *openstack.IdentityClient](1024)
	if err != nil {
		return nil, err
	}

	computeClientCache, err := lru.New[string, *openstack.ComputeClient](1024)
	if err != nil {
		return nil, err
	}

	blockStorageClientCache, err := lru.New[string, *openstack.BlockStorageClient](1024)
	if err != nil {
		return nil, err
	}

	networkClientCache, err := lru.New[string, *openstack.NetworkClient](1024)
	if err != nil {
		return nil, err
	}

	o := &Openstack{
		endpoint:                authenticator.Endpoint(),
		identityClientCache:     identityClientCache,
		computeClientCache:      computeClientCache,
		blockStorageClientCache: blockStorageClientCache,
		networkClientCache:      networkClientCache,
	}

	return o, nil
}

func getToken(r *http.Request) (string, error) {
	claims, err := authorization.ClaimsFromContext(r.Context())
	if err != nil {
		return "", errors.OAuth2ServerError("failed get token claims").WithError(err)
	}

	if claims.UnikornClaims == nil {
		return "", errors.OAuth2ServerError("failed get token claim")
	}

	return claims.UnikornClaims.Token, nil
}

func getUser(r *http.Request) (string, error) {
	claims, err := authorization.ClaimsFromContext(r.Context())
	if err != nil {
		return "", errors.OAuth2ServerError("failed get token claims").WithError(err)
	}

	if claims.UnikornClaims == nil {
		return "", errors.OAuth2ServerError("failed get token claim")
	}

	return claims.UnikornClaims.User, nil
}

func (o *Openstack) IdentityClient(r *http.Request) (*openstack.IdentityClient, error) {
	token, err := getToken(r)
	if err != nil {
		return nil, err
	}

	if client, ok := o.identityClientCache.Get(token); ok {
		return client, nil
	}

	client, err := openstack.NewIdentityClient(openstack.NewTokenProvider(o.endpoint, token))
	if err != nil {
		return nil, errors.OAuth2ServerError("failed get identity client").WithError(err)
	}

	o.identityClientCache.Add(token, client)

	return client, nil
}

func (o *Openstack) ComputeClient(r *http.Request) (*openstack.ComputeClient, error) {
	token, err := getToken(r)
	if err != nil {
		return nil, err
	}

	if client, ok := o.computeClientCache.Get(token); ok {
		return client, nil
	}

	client, err := openstack.NewComputeClient(openstack.NewTokenProvider(o.endpoint, token))
	if err != nil {
		return nil, errors.OAuth2ServerError("failed get compute client").WithError(err)
	}

	o.computeClientCache.Add(token, client)

	return client, nil
}

func (o *Openstack) BlockStorageClient(r *http.Request) (*openstack.BlockStorageClient, error) {
	token, err := getToken(r)
	if err != nil {
		return nil, err
	}

	if client, ok := o.blockStorageClientCache.Get(token); ok {
		return client, nil
	}

	client, err := openstack.NewBlockStorageClient(openstack.NewTokenProvider(o.endpoint, token))
	if err != nil {
		return nil, errors.OAuth2ServerError("failed get block storage client").WithError(err)
	}

	o.blockStorageClientCache.Add(token, client)

	return client, nil
}

func (o *Openstack) NetworkClient(r *http.Request) (*openstack.NetworkClient, error) {
	token, err := getToken(r)
	if err != nil {
		return nil, err
	}

	if client, ok := o.networkClientCache.Get(token); ok {
		return client, nil
	}

	client, err := openstack.NewNetworkClient(openstack.NewTokenProvider(o.endpoint, token))
	if err != nil {
		return nil, errors.OAuth2ServerError("failed get network client").WithError(err)
	}

	o.networkClientCache.Add(token, client)

	return client, nil
}

func (o *Openstack) ListAvailabilityZonesCompute(r *http.Request) (generated.OpenstackAvailabilityZones, error) {
	client, err := o.ComputeClient(r)
	if err != nil {
		return nil, errors.OAuth2ServerError("failed get compute client").WithError(err)
	}

	result, err := client.AvailabilityZones(r.Context())
	if err != nil {
		return nil, errors.OAuth2ServerError("failed list availability zones").WithError(err)
	}

	azs := make(generated.OpenstackAvailabilityZones, len(result))

	for i, az := range result {
		azs[i].Name = az.ZoneName
	}

	return azs, nil
}

func (o *Openstack) ListAvailabilityZonesBlockStorage(r *http.Request) (generated.OpenstackAvailabilityZones, error) {
	client, err := o.BlockStorageClient(r)
	if err != nil {
		return nil, errors.OAuth2ServerError("failed get block storage client").WithError(err)
	}

	result, err := client.AvailabilityZones(r.Context())
	if err != nil {
		return nil, errors.OAuth2ServerError("failed list availability zones").WithError(err)
	}

	azs := make(generated.OpenstackAvailabilityZones, len(result))

	for i, az := range result {
		azs[i].Name = az.ZoneName
	}

	return azs, nil
}

func (o *Openstack) ListExternalNetworks(r *http.Request) (interface{}, error) {
	client, err := o.NetworkClient(r)
	if err != nil {
		return nil, errors.OAuth2ServerError("failed get network client").WithError(err)
	}

	result, err := client.ExternalNetworks(r.Context())
	if err != nil {
		return nil, errors.OAuth2ServerError("failed list external networks").WithError(err)
	}

	externalNetworks := make(generated.OpenstackExternalNetworks, len(result))

	for i, externalNetwork := range result {
		externalNetworks[i].Id = externalNetwork.Network.ID
		externalNetworks[i].Name = externalNetwork.Network.Name
	}

	return externalNetworks, nil
}

// convertFlavor traslates from Openstack's mess into our API types.
func convertFlavor(flavor *flavors.Flavor, extraSpecs map[string]string) (*generated.OpenstackFlavor, error) {
	f := &generated.OpenstackFlavor{
		Id:   flavor.ID,
		Name: flavor.Name,
		Cpus: flavor.VCPUs,
		// Convert MiB to GiB
		// TODO: Should probably specify units too to disambiguate.
		Memory: flavor.RAM >> 10,
	}

	gpu, ok, err := openstack.FlavorGPUs(flavor, extraSpecs)
	if err != nil {
		return nil, errors.OAuth2ServerError("unable to get GPU flavor metadata").WithError(err)
	}

	if ok {
		f.Gpus = &gpu.GPUs
	}

	return f, nil
}

func (o *Openstack) ListFlavors(r *http.Request) (generated.OpenstackFlavors, error) {
	client, err := o.ComputeClient(r)
	if err != nil {
		return nil, errors.OAuth2ServerError("failed get compute client").WithError(err)
	}

	result, err := client.Flavors(r.Context())
	if err != nil {
		return nil, errors.OAuth2ServerError("failed list flavors").WithError(err)
	}

	flavors := generated.OpenstackFlavors{}
	lock := &sync.Mutex{}

	// Openstack sucks at this, so we need to fan out and run concurrently.
	group, gctx := errgroup.WithContext(r.Context())

	for i := range result {
		f := &result[i]

		group.Go(func() error {
			extraSpecs, err := client.FlavorExtraSpecs(gctx, f)
			if err != nil {
				return errors.OAuth2ServerError("failed list flavor extra specs").WithError(err)
			}

			// Filter out baremetal nodes.
			if _, ok := extraSpecs["resources:CUSTOM_BAREMETAL"]; ok {
				return nil
			}

			flavor, err := convertFlavor(f, extraSpecs)
			if err != nil {
				return err
			}

			lock.Lock()
			defer lock.Unlock()

			flavors = append(flavors, *flavor)

			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return nil, err
	}

	return flavors, nil
}

// GetFlavor does a list and find, while inefficient, it does do image filtering.
func (o *Openstack) GetFlavor(r *http.Request, name string) (*generated.OpenstackFlavor, error) {
	flavors, err := o.ListFlavors(r)
	if err != nil {
		return nil, err
	}

	for i := range flavors {
		if flavors[i].Name == name {
			return &flavors[i], nil
		}
	}

	return nil, errors.HTTPNotFound()
}

func (o *Openstack) ListImages(r *http.Request) (generated.OpenstackImages, error) {
	client, err := o.ComputeClient(r)
	if err != nil {
		return nil, errors.OAuth2ServerError("failed get compute client").WithError(err)
	}

	result, err := client.Images(r.Context())
	if err != nil {
		return nil, errors.OAuth2ServerError("failed list images").WithError(err)
	}

	images := make(generated.OpenstackImages, len(result))

	for i, image := range result {
		created, err := time.Parse(time.RFC3339, image.Created)
		if err != nil {
			return nil, errors.OAuth2ServerError("failed parse image creation time").WithError(err)
		}

		modified, err := time.Parse(time.RFC3339, image.Updated)
		if err != nil {
			return nil, errors.OAuth2ServerError("failed parse image modification time").WithError(err)
		}

		// images are pre-filtered by the provider library, so these keys exist.
		kubernetesVersion, ok := image.Metadata["k8s"].(string)
		if !ok {
			return nil, errors.OAuth2ServerError("failed parse image kubernetes version")
		}

		nvidiaDriverVersion, ok := image.Metadata["gpu"].(string)
		if !ok {
			return nil, errors.OAuth2ServerError("failed parse image gpu driver version")
		}

		images[i].Id = image.ID
		images[i].Name = image.Name
		images[i].Created = created
		images[i].Modified = modified
		images[i].Versions.Kubernetes = "v" + kubernetesVersion
		images[i].Versions.NvidiaDriver = nvidiaDriverVersion
	}

	return images, nil
}

// GetImage does a list and find, while inefficient, it does do image filtering.
func (o *Openstack) GetImage(r *http.Request, name string) (*generated.OpenstackImage, error) {
	images, err := o.ListImages(r)
	if err != nil {
		return nil, err
	}

	for i := range images {
		if images[i].Name == name {
			return &images[i], nil
		}
	}

	return nil, errors.HTTPNotFound()
}

// ListAvailableProjects lists projects that the token has roles associated with.
func (o *Openstack) ListAvailableProjects(r *http.Request) (generated.OpenstackProjects, error) {
	client, err := o.IdentityClient(r)
	if err != nil {
		return nil, errors.OAuth2ServerError("failed get identity client").WithError(err)
	}

	result, err := client.ListAvailableProjects(r.Context())
	if err != nil {
		return nil, errors.OAuth2ServerError("failed list projects").WithError(err)
	}

	projects := make(generated.OpenstackProjects, len(result))

	for i, project := range result {
		projects[i].Id = project.ID
		projects[i].Name = project.Name

		if project.Description != "" {
			projects[i].Description = &result[i].Description
		}
	}

	return projects, nil
}

func (o *Openstack) ListKeyPairs(r *http.Request) (generated.OpenstackKeyPairs, error) {
	client, err := o.ComputeClient(r)
	if err != nil {
		return nil, errors.OAuth2ServerError("failed get compute client").WithError(err)
	}

	result, err := client.KeyPairs(r.Context())
	if err != nil {
		return nil, errors.OAuth2ServerError("failed list key pairs").WithError(err)
	}

	keyPairs := generated.OpenstackKeyPairs{}

	for _, keyPair := range result {
		// Undocumented (what a shocker), but absence means SSH as that's
		// all that used to be supported.  Obviously it could be something else
		// being odd that means we have to parse the public key...
		if keyPair.Type != "" && keyPair.Type != "ssh" {
			continue
		}

		k := generated.OpenstackKeyPair{
			Name: keyPair.Name,
		}

		keyPairs = append(keyPairs, k)
	}

	return keyPairs, nil
}

// convertApplicationCredential converts from Openstack's API to ours.
func convertApplicationCredential(in *applicationcredentials.ApplicationCredential) *generated.OpenstackApplicationCredential {
	out := &generated.OpenstackApplicationCredential{
		Id:   in.ID,
		Name: in.Name,
	}

	if in.Secret != "" {
		out.Secret = &in.Secret
	}

	return out
}

// findApplicationCredential, in the spirit of making the platform usable, allows
// a client to use names, rather than IDs for lookups.
func findApplicationCredential(in []applicationcredentials.ApplicationCredential, name generated.ApplicationCredentialParameter) (*applicationcredentials.ApplicationCredential, error) {
	for i, c := range in {
		if c.Name == name {
			return &in[i], nil
		}
	}

	return nil, errors.HTTPNotFound()
}

func (o *Openstack) GetApplicationCredential(r *http.Request, name generated.ApplicationCredentialParameter) (*generated.OpenstackApplicationCredential, error) {
	user, err := getUser(r)
	if err != nil {
		return nil, err
	}

	client, err := o.IdentityClient(r)
	if err != nil {
		return nil, errors.OAuth2ServerError("failed get identity client").WithError(err)
	}

	result, err := client.ListApplicationCredentials(r.Context(), user)
	if err != nil {
		return nil, errors.OAuth2ServerError("failed list application credentials").WithError(err)
	}

	match, err := findApplicationCredential(result, name)
	if err != nil {
		return nil, err
	}

	return convertApplicationCredential(match), nil
}

func (o *Openstack) CreateApplicationCredential(r *http.Request, options *generated.ApplicationCredentialOptions, roles []string) (*generated.OpenstackApplicationCredential, error) {
	user, err := getUser(r)
	if err != nil {
		return nil, err
	}

	client, err := o.IdentityClient(r)
	if err != nil {
		return nil, errors.OAuth2ServerError("failed get identity client").WithError(err)
	}

	description := "Automatically generated by platform service [DO NOT DELETE]."

	result, err := client.CreateApplicationCredential(r.Context(), user, options.Name, description, roles)
	if err != nil {
		return nil, err
	}

	return convertApplicationCredential(result), nil
}

func (o *Openstack) DeleteApplicationCredential(r *http.Request, name generated.ApplicationCredentialParameter) error {
	user, err := getUser(r)
	if err != nil {
		return err
	}

	client, err := o.IdentityClient(r)
	if err != nil {
		return errors.OAuth2ServerError("failed get identity client").WithError(err)
	}

	result, err := client.ListApplicationCredentials(r.Context(), user)
	if err != nil {
		return errors.OAuth2ServerError("failed list application credentials").WithError(err)
	}

	match, err := findApplicationCredential(result, name)
	if err != nil {
		return err
	}

	if err := client.DeleteApplicationCredential(r.Context(), user, match.ID); err != nil {
		return errors.OAuth2ServerError("failed delete application credentials").WithError(err)
	}

	return nil
}