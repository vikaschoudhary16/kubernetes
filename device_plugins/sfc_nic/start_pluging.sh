#!/bin/bash
/usr/bin/sfc-nic-plugin $ONLOAD_VERSION $REG_EXP_SFC $SOCKET_NAME $RESOURCE_NAME $K8S_API $NODE_LABEL_ONLOAD_VERSION -logtostderr=true
