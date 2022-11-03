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

package create

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/eschercloudai/unikorn/generated/clientset/unikorn"
	unikornv1alpha1 "github.com/eschercloudai/unikorn/pkg/apis/unikorn/v1alpha1"
	"github.com/eschercloudai/unikorn/pkg/cmd/errors"
	"github.com/eschercloudai/unikorn/pkg/cmd/util"
	"github.com/eschercloudai/unikorn/pkg/constants"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/templates"
)

type createProjectOptions struct {
	// name is the name of the project to create.
	name string

	// projectID is the external management plane's project identifier.
	projectID string

	// client is the Kubernetes v1 client.
	client kubernetes.Interface

	// unikornClient gives access to our custom resources.
	unikornClient unikorn.Interface
}

// addFlags registers create cluster options flags with the specified cobra command.
func (o *createProjectOptions) addFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.projectID, "project", "", "management plane project identifier.")

	if err := cmd.MarkFlagRequired("project"); err != nil {
		panic(err)
	}
}

// complete fills in any options not does automatically by flag parsing.
func (o *createProjectOptions) complete(f cmdutil.Factory, args []string) error {
	var err error

	if o.client, err = f.KubernetesClientSet(); err != nil {
		return err
	}

	config, err := f.ToRESTConfig()
	if err != nil {
		return err
	}

	if o.unikornClient, err = unikorn.NewForConfig(config); err != nil {
		return err
	}

	if len(args) != 1 {
		return errors.ErrIncorrectArgumentNum
	}

	o.name = args[0]

	return nil
}

// validate validates any tainted input not handled by complete() or flags
// processing.
func (o *createProjectOptions) validate() error {
	if len(o.name) == 0 {
		return errors.ErrInvalidName
	}

	return nil
}

// run executes the command.
func (o *createProjectOptions) run() error {
	project := &unikornv1alpha1.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: o.name,
			Labels: map[string]string{
				constants.VersionLabel: constants.Version,
			},
			Finalizers: []string{
				metav1.FinalizerDeleteDependents,
			},
		},
		Spec: unikornv1alpha1.ProjectSpec{
			ProjectID: o.projectID,
		},
	}

	if _, err := o.unikornClient.UnikornV1alpha1().Projects().Create(context.TODO(), project, metav1.CreateOptions{}); err != nil {
		return err
	}

	return nil
}

var (
	//nolint:gochecknoglobals
	createProjectLong = templates.LongDesc(`
        Create a project.

	Projects are modelled as custom resources, as they are domain specific
	abstractions.  We tried initially modelling control planes on namespaces,
	with projects being an annotations, but it turns out this is a whole
	world of pain.

	Projects map 1:1 to a namespace, and these project namespaces also contain
	custom control plane resources.  Thus we can simply off-board users/projects
	with a single delete.  We also no longer need to do an indexed search of
	control planes by project, as control planes are encapsulated in projects.

	Projects are cluster scoped.`)

	//nolint:gochecknoglobals
	createProjectExample = util.TemplatedExample(`
        # Create a control plane named my-project-name.
        {{.Application}} create project my-project-name`)
)

// newCreateProjectCommand creates a command that can create control planes.
// The initial intention is to have a per-user/organization control plane that
// contains Cluster API in a virtual cluster.
func newCreateProjectCommand(f cmdutil.Factory) *cobra.Command {
	o := &createProjectOptions{}

	cmd := &cobra.Command{
		Use:     "project [flags] my-project-name",
		Short:   "Create a project.",
		Long:    createProjectLong,
		Example: createProjectExample,
		Run: func(cmd *cobra.Command, args []string) {
			util.AssertNilError(o.complete(f, args))
			util.AssertNilError(o.validate())
			util.AssertNilError(o.run())
		},
	}

	o.addFlags(cmd)

	return cmd
}