# Checksumer (Kustomize plugin)

# Usage

```bash
cd example
kustomize build . --load-restrictor LoadRestrictionsNone --enable-alpha-plugins
```

# Syntax

```yaml
apiVersion: github.com/seanly
kind: Checksumer
metadata:
  name: notImportantHere
  annotations:
    config.kubernetes.io/function: |
      container:
        image: registry.cn-hangzhou.aliyuncs.com/k8ops/kustomize-functions:checksumer-v0.1.0

spec:
  checksum:
    # Generate a checksum based on those objects 
    - key: checksum/config
      target:
        kind: ConfigMap
        name: myapp
  selectors:
    # Append the generated checksum to those objects
    - target:
        kind: Deployment
      fieldSpec:
        path: spec/template/metadata/annotations
        create: true

```
