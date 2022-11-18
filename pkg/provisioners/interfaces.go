/*
Copyright 2022 EscherCloud.

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

package provisioners

import (
	"context"
)

// Provisioner is an abstract type that allows provisioning of Kubernetes
// packages in a technology agnostic way.  For example some things may be
// installed as a raw set of resources, a YAML manifest, Helm etc.
type Provisioner interface {
	// Provision deploys the requested package.
	// Implementations should ensure this receiver is idempotent.
	Provision(context.Context) error

	// Deprovision does any special handling of resource/component
	// removal.  In the general case, you should rely on cascading
	// deletion i.e. kill the namespace, use owner references.
	// Deprovisioners should be gating, waiting for their resources
	// to be removed before indicating success.
	Deprovision(context.Context) error
}

// ReadinessCheck is an abstract way of reasoning about the readiness of
// a component installed by a provisioner.  It's a way of providing a
// barrier essentially, as one thing may depend on another being deployed
// in order to function correctly.
type ReadinessCheck interface {
	// Check performs a single iteration of a readiness check.
	// Retries are delegated to the caller.
	Check(context.Context) error
}
