

```yaml

apiVersion: github.com/seanly
kind: Checksumer
metadata:
    name: myTransformer

spec:
    checksum:
      - key: checksum/config
        target:
          kind: ConfigMap
          name: xxx
          
    # Where the above keys will be inserted in the resulting transformed resources
    selectors:
      - target:
          kind: Deployment
          name: xxx
        filedSpec:
          path: metadata/annotations
          create: true
      - target:
          kind: Deployment
          name: xxx
        fieldSpec:
          path: spec/template/metadata/annotations
          create: true

```
