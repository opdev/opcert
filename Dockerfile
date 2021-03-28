FROM python:buster

#jupyter-notebook
RUN pip install jupyterlab; \
# offline-cataloger
wget https://github.com/kevinrizza/offline-cataloger/releases/download/0.0.1/offline-cataloger; \
chmod +x offline-cataloger; \
mv offline-cataloger /usr/local/bin/offline-cataloger; \
# operator-sdk
wget https://github.com/operator-framework/operator-sdk/releases/download/v1.5.0/operator-sdk_linux_amd64; \
chmod +x operator-sdk_linux_amd64; \
mv operator-sdk_linux_amd64 /usr/local/bin/operator-sdk; \
# oc
wget https://mirror.openshift.com/pub/openshift-v4/clients/ocp/4.7.0/openshift-client-linux-4.7.0.tar.gz; \
tar -C /usr/local/bin/ -xzf openshift-client-linux-4.7.0.tar.gz; \
# jq and yq
apt update; apt install jq -y; \
pip3 install yq; \
# operator_courier
pip3 install operator-courier; \
# golang
wget https://golang.org/dl/go1.16.2.linux-amd64.tar.gz; \
tar -C /usr/local -xzf go1.16.2.linux-amd64.tar.gz; \
echo "export PATH=/usr/local/go/bin:$PATH" >> .bashrc; \
# opm
 wget https://github.com/operator-framework/operator-registry/releases/download/v1.16.1/linux-amd64-opm; \
chmod +x linux-amd64-opm; \
mv linux-amd64-opm /usr/local/bin/opm; \
# skopeo
echo 'deb https://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable/Debian_10/ /' > /etc/apt/sources.list.d/devel:kubic:libcontainers:stable.list; \
curl -L https://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable/Debian_10/Release.key | apt-key add - ; \
apt-get update; \
apt-get -y install skopeo; \
# umoci
wget https://github.com/opencontainers/umoci/releases/download/v0.4.6/umoci.amd64; \
chmod +x umoci.amd64; \
mv umoci.amd64 /usr/local/bin/umoci;

RUN useradd -ms /bin/bash jupyter
USER jupyter
WORKDIR /home/jupyter

COPY init.sh /home/jupyter/

CMD ["/bin/bash", "./init.sh"]