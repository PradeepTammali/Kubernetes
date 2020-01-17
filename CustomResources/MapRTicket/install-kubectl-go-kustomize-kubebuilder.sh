# Kubectl installation

sudo apt-get update && sudo apt-get install -y apt-transport-https
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee -a /etc/apt/sources.list.d/kubernetes.list
sudo apt-get update
sudo apt-get install -y kubectl


# Go installation
wget https://dl.google.com/go/go1.13.6.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.13.6.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
rm go*


# Kustomize installation 
wget https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv3.5.4/kustomize_v3.5.4_linux_amd64.tar.gz
tar -C /usr/local/bin -xzf kustomize_v3.5.4_linux_amd64.tar.gz
rm kustomize_*


# kubebuilder installtion 
os=$(go env GOOS)
arch=$(go env GOARCH)
curl -L https://go.kubebuilder.io/dl/2.2.0/${os}/${arch} | tar -xz -C /tmp/
sudo mv /tmp/kubebuilder_2.2.0_${os}_${arch} /usr/local/kubebuilder
export PATH=$PATH:/usr/local/kubebuilder/bin


# Reference:
# https://book.kubebuilder.io/quick-start.html
