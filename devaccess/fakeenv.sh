#!/bin/sh

cat << EOF
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
TERM=xterm
HOSTNAME=devaccess
NSS_SDB_USE_CACHE=no
PCIDEVICE_OPENSHIFT_IO_DU_MH_INFO={"0000:19:02.6":{"generic":{"deviceID":"0000:19:02.6"},"vfio":{"dev-mount":"/dev/vfio/204","mount":"/dev/vfio/vfio"}}}
PCIDEVICE_OPENSHIFT_IO_DU_MH=0000:19:02.6
KUBERNETES_SERVICE_HOST=172.30.0.1
KUBERNETES_SERVICE_PORT=443
KUBERNETES_SERVICE_PORT_HTTPS=443
KUBERNETES_PORT=tcp://172.30.0.1:443
KUBERNETES_PORT_443_TCP=tcp://172.30.0.1:443
KUBERNETES_PORT_443_TCP_PROTO=tcp
KUBERNETES_PORT_443_TCP_PORT=443
KUBERNETES_PORT_443_TCP_ADDR=172.30.0.1
container=oci
HOME=/root
EOF
