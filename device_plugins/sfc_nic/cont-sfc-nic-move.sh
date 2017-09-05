#!/bin/bash -x
containerName=${1}
NIC=${2}
rm -f /var/run/netns/$containerName
#containerID=`docker ps | grep $containerName | awk {'print $1'}`
containerID=`docker -H unix:///gopath/run/docker.sock ps | grep $containerName | awk {'print $1'}`
echo $containerID
while [ -z $containerID ]; do
	echo "sleep"
	containerID=`docker ps | grep $containerName | awk {'print $1'}`
	  sleep 0.1
done
#PID=`docker inspect --format '{{ .State.Pid }}' $containerID`
PID=`docker -H unix:///gopath/run/docker.sock inspect --format '{{ .State.Pid }}' $containerID`
#ln -s /proc/$PID/ns/net /var/run/netns/$containerName
#ip link set dev $NIC netns $containerName
ssh -o StrictHostKeyChecking=no 127.0.0.1 ln -s /proc/$PID/ns/net /var/run/netns/$containerName
ssh -o StrictHostKeyChecking=no 127.0.0.1 ip link set dev $NIC netns $containerName

