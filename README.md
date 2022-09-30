<div align="center">
	<p>
		<img alt="Thoughtworks Logo" src="https://raw.githubusercontent.com/ThoughtWorks-DPS/static/master/thoughtworks_flamingo_wave.png?sanitize=true" width=200 />
    <br />
		<img alt="DPS Title" src="https://raw.githubusercontent.com/ThoughtWorks-DPS/static/master/EMPCPlatformStarterKitsImage.png?sanitize=true" width=350/>
	</p>
  <h3>sidecar-mutatingwebhook-init-container</h3>
    <a href="https://app.circleci.com/pipelines/github/ThoughtWorks-DPS/sidecar-mutatingwebhook-init-container"><img src="https://circleci.com/gh/ThoughtWorks-DPS/sidecar-mutatingwebhook-init-container.svg?style=shield"></a> <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/license-MIT-blue.svg"></a>
</div>
<br />

Init container for managing the deployment of a MutatingWebhookConfiguration to trigger an admission-contoller based on deployments to a namespace based on matching namespace annotation.  

## Usage

Kubernetes admission controllers depend on a webhook configuration as the triggering event. Two types of webhooks are  supported, validating webhooks and mutating webhooks. Where the expected outcome from an admission controller is to modify a deployment, a mutating webhook is configured with the desired trigger parameters.  

This sidecar-mutatingwebhook-init-container defines a mutating webhook configured to watch for any pod creation events in any namespace that is annotated with a specified key and value. As a result, prior to performing the deployment the kubernetes api will call your custom api sending the contents of the deployment request. It is assumed your 'admission controller' service will modify the deployment in some way.  

The communication between the kubernetes api and your custom service must be encrtyped via mTls. When you deploy your service you must configure it with the necessary public certificate and private key. This same certificate must be included in the mutating webhook configuration so that the kubernetes api may present it to your service to establish mutual tls encryption.

As the name suggests, a typical use case is the addition of a sidecar container to all deployments within appropriately annotated namespaces.  

Inclusion with a service mesh is problemmatic since the kubernetes api operates outside the logical boundary of a mesh.  

The most common approach to managing these webhook certificates is either 1) use of a certificate injector integrated with an automated certificate management solution if available (such as [cainjector](https://cert-manager.io/docs/concepts/ca-injector/)), or 2) self-signed certificates generated during deployment.  

This init container is optimized to work in conjuction with a [certificate init container](https://github.com/ThoughtWorks-DPS/certificate-init-container), however it can be used independently so long as the required certificate is available on a (configurable) mount path location.  

The certificate-init-container follows a well known pattern of generating a CA certificate and then using this CA to sign the certificate made available to the pod. This is done in proc before the resulting certificate is written to the emptyDir shared pod mount, resulting in the signing CA no longer existing when the init container terminates. As a result, no private key or CA exists outside of the admission controllers deployment context.

When used with the certificate init container, the sidecar-mutatingwebhook init must run second. By default is will expect the required certificate to be available in the emptyDir shared pod volume mount unless otherwise specified.  

Add the sidecar-mutatingwebhook-init-container to an existing deployment:  
```yaml
				- name: sidecar-mutatingwebhook-init-container
          image: twdps/sidecar-mutatingwebhook-init-container:0.1.0
          imagePullPolicy: Always
          env:
            - name: NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
          args:
            - "--webhook-name=opa-injector.twdps.io"
            - "--object-meta-name=opa-injector-admission-controller-webhook"
            - "--service-name=opa-injector-admission-controller"
            - "--service-namespace=\$(NAMESPACE)"
            - "--service-path=/v0/data/istio/inject"
            - "--namespace-selector-key=opa-istio-injection"
            - "--namespace-selector-value=enabled"
            - "--cert-path=/etc/tls/tls.crt"
						- "--key-path=/etc/tls/tls.key"

          volumeMounts:
            - name: tls
              mountPath: /etc/tls
```

The resulting MutatingWebhookConfiguration and kubernetes Secret (with matching parameters shown):
```yaml
---
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: <object-meta-name>
webhooks:
  - name: <webhook-name>
    clientConfig:
      service:
        name: <opa-injector-admission-controller>
        namespace: <service-namespace>
        path: <service-path>
      caBundle: |-
        base64 encoded contents of <cert-path>
    rules:
      - operations: ["CREATE"]
        apiGroups: [""]
        apiVersions: ["v1"]
        resources: ["pods"]
    namespaceSelector:
      matchLabels:
        <namespace-selector-key>: <namespace-selector-value>
    failurePolicy: Fail
    admissionReviewVersions: ["v1beta1"]
    sideEffects: None

---
apiVersion: v1
kind: Secret
metadata:
  name: <serviceName>-certificate
  namespace: <serviceNamespace>
data:
  tls.crt: |-
    base64 encoded contents of <cert-path>
  tls.key: |-
		base64 encoded contents of <key-path>
    
```

Your admission controller will now be able to reference the certificate andn private key for tls either either on emptyDir pod shared mount or you may reference the kubernetes Secret.  


Where the init container is run again using the same object-meta-name, the existing webhook configuration and secret will be updated.  

_Note. When deleting the admission controller or changing the object-meta-name, the prior webhook configuration and secret (if created) will remain._