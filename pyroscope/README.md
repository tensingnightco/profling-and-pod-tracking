# Pyroscope - Tài liệu Kỹ Thuật Toàn Diện
# Pyroscope - Comprehensive Technical Documentation


## 📖 Giới thiệu / Introduction

Này là collection đầy đủ của technical documentation cho **Pyroscope by Grafana** - một công cụ continuous profiling mạnh mẽ để tối ưu hóa hiệu suất ứng dụng.

This is a comprehensive collection of technical documentation for **Pyroscope by Grafana** - a powerful continuous profiling tool to optimize application performance.

---

## ⚠️ IMPORTANT: Kubernetes + Air Gap ONLY / MỤC ĐÍCH: CHỈ K8s + Air Gap

> **🎯 Quyết định Quan trọng:**  
> Tài liệu này được tối ưu hóa cho **ONLY Kubernetes (K8s) + Air Gap Environment**.
>
> **Điều này có nghĩa:**
> - ✅ **CHỈ hỗ trợ Kubernetes** - Đây là phương pháp triển khai duy nhất
> - ✅ **KHÔNG sử dụng Docker Compose** - Loại bỏ hoàn toàn
> - ✅ **KHÔNG sử dụng Linux Binary** - Không có systemd/standalone binary deployment
> - ✅ **KHÔNG sử dụng Source Build** - Chỉ sử dụng pre-built image
> - ✅ **Air Gap Environment** - Mạng bị cô lập, không internet
> - ✅ **Private Registry REQUIRED** - Phải có internal Docker registry
>
> **Lộ trình triển khai của bạn:**
> 1. Pre-download Pyroscope image (trên máy có internet)
> 2. Transfer đến air gap environment
> 3. Load vào private Docker registry
> 4. Deploy bằng K8s manifests (ConfigMap, PVC, Deployment, Service)
> 5. Configure agents trong application pods
>
> **Tất cả các documentation chỉ cung cấp K8s + Air Gap guidance**

### Air Gap Environment Setup Tips:

```bash
# 1. Pre-download Pyroscope image
docker pull grafana/pyroscope:latest
docker save grafana/pyroscope:latest -o pyroscope.tar

# 2. Transfer to air gap environment
scp pyroscope.tar user@airgap-server:/tmp/

# 3. Load image in air gap
docker load -i pyroscope.tar

# 4. Tag for private registry
docker tag grafana/pyroscope:latest registry.internal/pyroscope:latest
docker push registry.internal/pyroscope:latest
```

See [INSTALLATION.md](INSTALLATION.md) for detailed air gap instructions.

---

## 📚 Tài liệu chính / Main Documentation

### 1. **[TECHNICAL_DOCS.md](TECHNICAL_DOCS.md)** - Tài liệu Kỹ thuật Chính
   - ✅ Architecture & Design overview
   - ✅ System requirements (Hardware, OS, Software)
   - ✅ Pre-deployment checklist & security considerations
   - ✅ Complete installation guide (4 methods)
   - ✅ Detailed configuration reference
   - ✅ API documentation with examples
   - ✅ Developer guide & building from source
   - ✅ Post-deployment operations
   - ✅ Monitoring & alerting setup
   - ✅ Maintenance & backup procedures
   - ✅ Performance optimization tips
   - ✅ Troubleshooting guide

📖 **Nên đọc trước:** Begin here for comprehensive overview
⏱️ **Reading time:** 60-90 minutes

### 2. **[INSTALLATION.md](INSTALLATION.md)** - Hướng dẫn Cài đặt Chi tiết (K8s ONLY)
   - ✅ Kubernetes deployment (única phương pháp được hỗ trợ)
   - ✅ Air gap environment setup
   - ✅ Private registry configuration
   - ✅ Image pre-download & transfer procedures
   - ✅ K8s manifests (ConfigMap, PVC, Deployment, Service)
   - ✅ Step-by-step deployment instructions
   - ✅ Prerequisites & network checklist
   - ✅ Verification procedures
   - ✅ K8s-specific troubleshooting

📖 **Nên đọc:** For K8s deployment (only method)
⏱️ **Reading time:** 45-60 minutes

### 3. **[AGENT_INTEGRATION.md](AGENT_INTEGRATION.md)** - Hướng dẫn Tích hợp Agent
   - ✅ Integration for 6+ programming languages:
     - Python (Django, FastAPI, Flask)
     - Go
     - Node.js (Express)
     - Java (Spring Boot)
     - Rust
     - Linux eBPF (no-code profiling)
   - ✅ Basic & advanced configuration
   - ✅ Custom tagging strategies
   - ✅ Production deployment patterns
   - ✅ Best practices & tips
   - ✅ Troubleshooting agent issues

📖 **Nên đọc:** For integrating with applications
⏱️ **Reading time:** 40-50 minutes

### 4. **[QUICK_REFERENCE.md](QUICK_REFERENCE.md)** - Hướng dẫn Tham chiếu Nhanh (K8s Focus)
   - ✅ K8s commands (kubectl, deployment operations)
   - ✅ Air gap deployment commands
   - ✅ API endpoints quick table
   - ✅ Configuration checklist
   - ✅ Key metrics to monitor
   - ✅ Common tasks & solutions
   - ✅ Curl examples for K8s testing
   - ✅ K8s troubleshooting quick fixes
   - ✅ Environment variables reference
   - ✅ Alert rules template

📖 **Nên đọc:** For daily reference & common K8s tasks
⏱️ **Reading time:** 15-20 minutes (reference only)

---

## 🚀 Khởi động nhanh K8s + Air Gap / Quick Start (K8s Only)

**⚠️ Deployment method: Kubernetes + Air Gap Environment ONLY**

### Step 1: Pre-Download Image (on Internet-Connected Machine)

```bash
# Pull specific version (not latest)
docker pull grafana/pyroscope:v0.37.0

# Save as compressed archive
docker save grafana/pyroscope:v0.37.0 | gzip > pyroscope-v0.37.0.tar.gz

# Create checksum for verification
sha256sum pyroscope-v0.37.0.tar.gz > pyroscope-v0.37.0.tar.gz.sha256
```

### Step 2: Transfer to Air Gap Environment

```bash
# Transfer via USB/SCP/Physical media to air gap environment
scp pyroscope-v0.37.0.tar.gz* user@airgap-env:/tmp/
```

### Step 3: Load to Private Registry (in Air Gap)

```bash
# Verify checksum
sha256sum -c pyroscope-v0.37.0.tar.gz.sha256

# Load image
docker load -i pyroscope-v0.37.0.tar.gz

# Tag for private registry
docker tag grafana/pyroscope:v0.37.0 registry.internal/pyroscope:v0.37.0

# Push to registry
docker push registry.internal/pyroscope:v0.37.0
```

### Step 4: Deploy to Kubernetes

```bash
# Apply K8s manifests (see INSTALLATION.md for full details)
kubectl apply -f k8s-configmap.yaml
kubectl apply -f k8s-pvc.yaml
kubectl apply -f k8s-deployment.yaml
kubectl apply -f k8s-service.yaml

# Verify deployment
kubectl get deployment pyroscope -n monitoring
kubectl get pods -n monitoring
```

### Step 5: Access & Verify

```bash
# Port-forward for testing
kubectl port-forward -n monitoring svc/pyroscope 4040:4040

# Test API
curl http://localhost:4040/api/v1/service-instances
```

**📖 Full guide:** See [INSTALLATION.md - Kubernetes](INSTALLATION.md#path-2-kubernetes-installation)

---

## 📋 What You'll Learn / Những gì bạn sẽ học

| Topic | Doc | Time |
|-------|-----|------|
| Architecture & Design | TECHNICAL_DOCS | 10 min |
| System Requirements | TECHNICAL_DOCS | 5 min |
| Pre-deployment Planning | TECHNICAL_DOCS | 15 min |
| Installation (Pick 1 method) | INSTALLATION | 30-45 min |
| Agent Integration | AGENT_INTEGRATION | 20-40 min |
| Configuration Tuning | QUICK_REFERENCE | 10-15 min |
| Monitoring & Alerting | TECHNICAL_DOCS | 15-20 min |
| Troubleshooting | QUICK_REFERENCE | 5-10 min |

---

## 🎯 Learning Paths / Lộ trình học tập

**⚠️ Note:** All paths assume K8s + Air Gap deployment. No other deployment methods are supported.

Choose your learning path based on your role:

### 👨💼 DevOps/SRE Engineer (K8s Deployment)
1. Read: [README_DOCS.md](README_DOCS.md#-lưu-ý-môi-trường--environment-notes) - K8s & Air Gap section (15 min)
2. Read: [TECHNICAL_DOCS.md](TECHNICAL_DOCS.md) - Architecture section (30 min)
3. Read: [INSTALLATION.md](INSTALLATION.md#path-2-kubernetes-installation) - K8s section (45 min)
4. Setup: Deploy Pyroscope to K8s (1-2 hours)
5. Reference: Keep [QUICK_REFERENCE.md](QUICK_REFERENCE.md) handy
6. Setup: Configure monitoring & alerting (30 min)
7. Operations: Review post-deployment steps (30 min)

**Total Time: 3-4 hours**

### 👨💻 Application Developer
1. Read: [AGENT_INTEGRATION.md](AGENT_INTEGRATION.md) - Your language (20-30 min)
2. Setup: Integrate agent into your app (30-60 min)
3. Test: Deploy to K8s cluster with Pyroscope (30 min)
4. Deploy: Push code with profiling enabled
5. Monitor: Use profiles to optimize (ongoing)

**Total Time: 2-3 hours initial setup**

### 🏛️ DevOps Team Lead / Architect (K8s Strategy)
1. Read: [README_DOCS.md](README_DOCS.md) - Full overview (30 min)
2. Read: [TECHNICAL_DOCS.md](TECHNICAL_DOCS.md) - All sections (90 min)
3. Review: [INSTALLATION.md](INSTALLATION.md#path-2-kubernetes-installation) - K8s section (60 min)
4. Plan: Create K8s deployment strategy (1-2 hours)
5. Document: Customize docs for your team (1-2 hours)
6. Train: Hold team session on K8s deployment (1 hour)

**Total Time: 5-7 hours**

### 🏭 K8s & Air Gap Environment Specialist (Your Profile)
**For Kubernetes + Air Gap deployment:**

1. **Foundation (1.5 hours)**
   - Read: [README_DOCS.md](README_DOCS.md#-lưu-ý-môi-trường--environment-notes) - K8s & Air Gap section (10 min)
   - Read: [TECHNICAL_DOCS.md](TECHNICAL_DOCS.md) - Architecture & K8s focus (30 min)
   - Review: [INSTALLATION.md](INSTALLATION.md#path-2-kubernetes-installation) - K8s section (30 min)
   - Understand: Network & security constraints (20 min)

2. **Preparation (2 hours)**
   - Identify: Private Docker registry details
   - Prepare: Air gap transfer method (USB, local mirror, etc.)
   - Plan: K8s resource allocation
   - Document: Your environment specifics

3. **Implementation (3-4 hours)**
   - Customize: K8s manifests for your environment
   - Download: Pre-download images on internet-connected machine
   - Transfer: Move images to air gap environment
   - Deploy: Apply K8s manifests
   - Verify: Test all components

4. **Integration & Testing (2 hours)**
   - Setup: Agent integration in app pods
   - Network: Verify service-to-service communication
   - Monitoring: Configure Prometheus & alerting
   - Test: End-to-end profiling workflow

5. **Operations (1 hour)**
   - Runbooks: Create troubleshooting guides
   - Documentation: Customize docs for your team
   - Training: Share knowledge with team

**Total Time: 9-11 hours (one-time setup)**

**Key Documents for Your Path:**
- [INSTALLATION.md - K8s Section](INSTALLATION.md#path-2-kubernetes-installation)
- [TECHNICAL_DOCS.md - Cấu hình Section](TECHNICAL_DOCS.md#cấu-hình)
- [AGENT_INTEGRATION.md - Your App Languages](AGENT_INTEGRATION.md)
- [QUICK_REFERENCE.md - K8s Commands](QUICK_REFERENCE.md#kubernetes)

**Special Considerations for Your Environment:**
- ✅ Private Docker registry setup & authentication
- ✅ Network connectivity constraints
- ✅ DNS resolution in closed networks
- ✅ TLS certificates for internal communication
- ✅ Resource quotas & limits for K8s
- ✅ PersistentVolume strategies
- ✅ Backup procedures without cloud storage

**Special Considerations for Your Environment:**
- ✅ Private Docker registry setup & authentication
- ✅ Network connectivity constraints
- ✅ DNS resolution in closed networks
- ✅ TLS certificates for internal communication
- ✅ Resource quotas & limits for K8s
- ✅ PersistentVolume strategies
- ✅ Backup procedures without cloud storage

---

## 🔑 Key Concepts / Khái niệm Chính

### What is Continuous Profiling?
Profiling là việc thu thập data về cách ứng dụng sử dụng CPU, memory, và resources khác. Continuous profiling làm điều này non-stop, 24/7.

**Why it matters:**
- Tìm performance bottlenecks
- Giảm CPU & memory usage
- Tối ưu chi phí cloud
- Phát hiện issues trước khi ảnh hưởng users

### How Pyroscope Works

```
Your Application
       ↓
  (Pyroscope Agent SDK)
       ↓
  Collects stack traces
  (Sampling, ~100 Hz)
       ↓
  Sends to Server
  (4040/tcp)
       ↓
  Server:
  - Receives profiles
  - Compresses (100:1)
  - Stores in Badger DB
  - Makes queryable
       ↓
  Web UI
  - Flame graphs
  - Analysis
  - Sharing
       ↓
  Developer
  - Views results
  - Optimizes code
```

### Key Components

1. **Pyroscope Server** - Central collection & storage point
2. **Pyroscope Agents** - SDKs in your apps that collect data
3. **Badger DB** - High-performance storage for profiles
4. **Web UI** - Interactive visualization & analysis
5. **API** - REST endpoints for programmatic access

---

## 📈 Use Cases / Trường hợp sử dụng

### Performance Optimization
```
Before: API response time: 500ms
Analysis: Find hotspot in database queries
Optimize: Add caching
After: API response time: 100ms
Savings: 4x faster, 400% improvement
```

### Cost Reduction
```
Before: Backend consumes 8 CPU cores
Find: Most time in unused features
Optimize: Remove dead code
After: Consumes 3 CPU cores
Savings: ~60% reduction in cloud costs
```

### Scaling Issues
```
Before: App not scaling past 1000 requests/sec
Find: Lock contention in user service
Optimize: Better concurrency patterns
After: Scales to 5000 requests/sec
Improvement: 5x capacity increase
```

### Memory Leaks
```
Before: Memory grows over time
Find: Unbounded cache growth
Fix: Add eviction policy
After: Stable memory usage
Prevents: OOM crashes in production
```

---

## 🏗️ Architecture Overview / Tổng quan kiến trúc

```
┌─────────────────────────────────────────────────┐
│        Production Environment                   │
│  ┌──────────────┐  ┌──────────────┐             │
│  │   App 1      │  │   App N      │  ...        │
│  │ + Pyroscope  │  │ + Pyroscope  │             │
│  │   Agent      │  │   Agent      │             │
│  └──────┬───────┘  └──────┬───────┘             │
└─────────┼──────────────────┼────────────────────┘
          │ HTTP/gRPC       │
          │ (4040/tcp)      │
          │                 │
        ┌─▼─────────────────▼──────┐
        │  Pyroscope Server        │
        │  ┌────────────────────┐  │
        │  │ API Handler        │  │
        │  ├────────────────────┤  │
        │  │ Storage (Badger DB)│  │
        │  ├────────────────────┤  │
        │  │ Query Engine       │  │
        │  └────────────────────┘  │
        └─┬──────────────────┬─────┘
          │                  │
        ┌─▼────────────────┐ │
        │ Web UI (React)   │ │
        │ - Flame Graphs   │ │
        │ - Analysis       │ │
        └──────────────────┘ │
                             │
                    ┌────────▼────────┐
                    │ Optional:       │
                    │ - Promotheus    │
                    │ - Grafana       │
                    │ - Loki (logs)   │
                    └─────────────────┘
```

---

## 📊 Resource Requirements / Yêu cầu tài nguyên

### Minimum (Development)
```
CPU:     2 cores
RAM:     2GB
Storage: 10GB
Network: 100 Mbps
```

### Recommended (Production)
```
CPU:      8+ cores
RAM:      16GB+
Storage:  100GB+ SSD
Network:  1 Gbps+
```

### Typical Storage Growth
```
Per application:     ~800 MB/day
30-day retention:    ~24 GB
365-day retention:   ~292 GB (with compression)
```

---

## 🔐 Security Considerations / Cân nhắc bảo mật

- ✅ Pyroscope should NOT be exposed to public internet
- ✅ Use private network or VPN only
- ✅ Enable TLS/SSL if accessed over network
- ✅ Implement authentication if multi-tenant
- ✅ Regular backups to prevent data loss
- ✅ Monitor access logs
- ✅ Use resource limits to prevent DoS
- ✅ Keep Pyroscope updated with security patches

See [TECHNICAL_DOCS.md](TECHNICAL_DOCS.md) for detailed security hardening steps.

---

## 🚨 Common Issues & Solutions / Vấn đề thường gặp

| Issue | Quick Fix | More Info |
|-------|-----------|-----------|
| Port already in use | Change port or kill process | [INSTALLATION.md](INSTALLATION.md#issue-port-already-in-use) |
| Can't connect | Check firewall & network | [QUICK_REFERENCE.md](QUICK_REFERENCE.md#troubleshooting-quick-fixes) |
| Out of disk | Reduce retention or add storage | [TECHNICAL_DOCS.md](TECHNICAL_DOCS.md#issue-disk-is-full) |
| No agent data | Verify connectivity & config | [AGENT_INTEGRATION.md](AGENT_INTEGRATION.md#issue-agent-cant-connect-to-server) |
| High memory | Reduce sample rate | [AGENT_INTEGRATION.md](AGENT_INTEGRATION.md#issue-high-cpumemory-overhead) |

See [TECHNICAL_DOCS.md](TECHNICAL_DOCS.md#khắc-phục-sự-cố) for complete troubleshooting guide.

---

## 🏭 K8s & Air Gap Best Practices / Thực hành tốt nhất cho K8s & Air Gap

### Pre-Deployment

**Image Management:**
- ✅ Pre-download images on internet-connected machine
- ✅ Use specific version tags (not `latest`)
- ✅ Store images in compressed format (.tar.gz) for transfer
- ✅ Verify image checksums before loading
- ✅ Maintain inventory of all images used

```bash
# Best practice: Use stable versions
docker pull grafana/pyroscope:v0.37.0
docker save grafana/pyroscope:v0.37.0 | gzip > pyroscope-v0.37.0.tar.gz

# Verify checksum
sha256sum pyroscope-v0.37.0.tar.gz > pyroscope-v0.37.0.tar.gz.sha256
```

**Network Planning:**
- ✅ Map all services that need to communicate
- ✅ Configure DNS properly (no external DNS calls)
- ✅ Plan for internal service discovery
- ✅ Document port requirements
- ✅ Setup private Docker registry beforehand

### Deployment

**Kubernetes Configuration:**
- ✅ Use ImagePullPolicy: IfNotPresent (don't pull from internet)
- ✅ Set resource requests/limits based on actual environment
- ✅ Use PersistentVolumes with appropriate storage class
- ✅ Configure NetworkPolicies for restricted communication
- ✅ Use Namespaces to isolate deployments

```yaml
# Recommended K8s settings for air gap
imagePullPolicy: IfNotPresent
resources:
  requests:
    cpu: 500m
    memory: 1Gi
  limits:
    cpu: 2000m
    memory: 4Gi
```

**Registry Configuration:**
- ✅ Setup private registry with proper authentication
- ✅ Use internal registry URL in manifests
- ✅ Create docker pull secrets in K8s
- ✅ Test registry access before deployment

```yaml
# ImagePullSecret for private registry (if needed)
imagePullSecrets:
- name: private-registry-secret
```

### Operations

**Monitoring & Metrics:**
- ✅ Setup Prometheus metrics collection (internal only)
- ✅ No external metric export
- ✅ Use internal dashboards (Grafana in cluster)
- ✅ Archive old metrics locally

**Backup Strategy:**
- ✅ Use PersistentVolume snapshots
- ✅ Regular full backups to local storage
- ✅ Test restore procedures regularly
- ✅ Keep backups on-site

```bash
# Backup PersistentVolume (example with kubectl)
kubectl get pvc -n monitoring pyroscope-data -o yaml > pvc-backup.yaml
```

**Agent Communication:**
- ✅ Use internal service DNS: `http://pyroscope.monitoring.svc.cluster.local:4040`
- ✅ No external network calls from agents
- ✅ Configure agents for internal retry policies
- ✅ Monitor agent heartbeats

### Security for Air Gap

- ✅ No auto-updates enabled
- ✅ Manual patching process
- ✅ Restricted file system access (read-only where possible)
- ✅ Resource quotas enforced
- ✅ Network policies restrict traffic
- ✅ No telemetry/phone-home enabled
- ✅ Audit logging for compliance

### Common Air Gap Challenges & Solutions

| Challenge | Solution |
|-----------|----------|
| Can't download from GitHub | Pre-download on internet machine, transfer via USB/SCP |
| Private registry needs auth | Create K8s Secret with registry credentials |
| DNS not working | Use IP addresses or local hosts file entries |
| No NTP sync | Configure NTP with internal time server |
| Agent can't connect | Verify internal DNS, check NetworkPolicies, test with curl |
| Out of storage slowly | Regular cleanup of old profiles, reduce retention |

---

## 📞 Getting Help / Nhận trợ giúp

### For K8s & Air Gap Specific Issues

**Check First:**
1. Network connectivity: `kubectl exec -it pod -- curl http://pyroscope:4040/healthz`
2. Image availability: `kubectl describe pod <pod-name>`
3. Resource usage: `kubectl top pods -n monitoring`
4. Logs: `kubectl logs -n monitoring deployment/pyroscope`

**Common Fixes:**
- Image not found → ensure image loaded to private registry
- Can't connect → check K8s NetworkPolicies and DNS
- Out of memory → reduce sample rate or retention

### Official Resources
- 📖 [Pyroscope Official Docs](https://pyroscope.io/docs)
- 💬 [GitHub Discussions](https://github.com/grafana/pyroscope/discussions)
- 🐛 [GitHub Issues](https://github.com/grafana/pyroscope/issues)
- 🔗 [Grafana Integration Docs](https://grafana.com/docs/grafana/latest/datasources/pyroscope/)

### This Documentation
- 📄 [Technical Documentation](TECHNICAL_DOCS.md) - Complete reference
- 🎯 [Installation Guide](INSTALLATION.md) - Step-by-step setup (K8s focused)
- 🔌 [Agent Integration](AGENT_INTEGRATION.md) - Application integration
- ⚡ [Quick Reference](QUICK_REFERENCE.md) - Commands & troubleshooting

---

## 🛠️ Support & Diagnosis / Hỗ trợ & chẩn đoán

### Before Contacting Support
1. Check logs: `journalctl -u pyroscope -f` (Linux) or `docker logs pyroscope` (Docker)
2. Review: [QUICK_REFERENCE.md](QUICK_REFERENCE.md#troubleshooting-quick-fixes)
3. Search: [GitHub Issues](https://github.com/grafana/pyroscope/issues)
4. Collect diagnostics:
   ```bash
   mkdir -p pyroscope-diagnostics
   docker logs pyroscope > pyroscope-diagnostics/logs.txt
   curl http://localhost:4041/metrics > pyroscope-diagnostics/metrics.txt
   df -h > pyroscope-diagnostics/disk.txt
   ```

### When Contacting Support
Include:
- Operating System & Version
- Docker/Kubernetes/Binary installation
- Pyroscope version
- Application being profiled
- Error messages & logs
- System metrics (CPU, memory, disk)
- Network configuration

---

## 📚 Documentation Structure / Cấu trúc tài liệu

```
pyroscope/
├── README.md                    ← You are here
├── TECHNICAL_DOCS.md            ← Full technical reference
├── INSTALLATION.md              ← Installation & deployment
├── AGENT_INTEGRATION.md         ← Application integration
├── QUICK_REFERENCE.md           ← Commands & troubleshooting
└── [config examples]            ← Config templates (optional)
```

---

## 📝 Document Versions & Updates / Phiên bản & cập nhật

| Document | Version | Last Updated |
|----------|---------|--------------|
| TECHNICAL_DOCS.md | 1.0 | April 2026 |
| INSTALLATION.md | 1.0 | April 2026 |
| AGENT_INTEGRATION.md | 1.0 | April 2026 |
| QUICK_REFERENCE.md | 1.0 | April 2026 |
| README.md | 1.0 | April 2026 |

**Note:** Check [Pyroscope GitHub Releases](https://github.com/grafana/pyroscope/releases) for latest version info.

---

## 🗺️ Recommended Reading Order / Thứ tự đọc được khuyến nghị

### For First-Time Users
1. **This README** (Current) - 10 minutes
2. **[INSTALLATION.md](INSTALLATION.md)** - Quick Start section - 5 minutes
3. **Deploy & Test** - 30 minutes
4. **[AGENT_INTEGRATION.md](AGENT_INTEGRATION.md)** - Your language - 20 minutes
5. **Integrate with your app** - 30 minutes
6. **[QUICK_REFERENCE.md](QUICK_REFERENCE.md)** - For daily reference - 15 minutes

### For Operators/SREs
1. **[TECHNICAL_DOCS.md](TECHNICAL_DOCS.md)** - All sections - 90 minutes
2. **[INSTALLATION.md](INSTALLATION.md)** - Your deployment method - 45 minutes
3. **Deploy & Monitor** - 1-2 hours
4. **[QUICK_REFERENCE.md](QUICK_REFERENCE.md)** - Keep bookmarked
5. **Setup alerting** - [TECHNICAL_DOCS.md#monitoring--alerting](TECHNICAL_DOCS.md#monitoring--alerting)

### For Application Developers
1. **This README** - Introduction - 10 minutes
2. **[AGENT_INTEGRATION.md](AGENT_INTEGRATION.md)** - Your language - 20 minutes
3. **Integrate SDK** - 30 minutes
4. **Test locally** - 30 minutes
5. **Deploy & profile** - Ongoing
6. **[QUICK_REFERENCE.md](QUICK_REFERENCE.md)** - For optimization tips

---

## ✅ Pre-Deployment Checklist / Danh sách kiểm tra trước triển khai

### Planning Phase
- [ ] Reviewed system requirements: [TECHNICAL_DOCS.md](TECHNICAL_DOCS.md#yêu-cầu-hệ-thống)
- [ ] Identified deployment method (Docker/K8s/Binary/Source)
- [ ] Estimated storage requirements
- [ ] Planned retention policy
- [ ] Reviewed security considerations

### Preparation Phase
- [ ] Infrastructure ready (VM/Container/K8s cluster)
- [ ] Storage/volumes provisioned
- [ ] Network rules configured
- [ ] Backup strategy established
- [ ] Monitoring/alerting system ready

### Deployment Phase
- [ ] Followed [INSTALLATION.md](INSTALLATION.md) for chosen method
- [ ] Verified service is running
- [ ] Accessed Web UI successfully
- [ ] API endpoints responding
- [ ] Monitoring/alerting configured

### Post-Deployment Phase
- [ ] Service scaling tested
- [ ] Backup/restore tested
- [ ] Agent integration verified
- [ ] Performance baseline established
- [ ] Team trained on usage

---

## 💡 Best Practices / Thực hành tốt nhất

### Deployment
- Always use persistent storage
- Set up monitoring from day 1
- Plan retention policy carefully
- Regular backups are essential
- Document your setup

### Operations
- Monitor disk space actively
- Review profiles regularly
- Keep Pyroscope updated
- Have runbooks for common issues
- Share learnings with team

### Development
- Use appropriate sample rates
- Tag profiles with context
- Profile early & often
- Share findings with team
- Use profiles for optimization, not debugging

---

## 🎓 Learning Resources / Tài nguyên học tập

## Free Resources
- 📖 [Pyroscope Official Documentation](https://pyroscope.io/docs)
- 📹 [Grafana Labs YouTube](https://www.youtube.com/c/GrafanaLabs)
- 💬 [Community Discussions](https://github.com/grafana/pyroscope/discussions)
- 📝 [Blog Posts](https://pyroscope.io/blog)

### Paid Resources (Optional)
- 🎓 [Grafana Training Courses](https://grafana.com/training/)
- 👥 [Grafana Professional Services](https://grafana.com/services/)

---

## 📄 License & Attribution / Giấy phép & ghi công

This documentation is created to support Pyroscope deployment and usage.

- **Pyroscope** is developed by [Grafana Labs](https://grafana.com)
- **License:** AGPL-3.0
- **GitHub:** https://github.com/grafana/pyroscope
- **Website:** https://pyroscope.io

---

## 🚀 K8s & Air Gap Environment - Quick Start / Khởi động nhanh cho K8s & Air Gap

**For your K8s + Air Gap setup, follow this optimized path:**

### Step 1: Pre-Download Images (Air Gap)
```bash
# On a machine with internet access
docker pull grafana/pyroscope:latest
docker save grafana/pyroscope:latest -o pyroscope-image.tar.gz

# Transfer to air gap environment
scp pyroscope-image.tar.gz user@airgap-env:/tmp/
```

### Step 2: Load Image to Private Registry (Air Gap)
```bash
# On air gap environment
docker load -i pyroscope-image.tar.gz
docker tag grafana/pyroscope:latest <your-registry>/pyroscope:latest
docker push <your-registry>/pyroscope:latest
```

### Step 3: Deploy to Kubernetes
```bash
# Update k8s manifests to use private registry
# In k8s-deployment.yaml, change image to:
# image: <your-registry>/pyroscope:latest
# imagePullPolicy: IfNotPresent

kubectl apply -f k8s-configmap.yaml
kubectl apply -f k8s-pvc.yaml
kubectl apply -f k8s-deployment.yaml
kubectl apply -f k8s-service.yaml

# Verify deployment
kubectl get deployment pyroscope -n monitoring
kubectl get pods -n monitoring
```

### Step 4: Verify & Access
```bash
# Check if running
kubectl get pods -n monitoring -l app=pyroscope

# Port-forward
kubectl port-forward -n monitoring svc/pyroscope 4040:4040

# Test: curl http://localhost:4040/healthz
```

**📖 Full K8s Guide:** [INSTALLATION.md - Kubernetes](INSTALLATION.md#path-2-kubernetes-installation)

---

## 🎯 Next Steps / Bước tiếp theo

### For K8s + Air Gap Deployments (Your Environment):

1. **Prepare air gap environment:**
   - [ ] Check network connectivity constraints
   - [ ] Identify local Docker registry (or setup private registry)
   - [ ] List all required images (Pyroscope + dependencies)
   
2. **Pre-download & transfer:**
   - [ ] Download `grafana/pyroscope:latest` on internet-connected machine
   - [ ] Transfer tar files to air gap environment
   - [ ] Load images to private registry
   
3. **Deploy to Kubernetes:**
   - [ ] Customize K8s manifests ([INSTALLATION.md](INSTALLATION.md#path-2-kubernetes-installation))
   - [ ] Update image references to private registry
   - [ ] Apply ConfigMap, PVC, Deployment, Service
   - [ ] Verify using `kubectl` commands
   
4. **Configure agents in K8s:**
   - [ ] Deploy agents in your application pods
   - [ ] Point agents to: `http://pyroscope:4040` (internal service)
   - [ ] Verify data ingestion in UI
   
5. **Setup monitoring:**
   - [ ] Configure Prometheus scraping (4041 metrics port)
   - [ ] Create dashboard/alerts for your environment
   - [ ] Test alert triggers

**⏱️ Estimated time:** 2-3 hours

---

## 📞 Contact & Support / Liên hệ và Hỗ trợ

Have questions or need help?

- 🐛 **Found a bug?** → [GitHub Issues](https://github.com/grafana/pyroscope/issues)
- 💬 **Have a question?** → [GitHub Discussions](https://github.com/grafana/pyroscope/discussions)
- 📖 **Need docs?** → [Pyroscope Docs](https://pyroscope.io/docs)
- 🏢 **Enterprise support?** → [Grafana Sales](https://grafana.com/about/contact-sales/)

---

**Ready to get started? → [Go to INSTALLATION.md (K8s Only) →](INSTALLATION.md#path-2-kubernetes-installation)**

---

**Document Version:** 1.0  
**Last Updated:** April 2026  
**Maintained By:** DevOps & Documentation Team  
**Pyroscope Version:** 0.37.0+  
**Deployment Method:** Kubernetes + Air Gap Environment ONLY  
**Status:** ✅ Complete & Production-Ready
