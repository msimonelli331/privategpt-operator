#!/bin/bash
set -x

install_kind() {
    curl -Lo ./kind https://kind.sigs.k8s.io/dl/latest/kind-linux-$(go env GOARCH)
    chmod +x ./kind
    mv ./kind /usr/local/bin/kind

    kind version
}

install_kubebuilder() {
    curl -L -o kubebuilder https://go.kubebuilder.io/dl/latest/linux/$(go env GOARCH)
    chmod +x kubebuilder
    mv kubebuilder /usr/local/bin/

    kubebuilder version
}

install_kubectl() {
    KUBECTL_VERSION=$(curl -L -s https://dl.k8s.io/release/stable.txt)
    curl -LO "https://dl.k8s.io/release/$KUBECTL_VERSION/bin/linux/$(go env GOARCH)/kubectl"
    chmod +x kubectl
    mv kubectl /usr/local/bin/kubectl

    kubectl version --client
}

install_npm() {
    curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.3/install.sh | bash
    # in lieu of restarting the shell
    \. "$HOME/.nvm/nvm.sh"
    nvm install 24

    node -v
    npm -v
}

install_cline() {
    install_npm
    apt update -y
    apt install -y python3 make gcc g++
    npm install -g cline

    cline version
}

install_helm() {
    curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-4 | bash

    helm version
}

docker --version
go version
install_kind
install_kubebuilder
install_kubectl
install_npm
install_cline
install_helm

docker network create -d=bridge --subnet=172.19.0.0/24 kind
