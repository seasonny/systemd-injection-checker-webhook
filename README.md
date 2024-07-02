# Systemd Property Injection Checker Webhook

## Overview

This repository implements a Kubernetes admission webhook to mitigate the CVE-2024-3154 vulnerability in CRI-O. The vulnerability allows arbitrary systemd property injection via Pod annotations, which could lead to unauthorized actions on the host system.

## Vulnerability Details

- **CVE**: CVE-2024-3154
- **Affected Component**: CRI-O
- **Issue**: Arbitrary systemd property can be injected via a Pod annotation with the prefix `org.systemd.property.`
- **Impact**: Users who can create pods with arbitrary annotations may perform unauthorized actions on the host system

For more information, see:
- [Red Hat Security Advisory](https://access.redhat.com/security/cve/CVE-2024-3154)
- [GitHub Advisory](https://github.com/cri-o/cri-o/security/advisories/GHSA-2cgq-h8xw-2v5j)

## Mitigate Solution

This webhook intercepts Pod creation and update requests, checking for annotations with the `org.systemd.property.` prefix. If such annotations are found, the webhook denies the request, preventing potential exploitation of the vulnerability.

## Testing in Your Local Golang Environment
```
go build -o webhook-server 
./webhook-server

curl -X POST -H "Content-Type: application/json" --data @test-systemd-injection.json http://localhost:8080/validate
wget --no-check-certificate --header="Content-Type: application/json" --post-file=test-systemd-injection.json http://localhost:8080/validate -O - -q
```

## Prerequisites

- OpenShift 4 cluster
- `oc` or `kubectl` CLI tool
- Podman / Docker 

## Deployment

1. Build the Docker image:

```
podman build -t $your-registry/$repo/systemd-checker-webhook:1.1 .
```
2. Testing in Your Local Container Environment
```
podman run -p 8080:8080 $your-registry/$repo/systemd-checker-webhook:1.1
curl -X POST -H "Content-Type: application/json" --data @test-systemd-injection.json http://localhost:8080/validate
wget --no-check-certificate --header="Content-Type: application/json" --post-file=test-systemd-injection.json http://localhost:8080/validate -O - -q
```

3. Push the image to your registry:
```
podman push $your-registry/$repo/systemd-injection-checker-webhook:1.1
```

4. Deploy the webhook:

```
oc apply -f systemd-injection-checker-webhook-service.yaml
oc apply -f systemd-injection-checker-webhook-deployment.yaml
oc apply -f systemd-injection-checker-webhook.yaml
```

5. Deploy a pod with systemd annotation for test
```
oc apply -f systemd-injection-pod.yaml
```

6. If everything is good, you will see something like this
```
Error from server: error when creating "systemd-injection-pod.yaml": admission webhook "systemd-injection-checker.example.redhat.com" denied the request: Annotation with prefix 'org.systemd.property.' is not allowed: org.systemd.property.SuccessAction
```

## Limitations & Disclaimers

* This webhook is a mitigation strategy and does not fix the underlying issue in CRI-O.
* It may impact legitimate use cases that require systemd property annotations.
* Regular updates to CRI-O and the container runtime are still necessary.
* This webhook is provided as-is without any guarantees. Always test thoroughly in a non-production environment before deploying to production.
