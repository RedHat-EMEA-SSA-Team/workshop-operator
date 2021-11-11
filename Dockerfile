# Build the manager binary
FROM golang:1.16 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

#TEMPORARY FIX - Conflict between kube-openapi & go-openapi in the ArgoCD
# To remove with the module ArgoCD will be ready for Kubernetes 1.22
RUN sed -i 's/github\.com\/go\-openapi\/spec/k8s\.io\/kube\-openapi\/pkg\/validation\/spec/g' /go/pkg/mod/github.com/argoproj/argo-cd/v2@v2.1.6/pkg/apis/application/v1alpha1/openapi_generated.go

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY common/ common/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM registry.access.redhat.com/ubi8/ubi-minimal:latest
WORKDIR /
COPY --from=builder /workspace/manager .

ENTRYPOINT ["/manager"]
