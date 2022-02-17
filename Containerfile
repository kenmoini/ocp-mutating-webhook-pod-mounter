## Builder Image Layer
FROM quay.io/polyglotsystems/golang-ubi as builder

# Set a build workspace
WORKDIR /workspace

# Copy source
COPY . /workspace/

# Download go sources
RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o server ./cmd

## Distributed Container Layer
FROM registry.access.redhat.com/ubi8/ubi-minimal:8.5

# set labels for metadata
LABEL maintainer="Ken Moini <ken@kenmoini.com>" \
  name="pod-injector" \
  description="An OpenShift mutating webhook server that implements pod injection of ConfigMaps" \
  summary="An OpenShift mutating webhook server that implements pod injection of ConfigMaps"

WORKDIR /

COPY --from=builder /workspace/server .

USER 65532:65532

EXPOSE 8443

#ENTRYPOINT ["/manager"]