# Tài Liệu Kỹ Thuật Tetragon - Triển Khai Airgapped Kubernetes

**Tác giả:** Hồ Công Thành  
**Phiên bản:** 1.6.1  
**Ngày cập nhật:** 2026-04-15  
**Mục đích:** Triển khai Tetragon eBPF monitoring trên cụm K8s airgapped với ghi log tập trung đến Loki

---

## Mục Lục

1. [Giới Thiệu Tetragon](#giới-thiệu-tetragon)
2. [Kiến Trúc Hệ Thống](#kiến-trúc-hệ-thống)
3. [Yêu Cầu Hệ Thống & Tiên Quyết](#yêu-cầu-hệ-thống--tiên-quyết)
4. [Chuẩn Bị Triển Khai (Airgapped)](#chuẩn-bị-triển-khai-airgapped)
5. [Triển Khai Tetragon với Helm](#triển-khai-tetragon-với-helm)
6. [Cấu Hình Ghi Log Đến Loki](#cấu-hình-ghi-log-đến-loki)
7. [TracingPolicy - Định Nghĩa Rules](#tracingpolicy---định-nghĩa-rules)
8. [Ví Dụ TracingPolicy Thực Tế](#ví-dụ-tracingpolicy-thực-tế)
9. [Vận Hành & Theo Dõi](#vận-hành--theo-dõi)
10. [Khắc Phục Sự Cố](#khắc-phục-sự-cố)
11. [Tài Liệu Tham Khảo](#tài-liệu-tham-khảo)

---

## Giới Thiệu Tetragon

### Tetragon Là Gì?

**Tetragon** là một công cụ **eBPF-based runtime security** do Cilium phát triển, cung cấp khả năng:

- **Real-time monitoring** - Theo dõi tất cả hành động tại kernel level (file, network, process)
- **Fine-grained policies** - Định nghĩa TracingPolicy để allow/deny/warn các hành động cụ thể
- **Zero-Trust compliance** - Ghi lại đầy đủ hành động bị chặn hoặc cảnh báo
- **Low overhead** - eBPF native - CPU/Memory footprint cực thấp
- **Container native** - Tích hợp sâu với Kubernetes, hỗ trợ namespace isolation

### Các Trường Hợp Sử Dụng

| Trường Hợp | Mô Tả |
|-----------|-------|
| **Security Monitoring** | Phát hiện shell access, suspicious file writes, unusual network connections |
| **Compliance Logging** | Tuân thủ PCI-DSS, HIPAA - ghi log chi tiết mọi hành động trên system |
| **Incident Response** | Replay các events, trace root cause của security breach |
| **Runtime Enforcement** | Block/Warn các hành động vi phạm policy trước khi xảy ra |

---

## Kiến Trúc Hệ Thống

### Tổng Quan Kiến Trúc Triển Khai

```
┌────────────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster (Airgapped)              │
│                                                                │
│  ┌────────────────────────────────────────────────────────┐    │
│  │ Node 1 - Tetragon DaemonSet Pod                        │    │
│  │ ┌──────────────────────────────────────────────────┐   │    │
│  │ │ Tetragon Agent (eBPF)                            │   │    │
│  │ │ - monitors syscalls (execve, open, socket, etc)  │   │    │
│  │ │ - applies TracingPolicies                        │   │    │
│  │ │ - gathers events from kernel                     │   │    │
│  │ └────────────────────┬─────────────────────────────┘   │    │
│  │                      │                                      │
│  │ ┌────────────────────▼─────────────────────────────┐   │    │
│  │ │ Tetragon Events Processor                        │   │    │
│  │ │ - parses eBPF events                             │   │    │
│  │ │ - formats to JSON                                │   │    │
│  │ │ - applies filtering rules                        │   │    │
│  │ └────────────────────┬─────────────────────────────┘   │    │
│  │                      │                                 |    │
│  │ ┌────────────────────▼─────────────────────────────┐   │    │
│  │ │ Tetragon gRPC Server (port 54321)                │   │    │
│  │ │ - sends events to exporters                      │   │    │
│  │ └────────────────────────────────────────────────┘     │    │
│  └────────────────────────────────────────────────────────┘    │
│                                                                │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ Node 2-N - Tetragon DaemonSet Pods (similar)             │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ Tetragon Operator (optional - for CRD management)        │  │
│  │ - manages TracingPolicy CRDs                             │  │
│  │ - distributes policies to agents                         │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ Loki (Log aggregation backend)                           │  │
│  │ - receives events from Tetragon exporters                │  │
│  │ - stores and indexes logs with labels                    │  │
│  │ - available for queries via Grafana                      │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                │
└─────────────────────────────────────────┬──────────────────────┘
                                          │
                        ┌─────────────────▼───────────────┐
                        │  Grafana Dashboard              │
                        │ - visualize Tetragon events     │
                        │ - query Loki logs               │
                        │ - create alerts                 │
                        └─────────────────────────────────┘
```

### Event Flow

```
1. Syscall Event (e.g., file open)
   ↓
2. eBPF Hook captures event (kernel level)
   ↓
3. Tetragon Agent checks TracingPolicy
   ├─ Action: ALLOW? → log to stdout
   ├─ Action: DENY? → block event + log to stdout  
   └─ Action: WARN? → allow but mark as warned + log to stdout
   ↓
4. Event formatted as JSON + kubectl logs captured
   ↓
5. Log Exporter ships event to Loki
   (via fluent-bit, logstash, etc.)
   ↓
6. Loki stores with labels (pod, namespace, policy, action)
   ↓
7. Grafana queries Loki for visualization/alerting
```

### Các Thành Phần Chính

| Thành Phần | Description | Port | Resource |
|-----------|-------------|------|----------|
| **Tetragon Agent** | DaemonSet - chạy trên mỗi node, monitors syscalls | 54321 (gRPC) | 100m CPU / 200Mi RAM |
| **Tetragon Operator** | Deployment (optional) - manages CRDs | 8080 (HTTP) | 100m CPU / 256Mi RAM |
| **TracingPolicy CRDs** | K8s custom resources định nghĩa monitoring rules | - | - |
| **Loki** | Log aggregation backend (pre-existing) | 3100 (HTTP) | Per existing setup |
| **Grafana** | Frontend visualization (pre-existing) | 3000 (HTTP) | Per existing setup |

---

## Yêu Cầu Hệ Thống & Tiên Quyết

### Yêu Cầu Kubernetes

```yaml
Phiên bản Kubernetes: >= 1.21
- API version support: apiextensions.k8s.io/v1
- RBAC enabled: ✅ (required)
- Resource Quotas: Recommended
```

### Yêu Cầu Linux Kernel

| Tính Năng | Phiên Bản | Ghi Chú |
|----------|----------|--------|
| eBPF | >= 5.4 | Core eBPF support |
| kprobes | >= 5.4 | Function hooking |
| tracepoints | >= 4.7 | Kernel instrumentation |
| BTF (BPF Type Format) | >= 5.8 | Type information (recommended) |

**Kiểm tra Kernel:**
```bash
uname -r  # Should be >= 5.4
grep -i ebpf /boot/config-$(uname -r)  # Should have CONFIG_BPF=y
```

### Tiên Quyết

- [x] **Kubernetes cluster** đã running (airgapped hoặc không)
- [x] **Loki service** đã deployed và accessible
- [x] **Container registry** accessible từ cluster (MinIO/Harbor/Corporate Registry)
- [x] **helm** CLI installed (v3.x+)
- [x] **kubectl** configured với cluster access
- [x] **PVC provisioning** (optional - nếu lưu trữ local state)

### Yêu Cầu Network (Airgapped)

Trong môi trường airgapped, cần:

1. **Image Registry Access:** Tetragon images phải có sẵn
   - `quay.io/cilium/tetragon:v1.6.1` (Agent)
   - `quay.io/cilium/tetragon-operator:v1.6.1` (Operator)
   - `quay.io/cilium/hubble-ui:v0.x.x` (Optional - visualization)

2. **Loki Access:** Internal IP/hostname của Loki service
   - Không cần external connectivity
   - Dùng Kubernetes DNS: `loki.loki-ns.svc.cluster.local:3100`

3. **No outbound:** Block outbound network (OK - Tetragon not require external calls)

---

## Chuẩn Bị Triển Khai (Airgapped)

### Step 1: Pull và Push Images lên Registry

#### 1.1 Trên máy có internet access:

```bash
# Pull Tetragon images (chỉ cần images này)
REGISTRY_SOURCE="quay.io"
TETRAGON_VERSION="v1.6.1"

docker pull ${REGISTRY_SOURCE}/cilium/tetragon:${TETRAGON_VERSION}
docker pull ${REGISTRY_SOURCE}/cilium/tetragon-operator:${TETRAGON_VERSION}

# Khi có internet connection:
# Hoặc dùng podman, containerd tùy hệ thống
```

#### 1.2 Save images để chuyển:

```bash
# Save to tar files (khoảng 500MB mỗi file)
docker save quay.io/cilium/tetragon:v1.6.1 > tetragon-agent.tar.gz
docker save quay.io/cilium/tetragon-operator:v1.6.1 > tetragon-operator.tar.gz

# Chuyển files qua airgapped network (USB, secure transfer, etc.)
```

#### 1.3 Trong airgapped cluster - Load images:

```bash
# Load vào container runtime
docker load < tetragon-agent.tar.gz
docker load < tetragon-operator.tar.gz

# Hoặc push trực tiếp lên local registry (Harbor/MinIO)
docker tag quay.io/cilium/tetragon:v1.6.1 \
  harbor.internal.corp/tetragon/tetragon:v1.6.1

docker push harbor.internal.corp/tetragon/tetragon:v1.6.1
docker push harbor.internal.corp/tetragon/tetragon-operator:v1.6.1
```

### Step 2: Chuẩn Bị Helm Chart

#### 2.1 Extract Helm chart từ tarball:

```bash
cd /home/bigboss/vnpost/tetragon/
tar xzf tetragon-1.6.1.tgz
ls -la tetragon/  # Xem templates, Chart.yaml, values.yaml
```

#### 2.2 Xem default values:

```bash
cat tetragon/values.yaml | head -50
# Hoặc
helm show values ./tetragon/
```

### Step 3: Chuẩn Bị Namespace & RBAC

```bash
# Create namespace for Tetragon
kubectl create namespace tetragon
kubectl label namespace tetragon pod-security.kubernetes.io/enforce=privileged

# Verify
kubectl get ns tetragon -o yaml | grep labels -A 5
```

**Giải thích:**
- `pod-security.kubernetes.io/enforce=privileged` cần thiết vì Tetragon eBPF agent yêu cầu privileged access

### Step 4: Chuẩn Bị Credentials để Push Logs Loki (nếu cần)

```bash
# Nếu Loki có authentication
kubectl create secret generic loki-credentials \
  --from-literal=username=admin \
  --from-literal=password=<your-password> \
  -n tetragon
```

---

## Triển Khai Tetragon với Helm

### Step 1: Tạo Custom Values File

**File: `/home/bigboss/vnpost/tetragon/values-airgapped.yaml`**

```yaml
# Global settings
enabled: true
imagePullSecrets: []
# Tetragon agent settings
priorityClassName: ""
imagePullPolicy: IfNotPresent
serviceAccount:
  create: true
  annotations: {}
  name: ""
podAnnotations: {}
podSecurityContext: {}
nodeSelector: {}
tolerations:
  - operator: Exists
affinity: {}
extraHostPathMounts: []
extraConfigmapMounts: []
daemonSetAnnotations: {}
extraVolumes: []
updateStrategy: {}
podLabels: {}
daemonSetLabelsOverride: {}
selectorLabelsOverride: {}
podLabelsOverride: {}
serviceLabelsOverride: {}
# -- DNS policy for Tetragon pods.
#
# https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-s-dns-policy
dnsPolicy: Default
# -- Directory to put Tetragon JSON export files.
exportDirectory: "/var/run/cilium/tetragon"
# -- Configures whether Tetragon pods run on the host network.
#
# IMPORTANT: Tetragon must be on the host network for the process visibility to
# function properly.
hostNetwork: true
tetragon:
  enabled: true
  nameOverride: ""
  image:
    override: ~
    repository: quay.io/cilium/tetragon
    tag: v1.6.1
  resources: {}
  extraArgs: {}
  extraEnv: []
  # extraEnv:
  #   - name: foo
  #     value: bar
  podAnnotations:
    enabled: false
  extraVolumeMounts: []
  securityContext:
    privileged: true
  # -- Overrides the default livenessProbe for the tetragon container.
  livenessProbe: {}
  #  grpc:
  #    port: 54321

  # -- Tetragon puts processes in an LRU cache. The cache is used to find ancestors
  # for subsequently exec'ed processes.
  processCacheSize: 65536
  # -- If you want to run Tetragon in debug mode change this value to true
  debug: false
  # -- JSON export filename. Set it to an empty string to disable JSON export altogether.
  exportFilename: tetragon.log
  # -- JSON export file permissions as a string. Typically it's either "600" (to restrict access to
  # owner) or "640"/"644" (to allow read access by logs collector or another agent).
  exportFilePerm: "600"
  # -- Size in megabytes at which to rotate JSON export files.
  exportFileMaxSizeMB: 10
  # -- Number of rotated files to retain.
  exportFileMaxBackups: 5
  # -- Compress rotated JSON export files.
  exportFileCompress: false
  # -- Rate-limit event export (events per minute), Set to -1 to export all events.
  exportRateLimit: -1
  # -- Allowlist for JSON export. For example, to export only process_connect events from
  # the default namespace:
  #
  # exportAllowList: |
  #   {"namespace":["default"],"event_set":["PROCESS_EXEC"]}
  exportAllowList: |-
    {"event_set":["PROCESS_EXEC", "PROCESS_EXIT", "PROCESS_KPROBE", "PROCESS_UPROBE", "PROCESS_TRACEPOINT", "PROCESS_LSM"]}
  # -- Denylist for JSON export **(for file sinks only; does not filter gRPC output)**. For example, to exclude exec events that look similar to
  # Kubernetes health checks and all the events from kube-system namespace and the host:
  #
  # exportDenyList: |
  #   {"health_check":true}
  #   {"namespace":["kube-system",""]}
  #
  exportDenyList: |-
    {"health_check":true}
    {"namespace":["", "cilium", "kube-system"]}
  # -- Filters to include or exclude fields from Tetragon events. Without any filters, all
  # fields are included by default. The presence of at least one inclusion filter implies
  # default-exclude (i.e. any fields that don't match an inclusion filter will be
  # excluded). Field paths are expressed using dot notation like "a.b.c" and multiple
  # field paths can be separated by commas like "a.b.c,d,e.f". An optional "event_set" may
  # be specified to apply the field filter to a specific set of events.
  #
  # For example, to exclude the "parent" field from all events and include the "process"
  # field in PROCESS_KPROBE events while excluding all others:
  #
  # fieldFilters: |
  #   {"fields": "parent", "action": "EXCLUDE"}
  #   {"event_set": ["PROCESS_KPROBE"], "fields": "process", "action": "INCLUDE"}
  #
  fieldFilters: ""
  # -- Filters to redact secrets from the args fields in Tetragon events. To perform
  # redactions, redaction filters define RE2 regular expressions in the `redact`
  # field. Any capture groups in these RE2 regular expressions are redacted and
  # replaced with "*****".
  #
  # For more control, you can select which binary or binaries should have their
  # arguments redacted with the `binary_regex` field.
  #
  # NOTE: This feature uses RE2 as its regular expression library. Make sure that you follow
  # RE2 regular expression guidelines as you may observe unexpected results otherwise.
  # More information on RE2 syntax can be found [here](https://github.com/google/re2/wiki/Syntax).
  #
  # NOTE: When writing regular expressions in JSON, it is important to escape
  # backslash characters. For instance `\Wpasswd\W?` would be written as
  # `{"redact": "\\Wpasswd\\W?"}`.
  #
  # As a concrete example, the following will redact all passwords passed to
  # processes with the "--password" argument:
  #
  #   {"redact": ["--password(?:\\s+|=)(\\S*)"]}
  #
  # Now, an event which contains the string "--password=foo" would have that
  # string replaced with "--password=*****".
  #
  # Suppose we also see some passwords passed via the -p shorthand for a specific binary, foo.
  # We can also redact these as follows:
  #
  #   {"binary_regex": ["(?:^|/)foo$"], "redact": ["-p(?:\\s+|=)(\\S*)"]}
  #
  # With both of the above redaction filters in place, we are now redacting all
  # password arguments.
  redactionFilters: ""
  # -- Name of the cluster where Tetragon is installed. Tetragon uses this value
  # to set the cluster_name field in GetEventsResponse messages.
  clusterName: ""
  # -- Access Kubernetes API to associate Tetragon events with Kubernetes pods.
  enableK8sAPI: true
  # -- Enable Capabilities visibility in exec and kprobe events.
  enableProcessCred: false
  # -- Enable Namespaces visibility in exec and kprobe events.
  enableProcessNs: false
  processAncestors:
    # -- Comma-separated list of process event types to enable ancestors for.
    # Supported event types are: base, kprobe, tracepoint, uprobe, lsm, usdt. Unknown event types will be ignored.
    # Type "base" is required by all other supported event types for correct reference counting.
    # Set it to "" to disable ancestors completely.
    enabled: ""
  # -- Set --btf option to explicitly specify an absolute path to a btf file. For advanced users only.
  btf: ""
  # -- Override the command. For advanced users only.
  commandOverride: []
  # -- Override the arguments. For advanced users only.
  argsOverride: []
  prometheus:
    # -- Whether to enable exposing Tetragon metrics.
    enabled: true
    # -- The address at which to expose metrics. Set it to "" to expose on all available interfaces.
    address: ""
    # -- The port at which to expose metrics.
    port: 2112
    # -- Comma-separated list of enabled metrics labels.
    # The configurable labels are: namespace, workload, pod, binary. Unknown labels will be ignored.
    # Removing some labels from the list might help reduce the metrics cardinality if needed.
    metricsLabelFilter: "namespace,workload,pod,binary"
    serviceMonitor:
      # -- Whether to create a 'ServiceMonitor' resource targeting the tetragon pods.
      enabled: false
      # -- The set of labels to place on the 'ServiceMonitor' resource.
      labelsOverride: {}
      # -- Extra labels to be added on the Tetragon ServiceMonitor.
      extraLabels: {}
      # -- Interval at which metrics should be scraped. If not specified, Prometheus' global scrape interval is used.
      scrapeInterval: 60s
  grpc:
    # -- Whether to enable exposing Tetragon gRPC.
    enabled: true
    # -- The address at which to expose gRPC. Examples: localhost:54321, unix:///var/run/cilum/tetragon/tetragon.sock
    address: "localhost:54321"
  gops:
    # -- Whether to enable exposing gops server.
    enabled: true
    # -- The address at which to expose gops.
    address: "localhost"
    # -- The port at which to expose gops.
    port: 8118
  pprof:
    # -- Whether to enable exposing pprof server.
    enabled: false
    # -- The address at which to expose pprof.
    address: "localhost"
    # -- The port at which to expose pprof.
    port: 6060
  # -- Enable policy filter. This is required for K8s namespace and pod-label filtering.
  enablePolicyFilter: True
  # -- Enable policy filter cgroup map.
  enablePolicyFilterCgroupMap: false
  # -- Enable policy filter debug messages.
  enablePolicyFilterDebug: false
  # -- Enable latency monitoring in message handling
  enableMsgHandlingLatency: false
  healthGrpc:
    # -- Whether to enable health gRPC server.
    enabled: true
    # -- The port at which to expose health gRPC.
    port: 6789
    # -- The interval at which to check the health of the agent.
    interval: 10
  # -- Location of the host proc filesystem in the runtime environment. If the runtime runs in the
  # host, the path is /proc. Exceptions to this are environments like kind, where the runtime itself
  # does not run on the host.
  hostProcPath: "/proc"
  # -- Configure the number of retries in tetragon's event cache.
  eventCacheRetries: 15
  # -- Configure the delay (in seconds) between retires in tetragon's event cache.
  eventCacheRetryDelay: 2
  # -- Persistent enforcement to allow the enforcement policy to continue running even when its Tetragon process is gone.
  enableKeepSensorsOnExit: false
  # -- Configure the interval (suffixed with s for seconds, m for minutes, etc) for the process cache garbage collector.
  processCacheGCInterval: 30s
  # -- Configure tetragon pod so that it can contact the CRI running on the host
  cri:
    enabled: false
    # -- path of the CRI socket on the host. This will typically be
    # "/run/containerd/containerd.sock" for containerd or "/var/run/crio/crio.sock"  for crio.
    socketHostPath: ""
  # -- Enabling cgidmap instructs the Tetragon agent to use cgroup ids (instead of cgroup names) for
  # pod association. This feature depends on cri being enabled.
  cgidmap:
    enabled: false
  usePerfRingBuffer: false
# Tetragon Operator settings
tetragonOperator:
  # -- Enables the Tetragon Operator.
  enabled: true
  # -- The name of the Tetragon Operator deployment.
  nameOverride: ""
  # -- Number of replicas to run for the tetragon-operator deployment
  replicas: 1
  # -- Lease handling for an automated failover when running multiple replicas
  failoverLease:
    # -- Enable lease failover functionality
    enabled: false
    # -- Kubernetes Namespace in which the Lease resource is created. Defaults to the namespace where Tetragon is deployed in, if it's empty.
    namespace: ""
    # -- If a lease is not renewed for X duration, the current leader is considered dead, a new leader is picked
    leaseDuration: 15s
    # -- The interval at which the leader will renew the lease
    leaseRenewDeadline: 5s
    # -- The timeout between retries if renewal fails
    leaseRetryPeriod: 2s
  # -- Annotations for the Tetragon Operator Deployment.
  annotations: {}
  # -- Annotations for the Tetragon Operator Deployment Pods.
  podAnnotations: {}
  # -- Extra labels to be added on the Tetragon Operator Deployment.
  extraLabels: {}
  # -- Extra labels to be added on the Tetragon Operator Deployment Pods.
  extraPodLabels: {}
  # -- priorityClassName for the Tetragon Operator Deployment Pods.
  priorityClassName: ""
  # -- tetragon-operator service account.
  serviceAccount:
    create: true
    annotations: {}
    name: ""
  # -- securityContext for the Tetragon Operator Deployment Pods.
  podSecurityContext: {}
  # -- securityContext for the Tetragon Operator Deployment Pod container.
  containerSecurityContext:
    runAsUser: 65532
    runAsGroup: 65532
    runAsNonRoot: true
    allowPrivilegeEscalation: false
    capabilities:
      drop:
        - "ALL"
  # -- securityContext for the Tetragon Operator Deployment Pod container. (DEPRECATED: Use containerSecurityContext instead. TODO: Remove in v1.6.0)
  securityContext: {}
  # -- resources for the Tetragon Operator Deployment Pod container.
  resources:
    limits:
      cpu: 500m
      memory: 128Mi
    requests:
      cpu: 10m
      memory: 64Mi
  # -- resources for the Tetragon Operator Deployment update strategy
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  # -- Steer the Tetragon Operator Deployment Pod placement via nodeSelector, tolerations and affinity rules.
  nodeSelector: {}
  tolerations: []
  affinity:
    podAntiAffinity:
      preferredDuringSchedulingIgnoredDuringExecution:
        - weight: 100
          podAffinityTerm:
            topologyKey: kubernetes.io/hostname
            labelSelector:
              matchLabels:
                app.kubernetes.io/name: tetragon-operator
  # -- tetragon-operator image.
  image:
    override: ~
    repository: quay.io/cilium/tetragon-operator
    tag: v1.6.1
    pullPolicy: IfNotPresent
  # -- Extra volumes for the Tetragon Operator Deployment.
  extraVolumes: []
  extraVolumeMounts: []
  forceUpdateCRDs: false
  podInfo:
    # -- Enables the PodInfo CRD and the controller that reconciles PodInfo
    # custom resources.
    enabled: false
  tracingPolicy:
    # -- Enables the TracingPolicy and TracingPolicyNamespaced CRD creation.
    enabled: true
  prometheus:
    # -- Enables the Tetragon Operator metrics.
    enabled: true
    # -- The address at which to expose Tetragon Operator metrics. Set it to "" to expose on all available interfaces.
    address: ""
    # -- The port at which to expose metrics.
    port: 2113
    serviceMonitor:
      # -- Whether to create a 'ServiceMonitor' resource targeting the tetragonOperator pods.
      enabled: false
      # -- The set of labels to place on the 'ServiceMonitor' resource.
      labelsOverride: {}
      # -- Extra labels to be added on the Tetragon Operator ServiceMonitor.
      extraLabels: {}
      # -- Interval at which metrics should be scraped. If not specified, Prometheus' global scrape interval is used.
      scrapeInterval: 60s
# -- Tetragon events export settings
export:
  # "stdout". "" to disable.
  mode: "stdout"
  resources: {}
  securityContext: {}
  # filenames defines list of files for fluentd to tail and export.
  filenames:
    - tetragon.log
  # stdout specific exporter settings
  stdout:
    # -- Extra environment variables to add to the export-stdout container.
    # Example:
    # extraEnv:
    #   - name: FOO
    #     value: bar
    #   - name: SECRET_KEY
    #     valueFrom:
    #       secretKeyRef:
    #         name: my-secret
    #         key: secret-key
    extraEnv: []
    # -- Extra envFrom sources to add to the export-stdout container.
    # This allows adding any type of envFrom source (configMapRef, secretRef, etc.).
    # Example:
    # extraEnvFrom:
    #   - configMapRef:
    #       name: my-config-map
    #   - secretRef:
    #       name: my-secret
    #       optional: true
    extraEnvFrom: []
    # -- A simplified way to add secret references to envFrom.
    # Can be specified either as a string (just the secret name) or as an object with additional parameters.
    # Example:
    # envFromSecrets:
    #   - my-simple-secret
    #   - name: my-optional-secret
    #     optional: true
    envFromSecrets: []
    # * When enabledCommand=true and commandOverride is not set, the command inserted will be hubble-export-stdout.
    #   This supports the default for the current deployment instructions to deploy stdout-export sidecar container.
    # * When enabledCommand=true and commandOverride override is set, the command inserted will be the value of commandOverride.
    #   This is useful for inserting another sidecar container that requires a command override.
    # * When enabledCommand=false, no command will be specified in the manifest and container's default command will take over.
    enabledCommand: true
    # * When enabledArgs=true and argsOverride is not set, the args inserted will be the default ones for export-stdout.
    # * When enabledArgs=true and argsOverride override is set, the args value inserted will be the value of argsOverride.
    #   This is useful for inserting another sidecar container that requires args override.
    # * When enabledArgs=false, no command will be specified in the manifest and container's default args value will take over.
    enabledArgs: true
    # specific manifest command to use
    commandOverride: []
    # specific manifest args to use
    argsOverride: []
    # Extra volume mounts to add to stdout export pod
    extraVolumeMounts: []
    image:
      override: ~
      repository: quay.io/cilium/hubble-export-stdout
      tag: v1.1.0
crds:
  # -- Method for installing CRDs. Supported values are: "operator", "helm" and "none".
  # The "operator" method allows for fine-grained control over which CRDs are installed and by
  # default doesn't perform CRD downgrades. These can be configured in tetragonOperator section.
  # The "helm" method always installs all CRDs for the chart version.
  installMethod: "operator"
# -- Method for installing Tetagon rthooks (tetragon-rthooks) daemonset
# The tetragon-rthooks daemonset is responsible for installing run-time hooks on the host.
# See: https://tetragon.io/docs/concepts/runtime-hooks
rthooks:
  # -- Enable the Tetragon rthooks daemonset
  enabled: false
  # -- tetragon-rthooks name override
  nameOverride: ""
  # -- Method to use for installing  rthooks. Values:
  #
  #    "oci-hooks":
  #       Add an apppriate file to "/usr/share/containers/oci/hooks.d". Use this with CRI-O.
  #       See https://github.com/containers/common/blob/main/pkg/hooks/docs/oci-hooks.5.md
  #       for more details.
  #       Specific configuration for this interface can be found under "ociHooks".
  #
  #    "nri-hook":
  #      Install the hook via NRI. Use this with containerd. Requires NRI being enabled.
  #      see: https://github.com/containerd/containerd/blob/main/docs/NRI.md.
  #      Specific configuration for this interface can be found under "nriHook".
  #
  interface: ""
  # -- Annotations for the Tetragon rthooks daemonset
  annotations: {}
  # -- Extra labels for the Tetrargon rthooks daemonset
  extraLabels: {}
  # -- Pod annotations for the Tetrargon rthooks pod
  podAnnotations: {}
  # -- priorityClassName for the Tetrargon rthooks pod
  priorityClassName: ""
  # -- security context for the Tetrargon rthooks pod
  podSecurityContext: {}
  # -- installDir is the host location where the tetragon-oci-hook binary will be installed
  installDir: "/opt/tetragon"
  # -- Comma-separated list of namespaces to allow Pod creation for, in case tetragon-oci-hook fails to reach Tetragon agent.
  # The namespace Tetragon is deployed in is always added as an exception and must not be added again.
  failAllowNamespaces: ""
  # -- Extra volume mounts to add to the oci-hook-setup init container
  extraVolumeMounts: []
  # -- resources for the the oci-hook-setup init container
  resources: {}
  # -- extra args to pass to tetragon-oci-hook
  extraHookArgs: {}
  # -- configuration for "oci-hooks" interface
  ociHooks:
    # -- directory to install .json file for running the hook
    hooksPath: "/usr/share/containers/oci/hooks.d"
  # -- configuration for the "nri-hook" interface
  nriHook:
    # -- path to NRI socket
    nriSocket: "/var/run/nri/nri.sock"
  # -- image for the Tetragon rthooks pod
  image:
    override: ~
    repository: quay.io/cilium/tetragon-rthooks
    tag: v0.8
  # -- rthooks service account.
  serviceAccount:
    name: ""

```

### Step 2: Deploy Tetragon

```bash
# Kiểm tra chart
helm lint ./tetragon

# Dry-run - xem resources sẽ tạo
helm install tetragon ./tetragon \
  -n tetragon \
  -f /home/bigboss/vnpost/tetragon/values-airgapped.yaml \
  --dry-run --debug

# Deploy thực tế
helm install tetragon ./tetragon \
  -n tetragon \
  -f /home/bigboss/vnpost/tetragon/values-airgapped.yaml

# Verify
kubectl get pods -n tetragon
kubectl get daemonset -n tetragon
kubectl get deployment -n tetragon
```

### Step 3: Kiểm Tra Deployment

```bash
# Check pod status
kubectl get pods -n tetragon -w

# Check logs
kubectl logs -n tetragon -l app=tetragon --tail=50
kubectl logs -n tetragon -l app=tetragon-operator --tail=50

# Check gRPC server listening
kubectl exec -n tetragon <pod-name> -- netstat -tlnp | grep 54321

# Test event capture (should see events)
kubectl logs -n tetragon -l app=tetragon --tail=100 | grep -i event
```

Cấu hình ```values.yaml``` cần tham khảo: [Tetragon helm chart](https://artifacthub.io/packages/helm/cilium/tetragon)

### Step 4: Verify eBPF Programs Loaded

```bash
# SSH vào node và kiểm tra eBPF programs
kubectl debug node/<node-name> -it --image=ubuntu
mount -t debugfs none /debug
cat /debug/tracing/available_tracers 
cat /sys/kernel/debug/tracing/available_tracers | grep -i kprobe
```

---

## Cấu Hình Ghi Log Đến Loki

### Kiến Trúc Log Flow

Tetragon có hai cách gửi events:

1. **Method 1: Log từ Pod stdout → Log shipper** (Recommended for Kubernetes)
   ```
   Tetragon Agent → stdout → kubelet → container log → Fluent-bit → Loki
   ```

2. **Method 2: Direct gRPC export** (Advanced)
   ```
   Tetragon Agent → gRPC exporter → Loki
   ```

### Option 1: Via Stdout + Fluent-bit (Recommended)

#### 1.1 Tetragon Configuration (của tôi - đã set ở values file)

**tetragon/values-airgapped.yaml** - phần logging:
```yaml
tetragonLogging:
  logLevel: info
  format: json  # ← Critical: JSON format cho Loki parsing
```

**Kết quả:** Tetragon ghi JSON events ra stdout:
```json
{
  "process_kube": {"namespace": "default", "pod": "webserver", "container": "nginx"},
  "process": {"name": "curl", "pid": 1234},
  "parent_process": {"name": "bash", "pid": 567},
  "event_type": "execve",
  "return_code": 0,
  "timestamp": "2026-04-15T10:30:45.123Z"
}
```

#### 1.2 Deploy Fluent-bit để Ship Logs

Nếu chưa có fluent-bit, tạo ConfigMap + DaemonSet:

**File: `/home/bigboss/vnpost/tetragon/fluent-bit-config.yaml`**

```yaml
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluent-bit-tetragon-config
  namespace: tetragon
data:
  fluent-bit.conf: |
    [SERVICE]
        Flush        5
        Daemon       Off
        Log_Level    info
        Parsers_File parsers.conf

    [INPUT]
        Name              tail
        Path              /var/log/containers/*tetragon*.log
        Parser            docker
        Tag               kubernetes.*
        Mem_Buf_Limit     50MB
        Skip_Long_Lines   On
        Refresh_Interval  10

    [FILTER]
        Name                kubernetes
        Match               kubernetes.*
        Kube_URL            https://kubernetes.default.svc:443
        Kube_CA_File        /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
        Kube_Token_File     /var/run/secrets/kubernetes.io/serviceaccount/token
        Use_Kubelet         true
        Labels              on
        Annotations         on

    [OUTPUT]
        Name   loki
        Match  kubernetes.var.log.containers.*tetragon*
        Host   loki.loki.svc.cluster.local
        Port   3100
        Labels job=tetragon, source=kubernetes
        Tenant_ID tetragon_logs
        Drop_Single_Key_Values on

  parsers.conf: |
    [PARSER]
        Name   docker
        Format json
        Time_Key time
        Time_Format %Y-%m-%dT%H:%M:%S.%L%Z

---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: fluent-bit-tetragon
  namespace: tetragon
  labels:
    app: fluent-bit-tetragon
spec:
  selector:
    matchLabels:
      app: fluent-bit-tetragon
  template:
    metadata:
      labels:
        app: fluent-bit-tetragon
    spec:
      serviceAccountName: fluent-bit
      tolerations:
        - effect: NoSchedule
          operator: Exists
        - effect: NoExecute
          operator: Exists
      containers:
      - name: fluent-bit
        image: fluent/fluent-bit:2.1.10
        env:
          - name: LOKI_URL
            value: "http://loki.loki.svc.cluster.local:3100"
          - name: LOKI_TENANT_ID
            value: "tetragon_logs"
        volumeMounts:
          - name: config
            mountPath: /fluent-bit/etc/
          - name: varlog
            mountPath: /var/log
          - name: varlibdockercontainers
            mountPath: /var/lib/docker/containers
            readOnly: true
        resources:
          requests:
            cpu: 50m
            memory: 100Mi
          limits:
            cpu: 200m
            memory: 200Mi
      volumes:
        - name: config
          configMap:
            name: fluent-bit-tetragon-config
        - name: varlog
          hostPath:
            path: /var/log
        - name: varlibdockercontainers
          hostPath:
            path: /var/lib/docker/containers

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: fluent-bit
  namespace: tetragon

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: fluent-bit
rules:
  - apiGroups: [""]
    resources:
      - namespaces
      - pods
      - pods/logs
    verbs: ["get", "list", "watch"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: fluent-bit
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: fluent-bit
subjects:
  - kind: ServiceAccount
    name: fluent-bit
    namespace: tetragon
```

**Deploy:**
```bash
kubectl apply -f /home/bigboss/vnpost/tetragon/fluent-bit-config.yaml

# Verify
kubectl get daemonset -n tetragon
kubectl logs -n tetragon -l app=fluent-bit-tetragon --tail=20
```

#### 1.3 Query Logs từ Loki

```bash
# Port-forward Loki (nếu cần từ local machine)
kubectl port-forward -n loki svc/loki 3100:3100

# Hoặc query từ Grafana UI
# Grafana → Explore → Select Loki data source
# LogQL query: {job="tetragon", container="tetragon"}
#             | json | process_kube_pod != ""
```

**Example LogQL Queries:**

```logql
# All Tetragon events
{job="tetragon"}

# Only events từ pod "webserver"
{job="tetragon"} | json process_kube_pod="webserver"

# Only file writes
{job="tetragon"} | json event_type="vfs_write"

# Only network connections
{job="tetragon"} | json event_type="connect"

# Group by action
{job="tetragon"} | json | stats count() by action

# Errors/warnings
{job="tetragon"} | json severity=~"warn|error"
```

### Option 2: Direct gRPC Export (Advanced)

Nếu muốn Tetragon trực tiếp gửi đến Loki hoặc endpoint gRPC:

**File: `/home/bigboss/vnpost/tetragon/tetragon-exporter.yaml`**

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: tetragon-exporter-config
  namespace: tetragon
data:
  exporter.yaml: |
    # Tetragon gRPC event exporter config
    # Gửi events từ Tetragon agent đến Loki hoặc OTLP collector
    
    exporters:
      - type: grpc
        address: loki.loki.svc.cluster.local:9096  # Loki gRPC port
        insecure: true
        batch:
          timeout: 30s
          send_batch_size: 100
```

**Ghi chú:** Cách này cần Loki support gRPC protocol - kiểm tra Loki version support trước.

---

## TracingPolicy - Định Nghĩa Rules

### Tetragon TracingPolicy Overview

**TracingPolicy** là K8s CRD (Custom Resource Definition) dùng để định nghĩa monitoring rules. Mỗi policy có:

- **Selectors:** Chọn syscall nào cần capture (e.g., `openat`, `execve`, `connect`)
- **Match patterns:** Điều kiện chi tiết (e.g., filename, network port, UID)
- **Actions:** Hành động thực hiện - ALLOW / WARN / DENY
- **Output:** Template định dạng output events

### TracingPolicy Structure

```yaml
apiVersion: cilium.io/v1alpha1
kind: TracingPolicy
metadata:
  name: my-policy
spec:
  events:
    # Danh sách event policies
    - event: "event_name"
      selectors:
        # Syscall selectors
        - matchPolicies: ["selector_1", "selector_2"]
          matchActions: []
          matchReturnCodes:
            - matchOperator: ">"
              value: 0
      actions:
        # Actions khi match
        - action: "ALLOW"  # hay WARN, DENY
        - action: "POST"   # Log event (internal action)
```

### Tài Liệu Tetragon Cho Team

**Link tham khảo chính:**
- [Cilium Tetragon Tracing Policies Guide](https://tetragon.io/docs/concepts/tracing-policy/)
- [Tetragon GitHub Examples](https://github.com/cilium/tetragon/tree/main/examples/policylibrary)
- [Event Definitions](https://tetragon.io/docs/concepts/events/)
- [Selectors & Filters](https://tetragon.io/docs/concepts/tracing-policy/selectors/)

**Tài liệu này sẽ được cập nhật với các link cụ thể.**

---

## Ví Dụ TracingPolicy Thực Tế

Dưới đây là các ví dụ TracingPolicy cho các use-case phổ biến. Lưu tất cả vào:
`/home/bigboss/vnpost/tetragon/policies/`

### Ví Dụ 1: Monitor Tất Cả Process Execution

**File: `/home/bigboss/vnpost/tetragon/policies/monitor-all-execve.yaml`**

```yaml
apiVersion: cilium.io/v1alpha1
kind: TracingPolicy
metadata:
  name: monitor-all-execve
  namespace: tetragon
spec:
  events:
  - event: "execve"
    selectors:
    - matchPolicies: ["always"]
    actions:
    - action: "ALLOW"  # Allow execution, but log it
    output:
      # Output fields
      fields:
      - event_type
      - timestamp
      - process_name
      - process_pid
      - process_uid
      - process_gid
      - parent_process_name
      - parent_process_pid
      - process_kube_namespace
      - process_kube_pod
      - process_kube_container
      - args
      - return_code
```

**Mục đích:** Ghi lại tất cả process executions (ai chạy lệnh gì)

**Events captured:**
```json
{
  "event_type": "execve",
  "timestamp": "2026-04-15T10:30:45.123Z",
  "process": {"name": "bash", "pid": 5678, "uid": 0},
  "parent_process": {"name": "sshd", "pid": 1234},
  "process_kube": {"namespace": "default", "pod": "worker-1", "container": "app"},
  "args": ["bash", "-c", "echo hello"]
}
```

### Ví Dụ 2: Monitor Suspicious File Writes

**File: `/home/bigboss/vnpost/tetragon/policies/monitor-sensitive-files.yaml`**

```yaml
apiVersion: cilium.io/v1alpha1
kind: TracingPolicy
metadata:
  name: monitor-sensitive-files
  namespace: tetragon
spec:
  events:
  - event: "vfs_write"
    selectors:
    # Match: viết vào các file nhạy cảm
    - matchPolicies: ["io_write"]
      matchArgsOperator: "and"
      matchArgs:
      # File patterns to monitor
      - index: 0
        operator: "Equal"
        values:
        - "/etc/passwd"
        - "/etc/shadow"
        - "/etc/sudoers"
        - "/root/.ssh/authorized_keys"
    actions:
    - action: "WARN"  # Warn on write - allow but log
    output:
      fields:
      - event_type
      - timestamp
      - process_name
      - process_pid
      - process_uid
      - file_path
      - bytes_written
      - return_code
      - process_kube_namespace
      - process_kube_pod
```

**Mục đích:** Phát hiện unauthorized writes vào sensitive files

**ĐÃ trigger khi:**
- Bất kỳ process nào viết vào `/etc/passwd`, `/etc/shadow`, etc.
- Ghi lại process details + uid + bytes written

### Ví Dụ 3: Monitor Outbound Network Connections

**File: `/home/bigboss/vnpost/tetragon/policies/monitor-network-connections.yaml`**

```yaml
apiVersion: cilium.io/v1alpha1
kind: TracingPolicy
metadata:
  name: monitor-network-connections
  namespace: tetragon
spec:
  events:
  - event: "connect"
    selectors:
    # Match: outbound connections NOT to local services
    - matchPolicies: ["outbound_conn"]
      matchArgs:
      - index: 0  # Socket protocol
        operator: "Equal"
        values:
        - "IPPROTO_TCP"
        - "IPPROTO_UDP"
      - index: 1  # Destination IP (exception: local ranges)
        operator: "NotEqual"
        values:
        - "127.0.0.1"
        - "10.0.0.0/8"
        - "172.16.0.0/12"
        - "192.168.0.0/16"
    actions:
    - action: "ALLOW"  # Allow but log
    output:
      fields:
      - event_type
      - timestamp
      - process_name
      - process_pid
      - process_uid
      - socket_family
      - destination_address
      - destination_port
      - protocol
      - return_code
      - process_kube_namespace
      - process_kube_pod
```

**Mục đích:** Detect side-channel data exfiltration attempts

**Events captured:**
```json
{
  "event_type": "connect",
  "timestamp": "2026-04-15T10:35:12.456Z",
  "process": {"name": "curl", "pid": 9876, "uid": 1000},
  "socket": {"family": "AF_INET", "protocol": "IPPROTO_TCP"},
  "destination": {"address": "203.0.113.42", "port": 443}
}
```

### Ví Dụ 4: Block Privilege Escalation

**File: `/home/bigboss/vnpost/tetragon/policies/block-privilege-escalation.yaml`**

```yaml
apiVersion: cilium.io/v1alpha1
kind: TracingPolicy
metadata:
  name: block-privilege-escalation
  namespace: tetragon
spec:
  events:
  - event: "execve"
    selectors:
    # Detect setuid/setgid/setcap execution
    - matchPolicies: ["setuid_detection"]
      matchArgs:
      - index: 0  # Binary path
        operator: "Prefix"
        values:
        - "/bin/sudo"
        - "/usr/bin/sudo"
        - "/bin/su"
        - "/usr/bin/su"
        - "/sbin/setcap"
        - "/usr/sbin/setcap"
      - index: 1  # Current UID must be non-root
        operator: "NotEqual"
        values:
        - "0"  # root UID
    actions:
    - action: "DENY"  # Deny execution
    - action: "POST"  # Log it
    output:
      fields:
      - event_type
      - timestamp
      - process_name
      - process_pid
      - process_uid
      - process_gid
      - args
      - return_code
      - process_kube_namespace
      - process_kube_pod
```

**Mục đích:** Block non-root users từ chạy setuid programs (unless allowed)

**Kết quả:** Execution sẽ bị deny + ghi log

### Ví Dụ 5: Monitor Shell Access vào Container

**File: `/home/bigboss/vnpost/tetragon/policies/monitor-shell-access.yaml`**

```yaml
apiVersion: cilium.io/v1alpha1
kind: TracingPolicy
metadata:
  name: monitor-shell-access
  namespace: tetragon
spec:
  events:
  - event: "execve"
    selectors:
    # Detect interactive shells (bash, sh, zsh, etc.)
    - matchPolicies: ["interactive_shell"]
      matchArgs:
      - index: 0  # Program name
        operator: "Equal"
        values:
        - "/bin/bash"
        - "/bin/sh"
        - "/bin/zsh"
        - "/bin/csh"
        - "/bin/ksh"
      - index: 1  # Check if parent is ssh, kubectl exec, etc.
        operator: "Equal"
        values:
        - "sshd"
        - "kubectl"
        - "docker"  # if user exec'd into container
    actions:
    - action: "WARN"  # Warn on interactive shell
    output:
      fields:
      - event_type
      - timestamp
      - process_name
      - process_pid
      - process_uid
      - parent_process_name
      - parent_process_pid
      - args
      - process_kube_namespace
      - process_kube_pod
      - process_kube_container
```

**Mục đích:** Audit interactive shell access vào containers

**Use-case:** Compliance - track who/when accessed containers

---

## Deploy TracingPolicies

### Step 1: Tạo Policies Directory

```bash
mkdir -p /home/bigboss/vnpost/tetragon/policies
# Copy các yaml files ở trên vào directory này
```

### Step 2: Deploy Policies

```bash
# Apply all policies
kubectl apply -f /home/bigboss/vnpost/tetragon/policies/

# Verify
kubectl get tracingpolicies -n tetragon
kubectl describe tracingpolicy monitor-all-execve -n tetragon
```

### Step 3: Monitor Events

```bash
# Watch events in real-time
kubectl logs -n tetragon -f -l app=tetragon --tail=0

# Filter cho specific events
kubectl logs -n tetragon -l app=tetragon | \
  grep -i "connect\|denied\|warn"
```

---

## Vận Hành & Theo Dõi

### Health Checks

```bash
# 1. Check pod status
kubectl get pods -n tetragon
# Expected: All pods Running (1/1)

# 2. Check operator is healthy
kubectl exec -n tetragon $(kubectl get pod -n tetragon -l app=tetragon-operator -o name) -- \
  curl -s http://localhost:8080/healthz

# 3. Check gRPC server responding
kubectl exec -n tetragon $(kubectl get pod -n tetragon -l app=tetragon -o name | head -1) -- \
  ps aux | grep tetragon | grep -v grep

# 4. Check Loki receiving logs
curl http://loki.loki.svc.cluster.local:3100/ready
# Response: ready
```

### Viewing Events

#### Via kubectl logs:
```bash
# Tail recent events
kubectl logs -n tetragon -l app=tetragon -f

# Search for specific patterns
kubectl logs -n tetragon -l app=tetragon | grep "action.*DENY"
```

#### Via Grafana Loki:
```
Grafana → Explore tab → Select Loki datasource
Query: {job="tetragon"} | json
```

### Metrics & Monitoring

Tetragon exposes Prometheus metrics:

```bash
# Port-forward operator metrics
kubectl port-forward -n tetragon svc/tetragon-operator 2112:2112

# Curl metrics endpoint
curl http://localhost:2112/metrics

# Useful metrics:
# - tetragon_events_total - total events processed
# - tetragon_events_cached - cached events
# - tetragon_perf_ring_buffer_lost - dropped events
# - tetragon_policies_loaded - active policies count
```

### Update Policies

```bash
# Edit TracingPolicy
kubectl edit tracingpolicy monitor-all-execve -n tetragon

# Or update from file
kubectl apply -f /home/bigboss/vnpost/tetragon/policies/monitor-all-execve.yaml

# Changes apply immediately (no restart needed)

# Verify update
kubectl get tracingpolicy monitor-all-execve -o yaml
```

### Upgrade Tetragon

```bash
# Check current version
kubectl get deployment -n tetragon tetragon-operator -o jsonpath='{.spec.template.spec.containers[0].image}'

# Upgrade
helm upgrade tetragon ./tetragon \
  -n tetragon \
  -f /home/bigboss/vnpost/tetragon/values-airgapped.yaml \
  --set tetragon.image.tag=v1.7.0  # newer version
```

---

## Khắc Phục Sự Cố

### Issue 1: Tetragon pods not starting

**Symptoms:**
```
kubectl get pods -n tetragon
# STATUS: CrashLoopBackOff
```

**Troubleshooting:**
```bash
# Check logs
kubectl logs -n tetragon -l app=tetragon --previous

# Common issues:
# 1. eBPF not supported on node
# 2. Privileged pod not allowed (PSP issue)
# 3. Image not found (registry issue)

# Check node kernel
kubectl debug node/<node-name> -it --image=ubuntu
uname -r  # Must be >= 5.4
```

**Fix:** Ensure nodes have eBPF support + privileged pods allowed

### Issue 2: No events appearing

**Symptoms:**
```
kubectl logs -n tetragon -l app=tetragon
# Output: [info] started, waiting for events...
# But no events showing up
```

**Troubleshooting:**
```bash
# 1. Check if TracingPolicies are loaded
kubectl get tracingpolicies -n tetragon
# If empty, policies not applied

# 2. Check policy syntax
kubectl explain tracingpolicies.spec

# 3. Check Tetragon operator logs
kubectl logs -n tetragon -l app=tetragon-operator

# 4. Manually trigger event
kubectl exec -n default $(kubectl get pod -n default -o name | head -1) -- \
  bash -c "echo test"

# Should see execve event in Tetragon logs
kubectl logs -n tetragon -l app=tetragon | tail -20
```

**Fix:** Verify TracingPolicies syntax and that operator loaded them

### Issue 3: Loki not receiving logs

**Symptoms:**
```
# Tetragon logs showing events, but Loki is empty
kubectl logs -n tetragon -l app=tetragon | head -5
# Shows many events

# But in Grafana/Loki: no results for {job="tetragon"}
```

**Troubleshooting:**
```bash
# 1. Check fluent-bit pod
kubectl get pods -n tetragon -l app=fluent-bit-tetragon
kubectl logs -n tetragon -l app=fluent-bit-tetragon

# 2. Check Loki connectivity
kubectl exec -n tetragon $(kubectl get pod -n tetragon -l app=fluent-bit-tetragon -o name | head -1) -- \
  curl -v http://loki.loki.svc.cluster.local:3100/loki/api/v1/labels

# 3. Check fluent-bit config
kubectl get configmap -n tetragon fluent-bit-tetragon-config -o yaml

# 4. Check Loki is running
kubectl get pods -n loki
kubectl logs -n loki -l app=loki
```

**Fix:** 
- Verify Loki DNS name, port reachable
- Check fluent-bit config is correct (Loki endpoint, labels)
- Check Loki itself has storage configured

### Issue 4: High memory/CPU usage

**Symptoms:**
```
kubectl top pods -n tetragon
# tetragon agent pods using 800Mi RAM, 400m CPU (too high)
```

**Troubleshooting:**
```bash
# 1. Check how many policies are loaded
kubectl get tracingpolicies -n tetragon | wc -l

# 2. Check policy complexity
kubectl get tracingpolicy -o json | jq '.items[].spec.events | length'

# 3. Check number of events captured
kubectl logs -n tetragon -l app=tetragon | wc -l  # Over 1min

# Optimization:
# - Remove unused policies: kubectl delete tracingpolicy <name>
# - Filter policies (more specific selectors)
# - Reduce verbosity: set log level to warn
```

**Fix:** Simplify policies, remove unnecessary selectors, reduce number of tracked syscalls

### Issue 5: Events being dropped

**Symptoms:**
```
Metrics show: tetragon_perf_ring_buffer_lost_high
# Means: events dropped due to buffer overflow
```

**Troubleshooting & Fix:**
```bash
# 1. Increase ring buffer size (in values.yaml if available)
# 2. Reduce event volume (fewer policies, more specific selectors)
# 3. Scale up Tetragon resources:
kubectl set resources daemonset tetragon -n tetragon \
  --limits=cpu=1000m,memory=1Gi

# 4. Increase perf buffer sizes (kernel tuning):
sysctl -w kernel.perf_event_paranoid=-1
```

---

## Tài Liệu Tham Khảo

### Official Documentation

| Topic | Link | Status |
|-------|------|--------|
| **Tetragon Getting Started** | [docs.cilium.io/tetragon](https://docs.cilium.io/en/stable/gettingstarted/tetragon/) | Primary |
| **Tracing Policies Guide** | [docs.cilium.io/tracing-policies](https://docs.cilium.io/en/stable/gettingstarted/tracing-policies/) | Primary |
| **Event Types Reference** | [docs.cilium.io/events](https://docs.cilium.io/en/stable/reference/tetragon/events/) | Primary |
| **Selectors & Filters** | [docs.cilium.io/selectors](https://docs.cilium.io/en/stable/reference/tetragon/selectors/) | Primary |
| **Tetragon CLI** | [docs.cilium.io/cli](https://docs.cilium.io/en/stable/reference/tetragon/cli/) | Reference |
| **GitHub - Tetragon Repo** | [github.com/cilium/tetragon](https://github.com/cilium/tetragon) | Code Examples |
| **GitHub - Policy Examples** | [github.com/cilium/tetragon/examples/policies](https://github.com/cilium/tetragon/tree/main/examples/policies) | Code Examples |

### Common Event Types

> **Tham khảo:** [docs.cilium.io/events](https://docs.cilium.io/en/stable/reference/tetragon/events/)

| Event | Description | Use Case |
|-------|-------------|----------|
| `execve` | Process execution | Detect unauthorized programs, track execution chains |
| `vfs_open` | File read | Track file access, detect sensitive file reads |
| `vfs_write` | File write | Detect unauthorized file modifications |
| `connect` | Network connection | Track outbound connections, detect C&C activity |
| `socket_connect_ipv4` | IPv4 connection | Similar to connect, IPv4 specific |
| `socket_connect_ipv6` | IPv6 connection | Similar to connect, IPv6 specific |
| `clone` | Process creation | Detect process forking, thread creation |
| `mmap` | Memory mapping | Track dynamic code loading, exploit attempts |
| `sched_process_exit` | Process termination | Cleanup tracking, process lifetime monitoring |

### Common Selectors

> **Tham khảo:** [docs.cilium.io/selectors](https://docs.cilium.io/en/stable/reference/tetragon/selectors/)

| Selector | Matches | Example |
|----------|---------|---------|
| `Equal` | Exact match | `operator: "Equal", values: ["/bin/bash"]` |
| `NotEqual` | Not matching | `operator: "NotEqual", values: ["0"]` |
| `Prefix` | String prefix | `operator: "Prefix", values: ["/etc"]` |
| `Suffix` | String suffix | `operator: "Suffix", values: [".so"]` |
| `Regex` | Regex pattern | `operator: "Regex", values: ["^/tmp/.*\\.sh$"]` |
| `GreaterThan` | Numeric greater | `operator: ">", value: "1024"` |
| `LessThan` | Numeric less | `operator: "<", value: "1024"` |

### Team Learning Path

1. **Week 1: Fundamentals**
   - Read: [Tetragon Getting Started](https://docs.cilium.io/en/stable/gettingstarted/tetragon/)
   - Task: Set up Tetragon in your cluster (follow this doc)
   - Explore: Write your first SimplerTracingPolicy

2. **Week 2: Policy Writing**
   - Study: [Tracing Policies Guide](https://docs.cilium.io/en/stable/gettingstarted/tracing-policies/)
   - Study: [Event Types Reference](https://docs.cilium.io/en/stable/reference/tetragon/events/)
   - Task: Create 3-5 custom policies for your apps

3. **Week 3: Advanced**
   - Study: [Event Filters & Selectors](https://docs.cilium.io/en/stable/reference/tetragon/selectors/)
   - Review: [Community Policies](https://github.com/cilium/tetragon/tree/main/examples/policies)
   - Task: Write complex policies with multiple conditions

4. **Ongoing: Operational Excellence**
   - Monitor metrics regularly
   - Adjust policies based on false positives
   - Share learnings with team

---

## Ghi Chú Khác

### Airgapped Environment Best Practices

1. **Image Management:**
   - Maintain local copy of images trong registry
   - Test image upgrades trước apply lên production

2. **Policy Management:**
   - Version control tất cả TracingPolicy files
   - Use GitOps (ArgoCD/Flux) để deploy policies

3. **Log Retention:**
   - Configure Loki retention policy (default: 30 days)
   - Backup important logs to external storage

4. **Performance Tuning:**
   - Start with basic policies, add incrementally
   - Monitor resource usage closely
   - Tune based on actual event volume

### Security Best Practices

1. **RBAC:**
   - Restrict who can create/modify TracingPolicies
   - Audit all policy changes

2. **Sensitive Data:**
   - Be careful: Tetragon logs process arguments, env vars
   - May contain secrets - configure output filtering

3. **Enforcement vs. Observability:**
   - Start with WARN/ALLOW actions
   - Move to DENY once rules are validated

---

## Liên Hệ & Support

- **Tetragon Slack:** [Cilium Slack - #tetragon](https://cilium.slack.com)
- **GitHub Issues:** [github.com/cilium/tetragon/issues](https://github.com/cilium/tetragon/issues)
- **Docs:** [docs.cilium.io](https://docs.cilium.io)

---
