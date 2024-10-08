# tesla-ble-proxy
Tesla BLE Rest API (proxy)

## Build for raspberry pi

```bash
docker run --rm --privileged multiarch/qemu-user-static --reset -p yes --credential yes

docker run --platform linux/arm64 --rm -ti -v $(pwd):/mnt arm64v8/alpine:3.20.1 /bin/sh

apk update
apk add go
cd /mnt
go build -o tesla-ble-proxy .
```

## Deploy on raspberry pi (quick and dirty)

on builder:
```bash
scp tesla-ble-proxy root@<target>:/usr/bin/tesla-ble-proxy
scp alpine/tesla-ble-proxy.initd root@<target>:/etc/init.d/tesla-ble-proxy
```

on target:
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
