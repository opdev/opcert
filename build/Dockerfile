FROM quay.io/podman/stable:latest

RUN yum install -y openscap-utils; \
yum install -y buildah wget scap-security-guide

WORKDIR /scorecard/certified/

COPY linux/opcert /scorecard/certified/opcert

# CMD [ "/scorecard/certified/opcert", "has_labels", "centos:latest" ]