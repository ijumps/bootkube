// Package internal holds asset templates used by bootkube.
package internal

var (
	KubeConfigTemplate = []byte(`apiVersion: v1
kind: Config
clusters:
- name: local
  cluster:
    server: {{ .Server }}
    certificate-authority-data: {{ .CACert }}
users:
- name: kubelet
  user:
    client-certificate-data: {{ .KubeletCert}}
    client-key-data: {{ .KubeletKey }}
contexts:
- context:
    cluster: local
    user: kubelet
`)

	KubeletTemplate = []byte(`apiVersion: extensions/v1beta1
kind: DaemonSet
metadata:
  name: kubelet
  namespace: kube-system
  labels:
    k8s-app: kubelet
spec:
  template:
    metadata:
      labels:
        k8s-app: kubelet
    spec:
      containers:
      - name: kubelet
        image: cargo.caicloud.io/caicloud/hyperkube:v0.11.2
        command:
        - ./hyperkube
        - kubelet
        - --network-plugin=cni
        - --cni-conf-dir=/etc/kubernetes/cni/net.d
        - --cni-bin-dir=/opt/cni/bin
        - --pod-manifest-path=/etc/kubernetes/manifests
        - --allow-privileged
        - --hostname-override=$(NODE_NAME)
        - --cluster-dns=10.3.0.10
        - --cluster-domain=cluster.local
        - --kubeconfig=/etc/kubernetes/kubeconfig
        - --require-kubeconfig
        - --lock-file=/var/run/lock/kubelet.lock
        - --containerized
        - --pod-infra-container-image=cargo.caicloud.io/caicloudgcr/google_containers_pause-amd64:3.0
        env:
          - name: NODE_NAME
            valueFrom:
              fieldRef:
                fieldPath: spec.nodeName
        securityContext:
          privileged: true
        volumeMounts:
        - name: dev
          mountPath: /dev
        - name: run
          mountPath: /run
        - name: sys
          mountPath: /sys
          readOnly: true
        - name: etc-kubernetes
          mountPath: /etc/kubernetes
          readOnly: true
        - name: etc-ssl-certs
          mountPath: /etc/ssl/certs
          readOnly: true
        - name: var-lib-docker
          mountPath: /var/lib/docker
        - name: var-lib-kubelet
          mountPath: /var/lib/kubelet
        - name: var-lib-rkt
          mountPath: /var/lib/rkt
        - name: rootfs
          mountPath: /rootfs
      hostNetwork: true
      hostPID: true
      volumes:
      - name: dev
        hostPath:
          path: /dev
      - name: run
        hostPath:
          path: /run
      - name: sys
        hostPath:
          path: /sys
      - name: etc-kubernetes
        hostPath:
          path: /etc/kubernetes
      - name: etc-ssl-certs
        hostPath:
          path: /usr/share/ca-certificates
      - name: var-lib-docker
        hostPath:
          path: /var/lib/docker
      - name: var-lib-kubelet
        hostPath:
          path: /var/lib/kubelet
      - name: var-lib-rkt
        hostPath:
          path: /var/lib/rkt
      - name: rootfs
        hostPath:
          path: /
`)

	APIServerTemplate = []byte(`apiVersion: "extensions/v1beta1"
kind: DaemonSet
metadata:
  name: kube-apiserver
  namespace: kube-system
  labels:
    k8s-app: kube-apiserver
spec:
  template:
    metadata:
      labels:
        k8s-app: kube-apiserver
      annotations:
        checkpointer.alpha.coreos.com/checkpoint: "true"
    spec:
      nodeSelector:
        master: "true"
      hostNetwork: true
      containers:
      - name: kube-apiserver
        image: cargo.caicloud.io/caicloud/hyperkube:v0.11.2
        command:
        - /hyperkube
        - apiserver
        - --bind-address=0.0.0.0
        - --secure-port=443
        - --insecure-port=8080
        - --etcd-servers={{ range $i, $e := .EtcdServers }}{{ if $i }},{{end}}{{ $e }}{{end}}
        - --storage-backend={{.StorageBackend}}
        - --allow-privileged=true
        - --service-cluster-ip-range=10.3.0.0/24
        - --admission-control=NamespaceLifecycle,LimitRanger,ServiceAccount,ResourceQuota
        - --runtime-config=api/all=true
        - --tls-cert-file=/etc/kubernetes/secrets/apiserver.crt
        - --tls-private-key-file=/etc/kubernetes/secrets/apiserver.key
        - --service-account-key-file=/etc/kubernetes/secrets/service-account.pub
        - --client-ca-file=/etc/kubernetes/secrets/ca.crt
        - --cloud-provider={{ .CloudProvider  }}
        - --anonymous-auth=false
        volumeMounts:
        - mountPath: /etc/ssl/certs
          name: ssl-certs-host
          readOnly: true
        - mountPath: /etc/kubernetes/secrets
          name: secrets
          readOnly: true
      volumes:
      - name: ssl-certs-host
        hostPath:
          path: /usr/share/ca-certificates
      - name: secrets
        secret:
          secretName: kube-apiserver
`)
	CheckpointerTemplate = []byte(`apiVersion: "extensions/v1beta1"
kind: DaemonSet
metadata:
  name: checkpoint-installer
  namespace: kube-system
  labels:
    k8s-app: pod-checkpoint-installer
spec:
  template:
    metadata:
      labels:
        k8s-app: pod-checkpoint-installer
    spec:
      nodeSelector:
        master: "true"
      hostNetwork: true
      containers:
      - name: checkpoint-installer
        image: quay.io/coreos/pod-checkpointer:5f0ba7be7c958eac22cfd3d1e8e5b004bc6988a0
        command:
        - /checkpoint-installer.sh
        volumeMounts:
        - mountPath: /etc/kubernetes/manifests
          name: etc-k8s-manifests
      volumes:
      - name: etc-k8s-manifests
        hostPath:
          path: /etc/kubernetes/manifests
`)
	ControllerManagerTemplate = []byte(`apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: kube-controller-manager
  namespace: kube-system
  labels:
    k8s-app: kube-controller-manager
spec:
  replicas: 2
  template:
    metadata:
      labels:
        k8s-app: kube-controller-manager
    spec:
      containers:
      - name: kube-controller-manager
        image: cargo.caicloud.io/caicloud/hyperkube:v0.11.2
        command:
        - ./hyperkube
        - controller-manager
        - --allocate-node-cidrs=true
        - --configure-cloud-routes=false
        - --cluster-cidr=10.2.0.0/16
        - --root-ca-file=/etc/kubernetes/secrets/ca.crt
        - --service-account-private-key-file=/etc/kubernetes/secrets/service-account.key
        - --leader-elect=true
        - --cloud-provider={{ .CloudProvider  }}
        - --configure-cloud-routes=false
        volumeMounts:
        - name: secrets
          mountPath: /etc/kubernetes/secrets
          readOnly: true
        - name: ssl-host
          mountPath: /etc/ssl/certs
          readOnly: true
      volumes:
      - name: secrets
        secret:
          secretName: kube-controller-manager
      - name: ssl-host
        hostPath:
          path: /usr/share/ca-certificates
      dnsPolicy: Default # Don't use cluster DNS.
`)
	SchedulerTemplate = []byte(`apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: kube-scheduler
  namespace: kube-system
  labels:
    k8s-app: kube-scheduler
spec:
  replicas: 2
  template:
    metadata:
      labels:
        k8s-app: kube-scheduler
    spec:
      containers:
      - name: kube-scheduler
        image: cargo.caicloud.io/caicloud/hyperkube:v0.11.2
        command:
        - ./hyperkube
        - scheduler
        - --leader-elect=true
`)
	ProxyTemplate = []byte(`apiVersion: "extensions/v1beta1"
kind: DaemonSet
metadata:
  name: kube-proxy
  namespace: kube-system
  labels:
    k8s-app: kube-proxy
spec:
  template:
    metadata:
      labels:
        k8s-app: kube-proxy
    spec:
      hostNetwork: true
      containers:
      - name: kube-proxy
        image: cargo.caicloud.io/caicloud/hyperkube:v0.11.2
        command:
        - /hyperkube
        - proxy
        - --kubeconfig=/etc/kubernetes/kubeconfig
        - --proxy-mode=iptables
        - --hostname-override=$(NODE_NAME)
        - --cluster-cidr=10.2.0.0/16
        env:
          - name: NODE_NAME
            valueFrom:
              fieldRef:
                fieldPath: spec.nodeName
        securityContext:
          privileged: true
        volumeMounts:
        - mountPath: /etc/ssl/certs
          name: ssl-certs-host
          readOnly: true
        - name: etc-kubernetes
          mountPath: /etc/kubernetes
          readOnly: true
      volumes:
      - hostPath:
          path: /usr/share/ca-certificates
        name: ssl-certs-host
      - name: etc-kubernetes
        hostPath:
          path: /etc/kubernetes
`)
	DNSDeploymentTemplate = []byte(`apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: kube-dns
  namespace: kube-system
  labels:
    k8s-app: kube-dns
    kubernetes.io/cluster-service: "true"
spec:
  # replicas: not specified here:
  # 1. In order to make Addon Manager do not reconcile this replicas parameter.
  # 2. Default is 1.
  # 3. Will be tuned in real time if DNS horizontal auto-scaling is turned on.
  strategy:
    rollingUpdate:
      maxSurge: 10%
      maxUnavailable: 0
  selector:
    matchLabels:
      k8s-app: kube-dns
  template:
    metadata:
      labels:
        k8s-app: kube-dns
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
        scheduler.alpha.kubernetes.io/tolerations: '[{"key":"CriticalAddonsOnly", "operator":"Exists"}]'
    spec:
      containers:
      - name: kubedns
        image: cargo.caicloud.io/caicloudgcr/google_containers_kubedns-amd64:1.9
        resources:
          # TODO: Set memory limits when we've profiled the container for large
          # clusters, then set request = limit to keep this container in
          # guaranteed class. Currently, this container falls into the
          # "burstable" category so the kubelet doesn't backoff from restarting it.
          limits:
            memory: 170Mi
          requests:
            cpu: 100m
            memory: 70Mi
        livenessProbe:
          httpGet:
            path: /healthz-kubedns
            port: 8080
            scheme: HTTP
          initialDelaySeconds: 60
          timeoutSeconds: 5
          successThreshold: 1
          failureThreshold: 5
        readinessProbe:
          httpGet:
            path: /readiness
            port: 8081
            scheme: HTTP
          # we poll on pod startup for the Kubernetes master service and
          # only setup the /readiness HTTP server once that's available.
          initialDelaySeconds: 3
          timeoutSeconds: 5
        args:
        - --domain=cluster.local.
        - --dns-port=10053
        - --config-map=kube-dns
        # This should be set to v=2 only after the new image (cut from 1.5) has
        # been released, otherwise we will flood the logs.
        - --v=0
        env:
        - name: PROMETHEUS_PORT
          value: "10055"
        ports:
        - containerPort: 10053
          name: dns-local
          protocol: UDP
        - containerPort: 10053
          name: dns-tcp-local
          protocol: TCP
        - containerPort: 10055
          name: metrics
          protocol: TCP
      - name: dnsmasq
        image: cargo.caicloud.io/caicloudgcr/google_containers_kube-dnsmasq-amd64:1.4
        livenessProbe:
          httpGet:
            path: /healthz-dnsmasq
            port: 8080
            scheme: HTTP
          initialDelaySeconds: 60
          timeoutSeconds: 5
          successThreshold: 1
          failureThreshold: 5
        args:
        - --cache-size=1000
        - --no-resolv
        - --server=127.0.0.1#10053
        - --log-facility=-
        ports:
        - containerPort: 53
          name: dns
          protocol: UDP
        - containerPort: 53
          name: dns-tcp
          protocol: TCP
        # see: https://github.com/kubernetes/kubernetes/issues/29055 for details
        resources:
          requests:
            cpu: 150m
            memory: 10Mi
      - name: dnsmasq-metrics
        image: cargo.caicloud.io/caicloudgcr/google_containers_dnsmasq-metrics-amd64:1.0
        livenessProbe:
          httpGet:
            path: /metrics
            port: 10054
            scheme: HTTP
          initialDelaySeconds: 60
          timeoutSeconds: 5
          successThreshold: 1
          failureThreshold: 5
        args:
        - --v=2
        - --logtostderr
        ports:
        - containerPort: 10054
          name: metrics
          protocol: TCP
        resources:
          requests:
            memory: 10Mi
      - name: healthz
        image: cargo.caicloud.io/caicloudgcr/google_containers_exechealthz-amd64:1.2
        resources:
          limits:
            memory: 50Mi
          requests:
            cpu: 10m
            # Note that this container shouldn't really need 50Mi of memory. The
            # limits are set higher than expected pending investigation on #29688.
            # The extra memory was stolen from the kubedns container to keep the
            # net memory requested by the pod constant.
            memory: 50Mi
        args:
        - --cmd=nslookup kubernetes.default.svc.cluster.local 127.0.0.1 >/dev/null
        - --url=/healthz-dnsmasq
        - --cmd=nslookup kubernetes.default.svc.cluster.local 127.0.0.1:10053 >/dev/null
        - --url=/healthz-kubedns
        - --port=8080
        - --quiet
        ports:
        - containerPort: 8080
          protocol: TCP
      dnsPolicy: Default  # Don't use cluster DNS.
`)
	DNSSvcTemplate = []byte(`apiVersion: v1
kind: Service
metadata:
  name: kube-dns
  namespace: kube-system
  labels:
    k8s-app: kube-dns
    kubernetes.io/cluster-service: "true"
    kubernetes.io/name: "KubeDNS"
spec:
  selector:
    k8s-app: kube-dns
  clusterIP: 10.3.0.10
  ports:
  - name: dns
    port: 53
    protocol: UDP
  - name: dns-tcp
    port: 53
    protocol: TCP
`)

	EtcdOperatorTemplate = []byte(`apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: etcd-operator
  namespace: kube-system
  labels:
    k8s-app: etcd-operator
spec:
  replicas: 1
  template:
    metadata:
      labels:
        k8s-app: etcd-operator
    spec:
      containers:
      - name: etcd-operator
        image: quay.io/coreos/etcd-operator
        env:
        - name: MY_POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
`)

	EtcdSvcTemplate = []byte(`apiVersion: v1
kind: Service
metadata:
  name: etcd-service
  namespace: kube-system
spec:
  selector:
    app: etcd
    etcd_cluster: etcd-cluster
  clusterIP: 10.3.0.15
  ports:
  - name: client
    port: 2379
    protocol: TCP
`)

	KubeFlannelCfgTemplate = []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: kube-flannel-cfg
  namespace: kube-system
  labels:
    tier: node
    app: flannel
data:
  cni-conf.json: |
    {
      "name": "cbr0",
      "type": "flannel",
      "delegate": {
        "isDefaultGateway": true
      }
    }
  net-conf.json: |
    {
      "Network": "10.2.0.0/16",
      "Backend": {
        "Type": "vxlan"
      }
    }
`)

	KubeFlannelTemplate = []byte(`apiVersion: extensions/v1beta1
kind: DaemonSet
metadata:
  name: kube-flannel
  namespace: kube-system
  labels:
    tier: node
    app: flannel
spec:
  template:
    metadata:
      labels:
        tier: node
        app: flannel
    spec:
      hostNetwork: true
      containers:
      - name: kube-flannel
        image: quay.io/coreos/flannel-git:v0.6.1-60-g4b752d1-amd64
        command: [ "/opt/bin/flanneld", "--ip-masq", "--kube-subnet-mgr", "--iface=$(POD_IP)"]
        securityContext:
          privileged: true
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: POD_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        volumeMounts:
        - name: run
          mountPath: /run
        - name: cni
          mountPath: /etc/cni/net.d
        - name: flannel-cfg
          mountPath: /etc/kube-flannel/
      - name: install-cni
        image: busybox
        command: [ "/bin/sh", "-c", "set -e -x; TMP=/etc/cni/net.d/.tmp-flannel-cfg; cp /etc/kube-flannel/cni-conf.json ${TMP}; mv ${TMP} /etc/cni/net.d/10-flannel.conf; while :; do sleep 3600; done" ]
        volumeMounts:
        - name: cni
          mountPath: /etc/cni/net.d
        - name: flannel-cfg
          mountPath: /etc/kube-flannel/
      volumes:
        - name: run
          hostPath:
            path: /run
        - name: cni
          hostPath:
            path: /etc/kubernetes/cni/net.d
        - name: flannel-cfg
          configMap:
            name: kube-flannel-cfg
`)
)
