FROM quay.io/podman/stable:latest

RUN yum install -y openscap-utils; \
yum install -y buildah wget scap-security-guide; \
yum install -y skopeo

WORKDIR /scorecard/certified/

COPY linux/opcert /scorecard/certified/opcert

# CMD [ "/scorecard/certified/opcert", "has_labels", "centos:latest" ]