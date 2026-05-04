# Tài Liệu Kỹ Thuật Pyroscope

Tác giả: Hồ Công Thành

## Mục Lục
1. [Giới Thiệu Pyroscope](#giới-thiệu-pyroscope)
2. [Kiến Trúc Hệ Thống](#kiến-trúc-hệ-thống)
3. [Yêu Cầu Hệ Thống](#yêu-cầu-hệ-thống)
4. [Chuẩn Bị Trước Triển Khai](#chuẩn-bị-trước-triển-khai)
5. [Triển Khai Kubernetes + Air Gap](#triển-khai-kubernetes--air-gap)
6. [Triển Khai Pyroscope với Helm](#triển-khai-pyroscope-với-helm)

8. [Cấu Hình Lưu Trữ S3](#cấu-hình-lưu-trữ-s3)
9. [Tích Hợp Grafana](#tích-hợp-grafana)
10. [API Tài Liệu](#api-tài-liệu)
11. [Vận Hành và Theo Dõi](#vận-hành-và-theo-dõi)
12. [Khắc Phục Sự Cố](#khắc-phục-sự-cố)

---

## Giới Thiệu Pyroscope

Pyroscope là một công cụ **continuous profiling** được phát triển bởi Grafana Labs. Nó giúp Devlopers và DevOps/SREs:

- **Phát hiện bottleneck về hiệu suất** - Xác định chính xác đoạn code tiêu tốn CPU/bộ nhớ
- **Giảm tiêu thụ tài nguyên** - Tối ưu hóa ứng dụng để tiết kiệm CPU, Memory và chi phí infrastructure
- **Cải thiện trải nghiệm người dùng** - Giảm latency và response time
- **Theo dõi hiệu suất thời gian thực** - Observable ngay lập tức mà không cần restart ứng dụng



---

## Kiến Trúc Hệ Thống

### Tổng Quan Kiến Trúc

```
┌─────────────────────────────────────────────────────────┐
│              Ứng Dụng (Application)                     │
│         (Python, Go, Node.js, Java, v.v.)               │
└────────────────────┬────────────────────────────────────┘
                     │
        ┌────────────▼────────────────────────┐
        │  Alloy eBPF Profiler                │
        │  (SDK/Library)                      │
        └────────────┬────────────────────────┘
                     │
        ┌────────────▼────────────────────────┐
        │   Pyroscope Server (Kubernetes Pod) │
        │ ┌──────────────────────────────────┐│
        │ │ Storage Layer (S3 Backend)       ││
        │ │ - Time Series Data               ││
        │ │ - Profiles Nén Cao               ││
        │ │ - MinIO hoặc S3-compatible       ││
        │ └──────────────────────────────────┘│
        │ ┌──────────────────────────────────┐│
        │ │ Query Engine                     ││
        │ │ - Merge Profiles                 ││
        │ │ - Generate Flame Graphs          ││
        │ └──────────────────────────────────┘│
        └────────────┬────────────────────────┘
                     │
        ┌────────────▼────────────────────────┐
        │  Prometheus + Grafana Dashboard     │
        │ - Visualization                     │
        │ - Performance Monitoring            │
        │ - Alerting Rules                    │
        └─────────────────────────────────────┘
```

### Các Thành Phần Chính

#### 1. **Alloy eBPF Profiler**
- Chạy trong ứng dụng của bạn
- **Tự động** thu thập dữ liệu từ stack traces (không cần code changes)
- Gửi dữ liệu đến Pyroscope Server
- Overhead cực thấp (< 0.5% CPU overhead)

#### 2. **Pyroscope Server (Chạy trên Kubernetes)**
- Nhận dữ liệu từ agents via gRPC hoặc HTTP
- Nén và lưu trữ profiles vào S3 backend
- Cung cấp API Query để lấy dữ liệu profiling
- Tạo flame graphs cho visualization tương tác
- Resource tối thiểu: 200m CPU, 512Mi Memory

#### 3. **S3 Storage Backend**
- Lưu trữ time-series data hiệu quả
- Nén cao: 100:1 compression ratio
- Hỗ trợ retention policies (mặc định 30 ngày)
- Có thể dùng MinIO (in-cluster) hoặc S3-compatible storage bên ngoài

#### 4. **Prometheus + Grafana**
- **Prometheus**: Thu thập metrics từ Pyroscope (port 4041)
- **Grafana**: Visualize flame graphs, performance dashboards
- **Alerting**: Rule-based alerts cho CPU spike, memory leak, S3 issues

---

## Yêu Cầu Hệ Thống

### Kubernetes Cluster (Yêu Cầu)

```
- Kubernetes 1.20+
- CoreDNS hoặc CNI phù hợp 
- Storage class (cho PersistentVolume)
- Namespace riêng: profiling
- Air-gapped environment
```

### S3 Storage (Yêu Cầu)

```
- S3-compatible storage (AWS S3, Ceph, OpenStack Swift, RustFS...)
- Bucket: pyroscope-profiles (tùy chỉnh)
- Dung lượng tối thiểu: 100GB (phụ thuộc retention)
- Credentials: Access Key + Secret Key
- NOTE: MinIO đã hết open source 
```

### Monitoring Stack

```
- Prometheus 2.30+
- Grafana 10.0+
```

### Hỗ Trợ Ngôn Ngữ (Cho Instrumentation)

```
- Python 3.7+
- Go 1.18+
- Node.js 14+
- Java 8+
- Ruby 2.7+
- PHP 7.4+
- Rust
- .NET
```

---

## Chuẩn Bị Trước Triển Khai

### 1. Danh Sách Kiểm Tra Pre-Deployment

- [ ] Xác định yêu cầu retention (lưu bao lâu?)
- [ ] Tính toán dung lượng S3 dựa trên traffic
  - Công thức: `(ứng dụng) × (traffic RPS) × 100 samples/s ÷ 1024 = MB/ngày`
  - Ví dụ: 10 app × 1000 RPS × 100 samples/s ÷ 1024 ≈ 976 MB/ngày
  - Với 30 ngày retention: 30 GB storage
- [ ] Chuẩn bị Kubernetes cluster (air-gap capable)
- [ ] Chuẩn bị MinIO hoặc S3-compatible storage
- [ ] Chuẩn bị Prometheus + Grafana
- [ ] Cấu hình network policies (nếu cần)
- [ ] Xác định resource limits cho Pods
- [ ] Chuẩn bị private container registry (nếu air-gap)
- [ ] Đảm bảo NTP đồng bộ (quan trọng cho time-series data)
- [ ] Chuẩn bị tài liệu & training cho team

### 2. Ước Tính Tài Nguyên

**CPU Estimation:**
```
Base:           100m
Per 1K RPS:     50m
Per app:        20m

Ví dụ: 10 ứng dụng @ 100 RPS mỗi cái
= 100m + (10 × 20m) = 300m CPU
= 1000m = 1 CPU (khuyến nghị)
```

**Memory Estimation:**
```
Base:           512Mi
Per 100K samples/s: 100Mi
Max recommended: 2Gi

Ví dụ: 100K samples/s
= 512Mi + 100Mi = 612Mi
= 1Gi (khuyến nghị)
```

**Storage Estimation:**
```
Per app per day:           ~800MB
Với 7 ngày retention:      5.6GB
Với 30 ngày retention:     24GB
Khuyến nghị S3 bucket:     100GB+ (buffer)
```

### 3. Bảo Mật Trước Triển Khai

- [ ] Tạo TLS certificates cho tất cả services
- [ ] Cấu hình RBAC trong Kubernetes
- [ ] Tạo dedicated service account cho Pyroscope
- [ ] Cấu hình network policies (restrict traffic)
- [ ] Chuẩn bị credential management (Secrets)
- [ ] Chuẩn bị backup encryption
- [ ] Cấu hình audit logging

---

## Triển Khai Kubernetes + Air Gap

### Bước 1: Chuẩn Bị Namespace

```bash
kubectl create namespace monitoring
kubectl label namespace monitoring name=monitoring
```

### Bước 2: Tạo S3 Credentials Secret

```bash
# Thay thế với S3 credentials thực tế
kubectl create secret generic s3-credentials \
  --from-literal=AWS_ACCESS_KEY_ID=minioadmin \
  --from-literal=AWS_SECRET_ACCESS_KEY=minioadmin \
  -n profiling
```

### Bước 3: Tạo ConfigMap cho Pyroscope(viết trong Helm values)
Tham khảo Helm values cho pyroscope do tác giả viết 

---

## Triển Khai Pyroscope với Helm

### Helm Chart Repositories

```bash
# Thêm Grafana Helm repository
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update

# Tìm available versions
helm search repo grafana/pyroscope --versions
```

### Cấu Hình Helm Values

Tạo file `values-daemonset.yaml` cho Pyroscope(default, sẽ được cập nhật sau)

### Pyroscope values để tham khảo

[Pyroscope repository](https://github.com/grafana/pyroscope/tree/main/operations/pyroscope/helm/pyroscope)

### Triển Khai DaemonSet Mode (Agent-based)

```bash
# 1. Tạo namespace
kubectl create namespace profiling

# 2. Cài đặt Pyroscope DaemonSet với Helm
helm install pyroscope-agent grafana/pyroscope \
  -n profiling \
  -f values-daemonset.yaml

# 3. Kiểm tra DaemonSet status
kubectl get daemonset -n profiling
kubectl get pods -n profiling

# 4. Xem logs từ một pod
kubectl logs -n profiling -l app.kubernetes.io/name=pyroscope-agent -f

# 5. Kiểm tra số nodes running agents
kubectl get pods -n profiling -o wide

# 6. Verify eBPF profiling
kubectl exec -n profiling <pod-name> -- ps aux | grep pyroscope
```

### Upgrade Helm Release

```bash
# Cập nhật Helm repo
helm repo update

# Upgrade to new version
helm upgrade pyroscope grafana/pyroscope \
  -n profiling \
  -f values-deployment.yaml \
  --version v0.38.0

# Kiểm tra upgrade status
kubectl rollout status deployment/pyroscope -n profiling

# Rollback nếu cần
helm rollback pyroscope 1 -n profiling
```

### Uninstall Pyroscope

```bash
# Xóa Helm release
helm uninstall pyroscope -n profiling

# Xóa namespace nếu cần
kubectl delete namespace profiling
```

### Helm Chart Customization for Air Gap

Để deploy trong Air Gap environment:

```bash
# 1. Download Helm chart locally (trước khi vào air-gap)
helm pull grafana/pyroscope --version v0.37.0

# 2. Transfer chart to air-gap cluster
scp pyroscope-0.37.0.tgz user@air-gap-host:/tmp/

# 3. Install từ local chart
helm install pyroscope /tmp/pyroscope-0.37.0.tgz \
  -n profiling \
  -f values-deployment.yaml \
  --set image.repository=private-registry/grafana/pyroscope \
  --set image.tag=v0.37.0

# Hoặc extract chart để custom:
tar xzf pyroscope-0.37.0.tgz
# Chỉnh sửa values.yaml trong pyroscope/ directory
helm install pyroscope ./pyroscope \
  -n profiling \
  -f custom-values.yaml
```

---

## Tích Hợp Alloy eBPF Profiler

### Tổng Quan

Alloy eBPF Profiler cho phép thu thập profiling data **hoàn toàn tự động** ở kernel level mà **không cần bất kỳ code modification** và **không cần application restart**. Đặc điểm:

1. **Zero-code method** - Không cần instrumentation code trong ứng dụng
2. **Kernel-level profiling** - Thu thập dữ liệu trực tiếp từ kernel via eBPF
3. **Universal** - Hoạt động với tất cả ngôn ngữ (Python, Go, Java, Node.js, Rust, C++, etc.)
4. **Low overhead** - Chi phí CPU < 0.1% thường là rất nhỏ
5. **Non-invasive** - Không cần restart ứng dụng, không cần code changes

**Kiến trúc hoạt động:**
- Alloy eBPF Profiler (DaemonSet) chạy trên mỗi node
- Collect CPU, Memory, I/O profiles via eBPF programs
- Forward data trực tiếp tới Pyroscope Server
- Pyroscope lưu trữ vào S3 backend

### Alloy eBPF Profiler - Helm Chart Values

Tạo file `alloy-values.yaml` cho Alloy eBPF Profiler (kernel-level profiling) (Đã được viết làm reference)

### Triển Khai Alloy eBPF Profiler + Pyroscope với Helm

```bash
# 1. Thêm Alloy Helm repository
helm repo add alloy https://grafana.com/helm-charts
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update

# 2. Tạo namespace
kubectl create namespace profiling

# 3. Tạo S3 credentials Secret
kubectl create secret generic s3-credentials \
  --from-literal=AWS_ACCESS_KEY_ID=minioadmin \
  --from-literal=AWS_SECRET_ACCESS_KEY=minioadmin \
  -n profiling

# 4. Cài đặt Pyroscope Server
helm install pyroscope grafana/pyroscope \
  -n profiling \
  -f values-deployment.yaml

# 5. Cài đặt Alloy eBPF Profiler (DaemonSet - trên tất cả nodes)
helm install alloy-profiler grafana/alloy \
  -n profiling \
  -f values-ebpf.yaml

# 6. Kiểm tra deployment
kubectl get daemonset -n profiling
kubectl get pods -n profiling -o wide

# 7. Xem logs để verify
kubectl logs -n profiling -l app=ebpf-profiler -f --tail=100

# 8. Verify profiling data được gửi tới Pyroscope
curl http://localhost:4040/api/v1/service-instances  # Port-forward nếu cần
```

### Kiến Trúc eBPF Profiler + Pyroscope

Khi sử dụng Alloy eBPF Profiler, kiến trúc hoạt động như sau:

```
┌─────────────────────────────────────────────────────┐
│    Applications (Python, Go, Java, Node.js, etc.)   │
│          (KHÔNG CẦN CODE CHANGES)                   │
│          (KHÔNG CẦN RESTART)                        │
└────────────────────┬────────────────────────────────┘
                     │
      ┌──────────────▼──────────────┐
      │  Kernel (eBPF Programs)     │
      │  - CPU profiling            │
      │  - Memory tracking          │
      │  - I/O tracing              │
      │  - Stack unwinding          │
      └──────────────┬──────────────┘
                     │
      ┌──────────────▼──────────────┐
      │ Alloy eBPF Profiler | 
      │ (DaemonSet trên mỗi node)   │
      │ - Collect eBPF data         │
      │ - Process & batch           │
      │ - Export via OTLP gRPC      │
      └──────────────┬──────────────┘
                     │ (OTLP gRPC:4317)
      ┌──────────────▼───────────────┐
      │  Pyroscope Server            │
      │  - Nhận profiling data       │
      │  - Lưu vào S3 backend        │
      │  - Visualization UI          │
      └──────────────┬───────────────┘
                     │
      ┌──────────────▼───────────────┐
      │   S3 Storage (MinIO)         │
      │   - Long-term storage        │
      │   - Retention: 30 ngày       │
      └──────────────────────────────┘
```




# OTEL_EXPORTER_OTLP_ENDPOINT đúng:
``` OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector.profiling.svc.cluster.local:4317```

# KHÔNG dùng:
``` OTEL_EXPORTER_OTLP_ENDPOINT=http://pyroscope.profiling.svc.cluster.local:4040```

### Ứng Dụng - Không Cần Thay Đổi Code

Với Alloy eBPF Profiler, **không cần bất kỳ thay đổi code** cho ứng dụng của bạn. Tất cả ứng dụng hoạt động như bình thường

**eBPF Profiler workflow:**

1. **Deploy ứng dụng** - Chạy như bình thường, không cần code changes
2. **eBPF Profiler runs** - Chạy như DaemonSet trên mỗi node
3. **Kernel profiling** - Tự động capture profiles từ kernel
4. **Data forwarding** - Gửi profiles tới Pyroscope via OTLP
5. **Visualization** - Xem profiles trong Grafana/Pyroscope UI

**Ngôn ngữ được hỗ trợ (tất cả):**

✅ Python, Go, Java, Node.js, Rust, C++, C, Ruby, PHP, .NET, Kotlin, Scala, Erlang, v.v.

**Điều kiện yêu cầu duy nhất:**

1. Kubernetes cluster với kernel >= 4.15 (hầu hết distros đều có)
2. eBPF Profiler DaemonSet đang chạy
3. Pyroscope Server đang chạy

**Kiểm tra xem profiling data có được collect:**

```bash
# 1. Verify eBPF Profiler pods đang chạy
kubectl get pods -n profiling -l app=ebpf-profiler

# 2. Xem logs từ profiler
kubectl logs -n profiling -l app=ebpf-profiler --tail=50

# 3. Check xem profiles đang được gửi tới Pyroscope
curl -s http://pyroscope.profiling.svc.cluster.local:4040/api/v1/service-instances | jq

# 4. View flame graphs
# Port-forward hoặc access via Grafana datasource
kubectl port-forward -n profiling svc/pyroscope 4040:4040
# Truy cập: http://localhost:4040
```

### Chi Tiết Kỹ Thuật eBPF Profiler

#### eBPF Programs Thu Thập

Profiler sử dụng eBPF programs để thu thập data từ kernel:

```
CPU Profiling (perf):
  - PMU (Performance Monitoring Unit) events
  - Schedule events
  - Context switches

Memory Profiling:
  - Kernel memory allocations via kmem:mm_page_alloc
  - User memory via malloc/mmap hooks
  - Garbage collection events (nếu available)

Stack Unwinding:
  - Kernel stack traces (built-in)
  - User space stack traces (dwarf/ORC)
  - Optimization frames skipping

Symbol Resolution:
  - /proc/[pid]/maps - binary mappings
  - ELF headers + symbol tables
  - DWARF debug info (nếu available)
```

## Tích Hợp Grafana

### Bước 1: Thêm Pyroscope Datasource

Trong Grafana UI:
1. Go to **Configuration > Data Sources**
2. Click **Add data source**
3. Chọn type **Pyroscope**
4. Cấu hình:
   - URL: `http://pyroscope.profiling.svc.cluster.local:4040`
   - Name: `Pyroscope`
   - Click **Save & Test**

**Hoặc qua provisioning:**

```yaml
# grafana-datasource.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-datasources
  namespace: monitoring
data:
  pyroscope.yaml: |
    apiVersion: 1
    datasources:
    - name: Pyroscope
      type: pyroscope
      access: proxy
      url: http://pyroscope.profiling.svc.cluster.local:4040
      isDefault: false
```

### Bước 2: Tạo Grafana Dashboards

#### Dashboard: CPU Profiling

```json
{
  "dashboard": {
    "title": "Pyroscope CPU Profiling",
    "panels": [
      {
        "title": "CPU Flame Graph - Python API",
        "type": "pyroscope",
        "targets": [
          {
            "expr": "select_docker_stats",
            "refId": "A",
            "service": "python-api"
          }
        ]
      }
    ]
  }
}
```

#### Dashboard: Memory Profiling

```json
{
  "dashboard": {
    "title": "Pyroscope Memory Profiling",
    "panels": [
      {
        "title": "Memory Allocation - Go API",
        "type": "pyroscope",
        "targets": [
          {
            "expr": "memory",
            "refId": "A",
            "service": "go-api"
          }
        ]
      }
    ]
  }
}
```

### Bước 3: Tích Hợp Prometheus Metrics

```yaml
# prometheus-scrape-config
- job_name: 'pyroscope'
  static_configs:
    - targets: ['pyroscope.profiling.svc.cluster.local:4041']
  scrape_interval: 30s
  scrape_timeout: 10s
```

### Bước 4: Alerting Rules

```yaml
# prometheus-rules.yaml
groups:
- name: pyroscope
  rules:
  - alert: PyroscopeCPUSpikeDetected
    expr: rate(pyroscope_ingestion_profiles_total[5m]) > 1000
    for: 5m
    annotations:
      summary: "CPU spike detected in {{ $labels.service }}"
  
  - alert: PyroscopeMemoryLeak
    expr: rate(pyroscope_memory_usage_bytes[5m]) > 0
    for: 10m
    annotations:
      summary: "Potential memory leak in {{ $labels.service }}"
  
  - alert: PyroscopeS3StorageCritical
    expr: pyroscope_storage_disk_usage_bytes / 1e9 > 90
    for: 5m
    annotations:
      summary: "S3 storage nearly full"
```

## Vận Hành và Theo Dõi

### Monitoring Metrics

**Các Metrics Quan Trọng:**

```
pyroscope_ingestion_profiles_total       - Tổng profiles nhận được
pyroscope_ingestion_queue_length         - Độ dài queue
pyroscope_storage_disk_usage_bytes       - Dung lượng disk của S3
pyroscope_api_request_duration_seconds   - API response time
pyroscope_memory_usage_bytes             - Memory usage của Pyroscope
```

### Alerting Rules

```yaml
groups:
- name: pyroscope_alerts
  rules:
  - alert: HighIngestionLatency
    expr: histogram_quantile(0.99, pyroscope_ingestion_duration_seconds) > 1
    for: 5m
  
  - alert: HighMemoryUsage
    expr: pyroscope_memory_usage_bytes > 2e9
    for: 5m
  
  - alert: S3StorageFull
    expr: pyroscope_storage_disk_usage_bytes / 1e11 > 0.9
    for: 5m
```

### Maintenance Tasks

**Hàng ngày:**
- Kiểm tra disk space
- Kiểm tra pod status
- Xem logs cho errors

**Hàng tuần:**
- Review performance trends
- Kiểm tra retention policy
- Verify S3 connectivity

**Hàng tháng:**
- Analyze storage growth
- Optimize retention settings
- Security audit

### Troubleshooting

#### Issue: Agents không kết nối

```bash
# Kiểm tra service
kubectl get svc -n profiling

# Kiểm tra DNS
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- \
  nslookup pyroscope.profiling.svc.cluster.local

# Kiểm tra logs
kubectl logs -n profiling -l app=pyroscope
```

#### Issue: Storage đầy

```bash
# Kiểm tra usage
kubectl exec -it -n monitoring deployment/pyroscope -- \
  df -h /var/lib/pyroscope

# Giảm retention period
# Edit ConfigMap và redeploy
```

#### Issue: High latency

```bash
# Kiểm tra queue length
curl http://pyrosc  ope.monitoring.svc.cluster.local:4041/metrics | grep queue

# Scale up resources hoặc giảm sample rate
```

---

## Khắc Phục Sự Cố

### Các Vấn Đề Phổ Biến

#### 1. Connection Refused
```
Lỗi: Connection refused
Nguyên nhân: Pod không chạy hoặc port sai
Giải pháp:
  1. kubectl get pods -n profiling
  2. kubectl logs -n profiling -l app=pyroscope
  3. Kiểm tra port: 4040 (HTTP), 4041 (Metrics)
```

#### 2. Out of Memory
```
Lỗi: OOMKilled
Nguyên nhân: Memory limit quá thấp
Giải pháp:
  1. Tăng memory limit trong Deployment
  2. Giảm sample rate
  3. Giảm retention period
```

#### 3. S3 Connection Error
```
Lỗi: S3 connection timeout
Nguyên nhân: MinIO không accessible hoặc credentials sai
Giải pháp:
  1. Kiểm tra MinIO pod: kubectl get pods -n profiling
  2. Kiểm tra credentials trong Secret
  3. Kiểm tra bucket: mc ls minio/pyroscope-profiles
```

#### 4. High Query Latency
```
Lỗi: Flame graph loading slowly
Nguyên nhân: Dữ liệu lớn hoặc resource constraints
Giải pháp:
  1. Giảm time range trong query
  2. Tăng CPU/Memory limits
  3. Kiểm tra S3 performance
```

### Debug Mode

```bash
# Enable debug logging
kubectl set env deployment/pyroscope LOG_LEVEL=debug -n profiling

# Check Pod status
kubectl describe pod -n profiling -l app=pyroscope

# Check events
kubectl get events -n profiling --sort-by='.lastTimestamp'

# Port-forward để test locally
kubectl port-forward -n profiling svc/pyroscope 4040:4040

# Test endpoint
curl -v http://localhost:4040/healthz
```

---
