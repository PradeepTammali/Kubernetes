
1. Please follow the documentation

   https://kind.sigs.k8s.io/docs/user/quick-start
2. To change the context use

```kubectl config use-context <context-name>```



To add admin.conf to existing config

```kind export kubeconfig```


#### To add IP address of the host in the api server using kubadm follow the procedure
##### Method 1:
1. Create cluster using the *dev-kind-config.yaml*
2. Find the reachable IP of host from the WSL.
3. Get kubeadm config with below command.
   
   ```kubectl -n kube-system get configmap kubeadm-config -o jsonpath='{.data.ClusterConfiguration}' > kubeadm.yaml```
4. Add Host IP in the "certSANs" section of *kubeadm.yaml* and move the *apiserver.crt*, *apiserver.key* to backup location.
5. Run the below command to regenerate certs of api server.(Have to run from inside the container as we don't have kubeadm in host)
   
   ```kubeadm init phase certs apiserver --config kubeadm.yaml```
6. To check whether the IP is been added api server certs.
   
   ```openssl x509 -in /etc/kubernetes/pki/apiserver.crt  -text```
7. To update in cluster configuration
   
   ```kubeadm config upload from-file --config kubeadm.yaml```
8. To verify the changes we can use 
   
   ```kubectl -n kube-system get configmap kubeadm-config -o yaml```

###### Reference: 

https://blog.scottlowe.org/2019/07/30/adding-a-name-to-kubernetes-api-server-certificate/

##### Method 2:
1. We can use custom-kubeadm.yaml configuration to direclty update the kubeadm configuration before creating the cluster.

###### Reference: 

https://kind.sigs.k8s.io/docs/user/quick-start/#enable-feature-gates-in-your-cluster
