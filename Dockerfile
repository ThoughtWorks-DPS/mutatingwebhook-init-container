FROM busybox:glibc
WORKDIR /
COPY ./mutatingwebhook-init-container ./mutatingwebhook-init-container
ENTRYPOINT ["./mutatingwebhook-init-container"]