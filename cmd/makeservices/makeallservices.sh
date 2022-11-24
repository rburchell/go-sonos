#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

declare -a arr=("VirtualLineIn" "GroupRenderingControl" "Queue" "AVTransport" "ConnectionManager" "RenderingControl")

for i in "${arr[@]}"
do
    go run ${DIR}/makeservice.go ${i} /MediaRenderer/${i}/Control /MediaRenderer/${i}/Event ${DIR}/xml/${i}1.xml >${i}.go
    goimports -w ${i}.go

    mkdir -p ${DIR}/../../services/${i}/
    mv ${i}.go ${DIR}/../../services/${i}/
done

declare -a arr=("ContentDirectory" "ConnectionManager")

for i in "${arr[@]}"
do
    go run ${DIR}/makeservice.go ${i} /MediaServer/${i}/Control /MediaServer/${i}/Event ${DIR}/xml/${i}1.xml >${i}.go
    goimports -w ${i}.go

    mkdir -p ${DIR}/../../services/${i}/
    mv ${i}.go ${DIR}/../../services/${i}/
done

declare -a arr=("AudioIn" "AlarmClock" "MusicServices" "DeviceProperties" "SystemProperties" "ZoneGroupTopology" "GroupManagement" "QPlay")

for i in "${arr[@]}"
do
    go run ${DIR}/makeservice.go ${i} /${i}/Control /${i}/Event ${DIR}/xml/${i}1.xml >${i}.go
    goimports -w ${i}.go

    mkdir -p ${DIR}/../../services/${i}/
    mv ${i}.go ${DIR}/../../services/${i}/
done
