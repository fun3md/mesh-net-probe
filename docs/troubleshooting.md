# Troubleshooting Guide

This guide helps you diagnose and resolve common issues with the Mesh Net Probe across different platforms and architectures.

## Table of Contents

1. [Platform-Specific Issues](#platform-specific-issues)
2. [Network Configuration Problems](#network-configuration-problems)
3. [Permission Errors](#permission-errors)
4. [Performance Issues](#performance-issues)
5. [Configuration Problems](#configuration-problems)
6. [Container Deployment Issues](#container-deployment-issues)
7. [Diagnostic Commands](#diagnostic-commands)
8. [Log Analysis](#log-analysis)

## Platform-Specific Issues

### Windows Issues

#### "Access Denied" for ICMP Operations

**Symptoms**: 
- ICMP measurements fail immediately
- "operation not permitted" errors
- Probe starts but cannot send ping packets

**Diagnosis**:
```cmd
# Check if running as Administrator
net session

# Check Windows version
winver

# Test ICMP manually
ping 127.0.0.1
```

**Solutions**:
1. **Run as Administrator**:
   ```cmd
   # Right-click Command Prompt → "Run as administrator"
   probe.exe start --config config.json
   ```

2. **Configure Windows Firewall**:
   ```cmd
   # Allow ICMP echo requests
   netsh advfirewall firewall add rule name="Mesh Probe ICMP" dir=in action=allow protocol=icmpv4:8,any

   # Allow probe ports
   netsh advfirewall firewall add rule name="Mesh Probe UDP" dir=in action=allow protocol=udp localport=8080
   ```

3. **Check Group Policy**:
   - Open `gpedit.msc`
   - Navigate to: Computer Configuration → Windows Settings → Security Settings → Windows Firewall with Advanced Security
   - Ensure ICMP rules are not blocked

#### "Winsock Error 10013" - Permission Denied

**Symptoms**:
- Raw socket operations fail
- Platform detection shows "no_raw_sockets" limitation

**Diagnosis**:
```cmd
# Check network adapter status
ipconfig /all

# Test socket creation
netsh winsock show catalog
```

**Solutions**:
1. **Use alternative network stack** (automatic):
   ```json
   {
     "network": {
       "use_raw_sockets": false,
       "buffer_size": 2048
     }
   }
   ```

2. **Enable legacy socket support**:
   ```cmd
   netsh winsock reset
   netsh int ip reset
   ```

#### High Latency on Windows

**Symptoms**:
- Measurements show millisecond precision instead of microsecond
- Unusually high RTT values
- Inconsistent timing

**Diagnosis**:
```cmd
# Check timer resolution
wmic path win32_pnpentity get deviceid,status

# Monitor CPU usage
tasklist | findstr probe
```

**Solutions**:
1. **Disable Windows power saving**:
   ```cmd
   powercfg /setactive 8c5e7fda-e8bf-4a96-9a85-a6e23a8c635c
   ```

2. **Enable high-performance mode**:
   ```cmd
   powercfg /setacvalueindex SCHEME_CURRENT 0012ee47-9041-4b5d-9b69-2869589e9949 1
   powercfg /setdcvalueindex SCHEME_CURRENT 0012ee47-9041-4b5d-9b69-2869589e9949 1
   powercfg /setactive SCHEME_CURRENT
   ```

### Linux Issues

#### "Operation not permitted" for Raw Sockets

**Symptoms**:
- ICMP socket creation fails
- "permission denied" errors
- Probe falls back to limited functionality

**Diagnosis**:
```bash
# Check current user
whoami

# Check capabilities
getcap /path/to/probe

# Test raw socket creation manually
sudo tcpdump -i eth0 -c 1 icmp
```

**Solutions**:
1. **Add CAP_NET_RAW capability**:
   ```bash
   sudo setcap cap_net_raw+ep /usr/local/bin/probe
   # Verify
   getcap /usr/local/bin/probe
   ```

2. **Run as root** (legacy):
   ```bash
   sudo /usr/local/bin/probe start --config config.json
   ```

3. **Check SELinux/AppArmor**:
   ```bash
   # Check SELinux status
   sestatus

   # Check AppArmor status
   sudo aa-status
   ```

#### Container Permission Issues

**Symptoms**:
- Works outside container but fails inside
- "permission denied" even with capabilities

**Diagnosis**:
```bash
# Check container capabilities
docker run --rm -it --cap-add=NET_RAW mesh-probe probe --test-platform

# Check container user
docker run --rm -it mesh-probe whoami
```

**Solutions**:
1. **Add required capabilities**:
   ```bash
   # Docker
   docker run --cap-add=NET_RAW --cap-add=NET_ADMIN mesh-probe

   # Docker Compose
   services:
     mesh-probe:
       cap_add:
         - NET_RAW
         - NET_ADMIN
   ```

2. **Use specific user**:
   ```bash
   # Add user to necessary groups
   sudo usermod -a -G netdev $USER

   # Run container as specific user
   docker run --user 1000:1000 mesh-probe
   ```

#### Network Interface Issues

**Symptoms**:
- No network interfaces detected
- Cannot bind to specified interface
- Interface shows as down

**Diagnosis**:
```bash
# List all interfaces
ip addr show

# Check interface status
ip link show eth0

# Test interface manually
ping -I eth0 8.8.8.8
```

**Solutions**:
1. **Specify correct interface**:
   ```json
   {
     "network": {
       "interface": "eth0",
       "source_ip": "auto"
     }
   }
   ```

2. **Bring interface up**:
   ```bash
   sudo ip link set eth0 up
   ```

### macOS Issues

#### "Sandbox" Security Restrictions

**Symptoms**:
- Application cannot access network
- "Operation not permitted" errors
- App may be blocked by macOS security

**Diagnosis**:
```bash
# Check quarantine attributes
xattr -l /path/to/probe

# Check Gatekeeper status
spctl --status

# Test manual network access
netstat -rn
```

**Solutions**:
1. **Remove quarantine attribute**:
   ```bash
   xattr -rd com.apple.quarantine /path/to/probe
   ```

2. **Allow in Security & Privacy**:
   - System Preferences → Security & Privacy → General
   - Allow applications downloaded from: App Store and identified developers

3. **Disable Gatekeeper temporarily** (not recommended for production):
   ```bash
   sudo spctl --master-disable
   ```

#### BPF Permission Issues

**Symptoms**:
- Cannot access Berkeley Packet Filter
- Limited packet capture functionality

**Diagnosis**:
```bash
# Check BPF devices
ls -la /dev/bpf*

# Test BPF manually
sudo tcpdump -i any icmp -c 1
```

**Solutions**:
1. **Run with sudo**:
   ```bash
   sudo ./probe start --config config.json
   ```

2. **Grant BPF access**:
   ```bash
   # Create BPF group
   sudo groupadd bpf
   sudo usermod -a -G bpf $USER
   
   # Set BPF device permissions
   sudo chmod g+rw /dev/bpf*
   ```

#### High CPU Usage

**Symptoms**:
- Probe consuming excessive CPU
- System becoming unresponsive
- Fan spinning constantly

**Diagnosis**:
```bash
# Monitor probe CPU usage
top -pid $(pgrep probe)

# Check timing precision
./probe timing-test

# Monitor system load
uptime
```

**Solutions**:
1. **Reduce measurement frequency**:
   ```json
   {
     "targets": [
       {
         "interval": "10s",
         "timeout": "5s"
       }
     ]
   }
   ```

2. **Disable logging**:
   ```json
   {
     "telemetry": {
       "enable_logging": false
     }
   }
   ```

## Network Configuration Problems

### Invalid IP Addresses

**Symptoms**:
- Configuration validation fails
- "Invalid IP address" errors
- Targets not reachable

**Diagnosis**:
```bash
# Validate configuration
./probe validate-config --config config.json

# Test IP manually
ping -c 1 192.168.1.1

# Check IP formatting
ip addr show
```

**Solutions**:
1. **Use correct IP format**:
   ```json
   {
     "targets": [
       {
         "address": "192.168.1.1",  // Correct format
         "display_name": "Gateway"
       }
     ]
   }
   ```

2. **Avoid common mistakes**:
   - Don't use leading zeros: `192.168.001.001` ❌
   - Use decimal notation: `192.168.1.1` ✅
   - IPv6 needs brackets: `"[2001:db8::1]"` ✅

### Port Binding Issues

**Symptoms**:
- "Address already in use" errors
- Cannot bind to specified port
- Service fails to start

**Diagnosis**:
```bash
# Check port usage
netstat -tlnp | grep :8080

# Test port availability
nc -zv localhost 8080

# Find process using port
lsof -i :8080
```

**Solutions**:
1. **Use different port**:
   ```json
   {
     "network": {
       "bind_port": 8081
     }
   }
   ```

2. **Kill conflicting process**:
   ```bash
   sudo kill -9 $(lsof -t -i:8080)
   ```

### DNS Resolution Problems

**Symptoms**:
- Hostname targets fail
- DNS timeout errors
- Intermittent connectivity

**Diagnosis**:
```bash
# Test DNS resolution
nslookup example.com

# Check DNS configuration
cat /etc/resolv.conf

# Test connectivity to DNS
ping 8.8.8.8
```

**Solutions**:
1. **Use IP addresses instead of hostnames**:
   ```json
   {
     "targets": [
       {
         "address": "8.8.8.8",  // Use IP instead of "google.com"
         "display_name": "Google DNS"
       }
     ]
   }
   ```

2. **Configure DNS servers**:
   ```bash
   # Linux
   echo "nameserver 8.8.8.8" | sudo tee /etc/resolv.conf

   # macOS
   sudo dscacheutil -flushcache
   ```

## Permission Errors

### File System Permissions

**Symptoms**:
- Cannot read/write configuration files
- Data directory access denied
- Log file creation fails

**Diagnosis**:
```bash
# Check file permissions
ls -la config.json

# Check directory permissions
ls -ld /var/log/mesh-probe

# Test write access
touch /tmp/test_write
```

**Solutions**:
1. **Fix file permissions**:
   ```bash
   # Make configuration readable
   chmod 644 config.json

   # Create data directory with proper permissions
   sudo mkdir -p /var/lib/mesh-probe
   sudo chown $USER:$USER /var/lib/mesh-probe
   sudo chmod 755 /var/lib/mesh-probe
   ```

2. **Run from writable directory**:
   ```bash
   # Create local data directory
   mkdir -p ~/mesh-probe-data

   # Use relative paths in configuration
   ```

### User/Group Issues

**Symptoms**:
- Process starts but immediately exits
- "Permission denied" in logs
- User lacks necessary privileges

**Diagnosis**:
```bash
# Check user groups
groups $USER

# Check process ownership
ps aux | grep probe

# Test with different user
su - $USER -c "./probe --test-platform"
```

**Solutions**:
1. **Add user to network group**:
   ```bash
   # Linux
   sudo usermod -a -G netdev $USER

   # Log out and log back in
   ```

2. **Run with sudo** (temporary):
   ```bash
   sudo ./probe start --config config.json
   ```

## Performance Issues

### High Latency

**Symptoms**:
- Measurements show unusually high RTT
- Timing precision lower than expected
- Inconsistent measurement results

**Diagnosis**:
```bash
# Check timing precision
./probe timing-test

# Monitor system load
uptime
top

# Check network interfaces
ethtool eth0  # Linux specific
```

**Solutions**:
1. **Optimize timing**:
   ```json
   {
     "network": {
       "buffer_size": 8192,
       "ttl": 64
     }
   }
   ```

2. **Check network quality**:
   ```bash
   # Test with simple ping
   ping -c 10 8.8.8.8

   # Check for packet loss
   mtr 8.8.8.8
   ```

### High CPU Usage

**Symptoms**:
- Probe consuming excessive CPU resources
- System becomes slow
- Fan noise increases

**Diagnosis**:
```bash
# Monitor CPU usage
top -p $(pgrep probe)

# Check measurement frequency
./probe stats

# Profile timing precision
./probe timing-test --verbose
```

**Solutions**:
1. **Reduce measurement frequency**:
   ```json
   {
     "targets": [
       {
         "interval": "30s",  // Increase interval
         "timeout": "10s"    // Adjust timeout
       }
     ]
   }
   ```

2. **Disable unnecessary features**:
   ```json
   {
     "telemetry": {
       "enable_logging": false,
       "export_interval": "300s"
     }
   }
   ```

### Memory Issues

**Symptoms**:
- Out of memory errors
- System becomes slow
- Probe crashes unexpectedly

**Diagnosis**:
```bash
# Check memory usage
free -h

# Monitor probe memory
ps aux | grep probe

# Check for memory leaks
valgrind --leak-check=full ./probe
```

**Solutions**:
1. **Reduce buffer sizes**:
   ```json
   {
     "network": {
       "buffer_size": 2048
     },
     "telemetry": {
       "buffer_size": 1000
     }
   }
   ```

2. **Limit concurrent operations**:
   ```json
   {
     "optimizations": {
       "batch_size": 32
     }
   }
   ```

## Configuration Problems

### JSON Syntax Errors

**Symptoms**:
- Configuration validation fails
- "invalid JSON" errors
- Probe fails to start

**Diagnosis**:
```bash
# Validate JSON syntax
jq . config.json

# Check for common syntax errors
grep -n "," config.json
```

**Solutions**:
1. **Use JSON validator**:
   ```bash
   # Install jq for JSON validation
   sudo apt install jq  # Ubuntu/Debian
   brew install jq      # macOS
   
   # Validate and pretty-print
   jq . config.json > config_valid.json
   ```

2. **Common JSON mistakes**:
   - Trailing commas: `,]` ❌ → `]` ✅
   - Single quotes: `'key': "value"` ❌ → `"key": "value"` ✅
   - Missing quotes: `key: value` ❌ → `"key": "value"` ✅

### Missing Required Fields

**Symptoms**:
- "Missing required field" errors
- Probe starts but behaves incorrectly
- Default values used unexpectedly

**Diagnosis**:
```bash
# Validate configuration schema
./probe validate-config --strict config.json

# Check documentation for required fields
```

**Solutions**:
1. **Add missing fields**:
   ```json
   {
     "id": "required_unique_id",
     "name": "Human readable name",
     "version": 1,
     "network": {
       "ttl": 64  // Required for ICMP
     },
     "targets": []  // Required array
   }
   ```

2. **Use configuration template**:
   ```bash
   ./probe generate-config --template basic > config.json
   ```

### Invalid Values

**Symptoms**:
- "Invalid value" warnings
- Configuration accepted but probe behaves unexpectedly
- Fields ignored with warnings

**Diagnosis**:
```bash
# Check value ranges
./probe validate-config --detailed config.json

# Test specific values
```

**Solutions**:
1. **Use valid value ranges**:
   ```json
   {
     "network": {
       "ttl": 64,        // Range: 1-255
       "buffer_size": 4096  // Range: 1024-65536
     },
     "targets": [
       {
         "priority": 10,    // Range: 1-10
         "timeout": "10s"   // Valid time duration
       }
     ]
   }
   ```

## Container Deployment Issues

### Docker Issues

#### Container Cannot Access Network

**Symptoms**:
- ICMP fails inside container
- Works on host but not in container
- Network interfaces not available

**Diagnosis**:
```bash
# Check container network
docker exec mesh-probe ip addr show

# Test network connectivity
docker exec mesh-probe ping 8.8.8.8

# Check container capabilities
docker run --rm -it mesh-probe probe --test-platform
```

**Solutions**:
1. **Use host network mode**:
   ```yaml
   services:
     mesh-probe:
       network_mode: host
       cap_add:
         - NET_RAW
   ```

2. **Map network interfaces**:
   ```yaml
   services:
     mesh-probe:
       networks:
         - host_network
       cap_add:
         - NET_RAW
   
   networks:
     host_network:
       external: true
       name: host
   ```

#### Permission Issues in Container

**Symptoms**:
- Cannot read/write data volumes
- Configuration file access denied
- User permissions mismatch

**Diagnosis**:
```bash
# Check container user
docker exec mesh-probe whoami

# Check volume permissions
docker exec mesh-probe ls -la /config

# Compare host and container users
id
docker exec mesh-probe id
```

**Solutions**:
1. **Use specific user**:
   ```yaml
   services:
     mesh-probe:
       user: "1000:1000"  # Match host user
       volumes:
         - ./config:/config:ro
         - ./data:/data
   ```

2. **Fix volume permissions**:
   ```bash
   # Set proper ownership on host
   sudo chown -R 1000:1000 ./config ./data
   ```

### Kubernetes Issues

#### Pod Scheduling Problems

**Symptoms**:
- Pod fails to schedule
- "Insufficient resources" errors
- Pod remains in Pending state

**Diagnosis**:
```bash
# Check pod status
kubectl get pods -l app=mesh-probe

# Check events
kubectl describe pod mesh-probe-xxx

# Check node resources
kubectl describe nodes
```

**Solutions**:
1. **Adjust resource limits**:
   ```yaml
   resources:
     requests:
       cpu: 100m
       memory: 128Mi
     limits:
       cpu: 500m
       memory: 512Mi
   ```

2. **Use node selectors**:
   ```yaml
   nodeSelector:
     kubernetes.io/arch: amd64
   ```

#### Network Policy Restrictions

**Symptoms**:
- ICMP packets blocked
- Cannot reach external targets
- DNS resolution fails

**Diagnosis**:
```bash
# Check network policies
kubectl get networkpolicies

# Test from within pod
kubectl exec -it mesh-probe-xxx -- ping 8.8.8.8

# Check service account permissions
kubectl auth can-i --list --as=system:serviceaccount:default:mesh-probe
```

**Solutions**:
1. **Create appropriate network policies**:
   ```yaml
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: mesh-probe-allow
   spec:
     podSelector:
       matchLabels:
         app: mesh-probe
     policyTypes:
     - Egress
     - Ingress
     egress:
     - to: []
       ports:
       - protocol: ICMP
         port: 8  # ICMP type 8 (echo request)
       - protocol: UDP
         port: 53  # DNS
   ```

## Diagnostic Commands

### System Information

```bash
# Platform detection and capabilities
./probe --test-platform

# System information
uname -a
lscpu  # Linux
system_profiler SPHardwareDataType  # macOS
wmic computersystem get name,model,manufacturer  # Windows

# Network interfaces
ip addr show  # Linux
ifconfig  # macOS
ipconfig /all  # Windows

# Memory information
free -h  # Linux
vm_stat  # macOS
wmic memorychip get size,speed  # Windows
```

### Network Diagnostics

```bash
# Test connectivity
ping -c 5 8.8.8.8

# Trace route
traceroute 8.8.8.8  # Linux/macOS
tracert 8.8.8.8  # Windows

# DNS resolution
nslookup google.com

# Check routing table
route -n  # Linux
netstat -rn  # macOS
route print  # Windows

# Monitor network traffic
tcpdump -i eth0 icmp  # Linux
sudo tcpdump -i any icmp  # macOS
```

### Process Monitoring

```bash
# Check probe process
ps aux | grep probe
pidof probe

# Monitor resource usage
top -p $(pgrep probe)
htop -p $(pgrep probe)

# Check open files
lsof -p $(pgrep probe)

# Monitor system calls (Linux)
strace -p $(pgrep probe)
```

### Configuration Validation

```bash
# Validate configuration
./probe validate-config --config config.json

# Test platform compatibility
./probe test-platform --config config.json

# Check network configuration
./probe check-network

# Validate targets
./probe validate-targets --config config.json
```

## Log Analysis

### Common Log Patterns

#### Permission Errors
```
ERROR: failed to create ICMP socket: operation not permitted
WARN: Windows platform requires administrator privileges
ERROR: insufficient permissions for raw socket access
```

**Solutions**:
- Run with appropriate privileges
- Add required capabilities
- Check platform-specific requirements

#### Network Errors
```
ERROR: failed to bind to interface eth0: address already in use
WARN: network interface eth0 not found, using default
ERROR: DNS resolution failed for target example.com
```

**Solutions**:
- Check for port conflicts
- Verify interface exists
- Use IP addresses instead of hostnames

#### Configuration Errors
```
ERROR: invalid JSON syntax in configuration file
WARN: missing required field 'network', using defaults
ERROR: invalid value for 'buffer_size': 100 (minimum 1024)
```

**Solutions**:
- Validate JSON syntax
- Provide all required fields
- Use valid value ranges

### Log Levels and Interpretation

#### Error Level
- Critical issues that prevent operation
- Requires immediate attention
- Examples: Permission denied, invalid configuration

#### Warning Level
- Issues that don't prevent operation but may affect performance
- Configuration warnings
- Platform limitations

#### Info Level
- Normal operation information
- Status updates
- Configuration loading

#### Debug Level
- Detailed information for troubleshooting
- Step-by-step execution details
- Performance metrics

### Log Collection

#### Linux/macOS
```bash
# View logs with journalctl
journalctl -u mesh-probe -f

# Follow log file
tail -f /var/log/mesh-probe/probe.log

# Search for errors
journalctl -u mesh-probe | grep ERROR

# Export logs for analysis
journalctl -u mesh-probe --since "2023-01-01" > probe_logs.txt
```

#### Windows
```cmd
REM View Event Viewer
eventvwr.msc

REM PowerShell log retrieval
Get-EventLog -LogName Application -Source "Mesh Probe" | Where-Object {$_.EntryType -eq "Error"}

REM Export logs
Get-EventLog -LogName Application -Source "Mesh Probe" | Export-Csv probe_logs.csv
```

#### Container Logs
```bash
# Docker logs
docker logs -f mesh-probe

# Kubernetes logs
kubectl logs -f deployment/mesh-probe

# Export container logs
docker logs mesh-probe > probe_logs.txt
kubectl logs deployment/mesh-probe > k8s_logs.txt
```

### Log Analysis Tools

#### Built-in Analysis
```bash
# Use probe's log analysis
./probe analyze-logs --file probe.log --summary

# Extract errors and warnings
./probe analyze-logs --file probe.log --errors --warnings

# Performance analysis
./probe analyze-logs --file probe.log --performance
```

#### External Tools
```bash
# Use grep for pattern matching
grep ERROR probe.log

# Use awk for structured analysis
awk '/ERROR/ {count++} END {print "Total errors:", count}' probe.log

# Use jq for JSON log parsing (if using JSON logs)
jq '.level' probe.log | sort | uniq -c
```

---

For additional help, see:
- [Cross-Platform Configuration Guide](cross-platform-configuration.md)
- [Quick Reference Guide](quick-reference.md)
- [Main README](../README.md)