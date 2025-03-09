# tesla-ble-proxy
Tesla BLE Rest API (proxy)

## fix go-ble deadlock

use `vincent-tr/ble_BleConnectFix`: in go.mod

replace github.com/go-ble/ble => github.com/vincent-tr/ble_BleConnectFix master

## Build for raspberry pi

```bash
docker run --rm --privileged multiarch/qemu-user-static --reset -p yes --credential yes

docker run --platform linux/arm64 --rm -ti -v $(pwd):/mnt arm64v8/alpine:3.21.3 /bin/sh

apk update
apk add go
cd /mnt
go build -o tesla-ble-proxy .
```

## Deploy on raspberry pi (quick and dirty)

On builder:
```bash
scp tesla-ble-proxy root@<target>:/usr/bin/tesla-ble-proxy
scp alpine/tesla-ble-proxy.initd root@<target>:/etc/init.d/tesla-ble-proxy
```

On target:
```bash
rc-update add tesla-ble-proxy
rc-service tesla-ble-proxy start
lbu include /usr/bin/tesla-ble-proxy
lbu include /etc/init.d/tesla-ble-proxy
lbu commit -d
```

## Test client 

```bash
TESLA_VIN=<xxx> go run test-cli/main.go
```

## Install tesla-control/public key

On builder:
```bash
docker run --rm --privileged multiarch/qemu-user-static --reset -p yes --credential yes

docker run --platform linux/arm64 --rm -ti -v $(pwd):/mnt arm64v8/alpine:3.21.3 /bin/sh

apk update
apk add go git
cd /tmp
git clone https://github.com/teslamotors/vehicle-command/
cd vehicle-command/cmd/tesla-control
go get .
go build -o /mnt/tesla-control .
```

On builder:
```bash
scp tesla-control root@<target>:/usr/bin/tesla-control
scp ../tesla-auth/extra/vehicle-public-key.pem root@<target>:/root/
scp ../tesla-auth/vehicle-private-key.pem root@<target>:/root/ # only for test
```

On target:
```bash
export TESLA_KEY_NAME=$(whoami)
export TESLA_TOKEN_NAME=$(whoami)
export TESLA_KEYRING_TYPE=file
export TESLA_VIN=XP7YGCEL9PB095641
export TESLA_CACHE_FILE=tesla-cache.json
export TESLA_KEY_FILE=vehicle-private-key.pem # only for test
# Install key on car
tesla-control -ble add-key-request vehicle-public-key.pem owner cloud_key
# Test ble usage
tesla-control -ble unlock
```
