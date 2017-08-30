#!/bin/bash -x
containerName=${1}
NIC=${2}
rm -f /var/run/netns/$containerName
containerID=`docker ps | grep $containerName | awk {'print $1'}`
echo $containerID
while [ -z $containerID ]; do
	echo "sleep"
	containerID=`docker ps | grep $containerName | awk {'print $1'}`
	  sleep 0.1
done
PID=`docker inspect --format '{{ .State.Pid }}' $containerID`
ln -s /proc/$PID/ns/net /var/run/netns/$containerName
ip link set dev $NIC netns $containerName

