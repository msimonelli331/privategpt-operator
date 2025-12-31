# privategpt-operator

// TODO(user): Add simple overview of use/purpose

## Description

// TODO(user): An in-depth paragraph about your project and overview of use

## Getting Started

### Prerequisites

- go version v1.24.6+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster

**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/privategpt-operator:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/privategpt-operator:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
> privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

> **NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall

**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

### Repo Setup Steps

Where the repo is `https://github.com/msimonelli331/privategpt-operator`, run the following:

```bash
mkdir -p $GOPATH/src/github.com/msimonelli331
ln -s $(pwd) $GOPATH/src/github.com/msimonelli331
kubebuilder init --domain eirl --repo=github.com/msimonelli331/privategpt-operator
kubebuilder create api --group=privategpt --version=v1alpha1 --kind=PrivateGPTInstance
# Search for EDIT THIS FILE
# Modified src/github.com/msimonelli331/privategpt-operator/api/v1alpha1/privategptinstance_types.go
make manifests
# Recreate zz_generated.deepcopy.go after changes to privategptinstance_types.go
make generate
# Search for TODO(user)
# Modified src/github.com/msimonelli331/privategpt-operator/internal/controller/privategptinstance_controller.go
make docker-build docker-push IMG=ghcr.io/msimonelli331/privategpt-operator:latest
```

After making changes to markers, like the RBAC markers, regen the yamls and helm

```bash
make manifests
kubebuilder edit --plugins=helm/v2-alpha
```

### Install Steps

1. Add this helm repo

   ```bash
   helm repo add privategpt-operator https://msimonelli331.github.io/privategpt-operator/
   ```

2. Install privategpt-operator

   ```bash
   helm install privategpt-operator privategpt-operator/privategpt-operator --create-namespace -n devops \
   --set privategpt.privateGPTInstance.ollamaURL=http://127.0.0.1:11434
   ```

### Test Steps

1. Create a test CRD that should trigger the creation of a privategpt instance

   ```bash
   cat > test-instance.yaml << EOF
   apiVersion: privategpt.eirl/v1alpha1
   kind: PrivateGPTInstance
   metadata:
     name: test-instance
     namespace: devops
   spec:
     ollamaURL: http://localhost:11434
     image: ghcr.io/msimonelli331/privategpt:latest
     domain: devops
   EOF
   kubectl apply -f test-instance.yaml
   ```

### Resources

- https://medium.com/developingnodes/mastering-kubernetes-operators-your-definitive-guide-to-starting-strong-70ff43579eb9
- https://book.kubebuilder.io/quick-start.html#installation
- https://book.kubebuilder.io/getting-started.html?highlight=appsv1.Deployment
- https://operatorhub.io/
- https://github.com/argoproj-labs/argocd-operator/blob/master/.github/workflows/ci-build.yaml
- https://nodejs.org/en/download/
- https://pkg.go.dev/k8s.io/api/networking/v1
- https://pkg.go.dev/k8s.io/api/networking/v1#PathType

## Project Distribution

Following the options to release and provide this solution to the users.

### By providing a bundle with all YAML files

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/privategpt-operator:tag
```

**NOTE:** The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without its
dependencies.

2. Using the installer

Users can just run 'kubectl apply -f <URL for YAML BUNDLE>' to install
the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/privategpt-operator/<tag or branch>/dist/install.yaml
```

### By providing a Helm Chart

1. Build the chart using the optional helm plugin

```sh
kubebuilder edit --plugins=helm/v2-alpha
```

2. See that a chart was generated under 'dist/chart', and users
   can obtain this solution from there.

**NOTE:** If you change the project, you need to update the Helm Chart
using the same command above to sync the latest changes. Furthermore,
if you create webhooks, you need to use the above command with
the '--force' flag and manually ensure that any custom configuration
previously added to 'dist/chart/values.yaml' or 'dist/chart/manager/manager.yaml'
is manually re-applied afterwards.

## Contributing

// TODO(user): Add detailed information on how you would like others to contribute to this project

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
