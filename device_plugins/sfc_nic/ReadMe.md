# Solarflare Device Plugin
---------------------------

### Steps to deploy
    $ git clone http:
    $ cd kubernetes
    $ docker build -t sfc-dev-plugin -f device_plugins/sfc_nic/Dockerfile .
 Adjust the config map parameters for onload configuration:

    $ cat device_plugins/sfc_nic/device_plugin.yml
    ---
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: configmap
    data:
      onload-version: 201606-u1.3
      reg-exp-sfc: (?m)[\r\n]+^.*SFC[6-9].*$
      socket-name: sfcNIC
      resource-name: pod.alpha.kubernetes.io/opaque-int-resource-sfcNIC
      k8s-api: https://<master-ip>:6443
      node-label-onload-version: device.sfc.onload-version
  And then deploy the daemonsets:

    $ kubectl apply -f device_plugins/sfc_nic/device-plugin.yml


### Verify if NICs got picked up by plugin and reported fine to kubelet

    [root@dell-r620-01 kubernetes]# kubectl get nodes -o json | jq     '.items[0].status.capacity'
    {
    "cpu": "16",
    "memory": "131816568Ki",
    "pod.alpha.kubernetes.io/opaque-int-resource-sfcNIC": "2",
    "pods": "110"
    }

### sample pod template to consume SFC NICs
    apiVersion: v1
    kind: Pod
    metadata:
        name: my.pod1
    spec:
        containers:
        - name: demo1
        image: sfc-dev-plugin:latest
        imagePullPolicy: Never
        resources:
            requests:
                pod.alpha.kubernetes.io/opaque-int-resource-sfcNIC: '1'
            limits:
                pod.alpha.kubernetes.io/opaque-int-resource-sfcNIC: '1'

