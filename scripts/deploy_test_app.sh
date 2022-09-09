
cat << EOF > deployment.yaml
---
apiVersion: v1
kind: Namespace
metadata:
  name: ci-dev

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tls-test-app
  namespace: ci-dev
  labels:
    app: tls-test-app
spec:
  selector:
    matchLabels:
      app: tls-test-app
  replicas: 1
  template:
    metadata:
      labels:
        app: tls-test-app
    spec:
      initContainers:
        - name: certificate-init-container
          image: twdps/certificate-init-container:dev.${CIRCLE_SHA1:0:8}
          imagePullPolicy: Always
          env:
            - name: NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
          args:
            - "--common-name=init-container-test"
            - "--organization='Thoughtworks, Inc.'"
            - "--organizational-unit=EMPC"
            - "--country=USA"
            - "--province=Illinois"
            - "--locality=Chicago"
            - "--street-address='200 E Randolph St 25th Floor'"
            - "--postal-code=60601"
            - "--service-names=init-container-test"
            - "--namespace=\$(NAMESPACE)"
            - "--cert-dir=/etc/tls"            

          volumeMounts:
            - name: tls
              mountPath: /etc/tls

      containers:
        - name: tls-test-app
          image: gcr.io/hightowerlabs/tls-app:1.0.0
          imagePullPolicy: Always
          args:
            - "-tls-cert=/etc/tls/tls.crt"
            - "-tls-key=/etc/tls/tls.key"
          ports:
            - containerPort: 443 
          resources:
            limits:
              memory: "50Mi"
              cpu: "100m"
          volumeMounts:
            - name: tls
              mountPath: /etc/tls
      volumes:
        - name: tls
          emptyDir: {}
EOF

kubectl apply -f deployment.yaml