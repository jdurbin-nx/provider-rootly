# The upstream Terraform provider is built from source at the pinned tag, with
# the patches in patches/ applied, instead of taken from its release. Each
# patch carries its own test, run here, so a patch that no longer does what it
# says fails the image build. Drop a patch once upstream ships the fix.
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS tfprovider
RUN apk --no-cache add git

ARG TARGETOS
ARG TARGETARCH
ARG TERRAFORM_PROVIDER_REPO
ARG TERRAFORM_PROVIDER_VERSION
ARG TERRAFORM_NATIVE_PROVIDER_BINARY

WORKDIR /src
RUN git clone -c advice.detachedHead=false --depth 1 --branch "v${TERRAFORM_PROVIDER_VERSION}" "${TERRAFORM_PROVIDER_REPO}" .
COPY patches/ /patches/
RUN git apply --verbose /patches/*.patch

RUN go test ./client/ -run 'TestMarshalWorkflowTask'

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath \
      -ldflags "-s -w -X github.com/rootlyhq/terraform-provider-rootly/v5/meta.version=${TERRAFORM_PROVIDER_VERSION}" \
      -o "/out/${TERRAFORM_NATIVE_PROVIDER_BINARY}" .

FROM alpine:3.23.4
RUN apk --no-cache add ca-certificates bash

ARG TARGETOS
ARG TARGETARCH

ADD "bin/${TARGETOS}_${TARGETARCH}/provider" /usr/local/bin/provider

ENV USER_ID=65532

# Setup Terraform environment

## Provider-dependent configuration
ARG TERRAFORM_VERSION
ARG TERRAFORM_PROVIDER_SOURCE
ARG TERRAFORM_PROVIDER_VERSION
ARG TERRAFORM_NATIVE_PROVIDER_BINARY

## End of - Provider-dependent configuration

ENV PLUGIN_DIR=/terraform/provider-mirror/registry.terraform.io/${TERRAFORM_PROVIDER_SOURCE}/${TERRAFORM_PROVIDER_VERSION}/${TARGETOS}_${TARGETARCH}
ENV TF_CLI_CONFIG_FILE=/terraform/.terraformrc
ENV TF_FORK=0

RUN mkdir -p ${PLUGIN_DIR}

ADD https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_${TARGETOS}_${TARGETARCH}.zip /tmp
COPY --from=tfprovider /out/${TERRAFORM_NATIVE_PROVIDER_BINARY} ${PLUGIN_DIR}/
ADD terraformrc.hcl ${TF_CLI_CONFIG_FILE}

RUN unzip /tmp/terraform_${TERRAFORM_VERSION}_${TARGETOS}_${TARGETARCH}.zip -d /usr/local/bin \
  && chmod +x /usr/local/bin/terraform \
  && rm /tmp/terraform_${TERRAFORM_VERSION}_${TARGETOS}_${TARGETARCH}.zip \
  && chmod +x ${PLUGIN_DIR}/* \
  && chown -R ${USER_ID}:${USER_ID} /terraform
# End of - Setup Terraform environment

# Provider controller needs these environment variable at runtime
ENV TERRAFORM_VERSION=${TERRAFORM_VERSION}
ENV TERRAFORM_PROVIDER_SOURCE=${TERRAFORM_PROVIDER_SOURCE}
ENV TERRAFORM_PROVIDER_VERSION=${TERRAFORM_PROVIDER_VERSION}
ENV TERRAFORM_NATIVE_PROVIDER_PATH=${PLUGIN_DIR}/${TERRAFORM_NATIVE_PROVIDER_BINARY}

USER ${USER_ID}
EXPOSE 8080

ENTRYPOINT ["provider"]
