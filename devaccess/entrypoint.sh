#!/bin/bash
set -ex

# do we have a `/dev/vfio` directory accessible at all?
if [ ! -d /dev/vfio ]; then
	exit 1
fi
# basic sanity check: is the control device available?
if [ ! -r /dev/vfio/vfio ] || [ ! -w /dev/vfio/vfio ]; then
	exit 2
fi

# ok, now we can check the real device access.
# 1. find all the <vfid>s allocated to the container.
#  Provided by the sriov device plugin, this will be a comma separated PCI address list
#  Starting OCP >= 4.13 the SRIOV device plugin is also exporting device info as JSON snippet.
#  We need to avoid that because it confuses our logic.
PCIDEVS=$( env | grep PCIDEVICE | grep -v '_INFO=' | cut -d\= -f2 )
# hold on just a sec: did we got any allocated devices at all?
if [ -z "${PCIDEVS}" ]; then
	exit 4
fi

# 2. figure out all /dev/vfio/<vfid> allocated devices
for PCIDEV in $( echo $PCIDEVS | tr ',' '\n' ); do
	# learn the vfio ID, which is the iommu_group (see https://docs.kernel.org/driver-api/vfio.html)
	# we can learn this value from the PCI device data exposed by the kernel on sysfs
	VFIOID=$( basename $( readlink /sys/bus/pci/devices/$PCIDEV/iommu_group ) );

	# 3. crash hard if any of the allocated devices are not are accessible within the container (R_ACCESS+W_ACCESS)
	if [ ! -r /dev/vfio/$VFIOID ] || [ ! -w /dev/vfio/$VFIOID ]; then
		exit 8
	fi
done

# 4. sleep forever otherwise
sleep inf
