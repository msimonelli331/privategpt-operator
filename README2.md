# PrivateGPT Kube Operator

`repo=https://github.com/msimonelli331/privategpt-operator`
```bash
export GOPATH=/workspaces/privategpt-operator
mkdir -p $GOPATH/src/github.com/msimonelli331/privategpt-operator
cd $GOPATH/src/github.com/msimonelli331/privategpt-operator
kubebuilder init --domain eirl --repo=github.com/msimonelli331/privategpt-operator
kubebuilder create api --group=privategpt --version=v1alpha1 --kind=PrivateGPTInstance
# Search for EDIT THIS FILE
# Modified src/github.com/msimonelli331/privategpt-operator/api/v1alpha1/privategptinstance_types.go
make manifests
# Search for TODO(user)
# Modified src/github.com/msimonelli331/privategpt-operator/internal/controller/privategptinstance_controller.go
make docker-build docker-push IMG=ghcr.io/msimonelli331/privategpt-operator:latest
```

## Resources

- https://medium.com/developingnodes/mastering-kubernetes-operators-your-definitive-guide-to-starting-strong-70ff43579eb9
- https://book.kubebuilder.io/quick-start.html#installation
- https://book.kubebuilder.io/getting-started.html?highlight=appsv1.Deployment
- https://github.com/argoproj-labs/argocd-operator/blob/master/.github/workflows/ci-build.yaml