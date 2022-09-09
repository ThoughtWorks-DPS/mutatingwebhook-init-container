FROM busybox:glibc
WORKDIR /
COPY ./sidecar-mutatingwebhook-init-container ./sidecar-mutatingwebhook-init-container
ENTRYPOINT ["./sidecar-mutatingwebhook-init-container"]