FROM docker.target.com/tap/alpine-certs
COPY /bin/gometricsroot /gometricsroot
RUN chmod u+x /gometricsroot
ENTRYPOINT ["/gometricsroot"]
