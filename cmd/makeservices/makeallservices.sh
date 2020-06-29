#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

declare -a arr=("VirtualLineIn" "GroupRenderingControl" "Queue" "AVTransport" "ConnectionManager" "RenderingControl")

for i in "${arr[@]}"
do
    mkdir -p ${DIR}/../../${i}/
    rm -f ${DIR}/../../${i}/*.go
    go run ${DIR}/makeservice.go ${i} /MediaRenderer/${i}/Control /MediaRenderer/${i}/Event ${DIR}/xml/${i}1.xml >${i}.go
    go fmt ${i}.go
    mv ${i}.go ${DIR}/../../${i}/
done

declare -a arr=("ContentDirectory" "ConnectionManager")

for i in "${arr[@]}"
do
    mkdir -p ${DIR}/../../${i}/
    rm -f ${DIR}/../../${i}/*.go
    go run ${DIR}/makeservice.go ${i} /MediaServer/${i}/Control /MediaServer/${i}/Event ${DIR}/xml/${i}1.xml >${i}.go
    go fmt ${i}.go
    mv ${i}.go ${DIR}/../../${i}/
done

declare -a arr=("AlarmClock" "MusicServices" "DeviceProperties" "SystemProperties" "ZoneGroupTopology" "GroupManagement" "QPlay")

for i in "${arr[@]}"
do
    mkdir -p ${DIR}/../../${i}/
    rm -f ${DIR}/../../${i}/*.go
    go run ${DIR}/makeservice.go ${i} /${i}/Control /${i}/Event ${DIR}/xml/${i}1.xml >${i}.go
    go fmt ${i}.go
    mv ${i}.go ${DIR}/../../${i}/
done
