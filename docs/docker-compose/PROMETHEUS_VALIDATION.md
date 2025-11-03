# Prometheus Configuration Validation Summary

## ✅ Fixed Syntax Errors

### prometheus-dev.yml
- **FIXED**: Moved `external_labels` from root level to under `global:` section
- **VALIDATED**: All scrape_configs, alerting, and rule_files sections are properly formatted

### prometheus-prod.yml  
- **FIXED**: Moved `external_labels` from root level to under `global:` section
- **FIXED**: Removed invalid `storage` section (storage config should be command line args)
- **VALIDATED**: All scrape_configs, alerting, and rule_files sections are properly formatted

## Configuration Structure

Both files now follow proper Prometheus configuration format:

```
global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    cluster: 'mesh-probe-<env>'
    region: '<env>'

rule_files:
  - "/etc/prometheus/alerts.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093

scrape_configs:
  - job_name: 'service-name'
    static_configs:
      - targets: ['service:port']
    metrics_path: '/metrics'
    scrape_interval: 30s
```

## Environment-Specific Labels

- **Development**: `environment: 'development'`, `cluster: 'mesh-probe-dev'`
- **Production**: `cluster: 'mesh-probe-prod'`, `region: 'prod'`

## Next Steps

Both configurations are now ready for deployment:
1. Copy `prometheus-dev.yml` to development environment
2. Copy `prometheus-prod.yml` to production environment
3. Configure storage retention via command line flags:
   ```bash
   --storage.tsdb.retention.time=30d
   --storage.tsdb.retention.size=10GB