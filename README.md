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