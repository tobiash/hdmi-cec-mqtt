![Docker Cloud Automated build](https://img.shields.io/docker/cloud/automated/tobiasha/hdmi-cec?style=flat-square)

# HDMI-CEC MQTT adapter

This is a simple connector to expose an HDMI-CEC adapter (e.g. PulseEight,
possibly Raspberry PI) via MQTT.

## Usage

This tool is primarily meant to be run from a docker container.

```
docker run --privileged -e MQTT_URL=tcp://localhost:1883 -v /dev/ttyACM0:/dev/ttyACM0 tobiash/hdmi-cec:latest
```

Note that privileged mode is required to be able to use the HDMI CEC device and
the device needs to be mounted into the container. Depending on the type of CEC
adapter you might need to change the device name.

### Kubernetes

A sample Kubernetes deployment file is provided in the `deploy` subdirectory, it
can also be used as a [kustomize](https://kustomize.io/) base.

### Manual build

If you want to build the tool yourself or use it without docker, check the
`Dockerfile` for build details. CGO is required!

Compile dependencies:

- `libcec-dev`
- `udev-dev`
- `p8-platform-dev`

### Raspberry PI

It should be possible to build and use this for a Raspberry PI, but I haven't
tested that yet and do not have a pre-made Docker image for it at this time. PRs
welcome!

## Configuration

Configuration is exclusively done via environment variables:

| Name | Default | Description |
| ---- | ------- | ----------- |
| MQTT_URL | | MQTT Endpoint URL |
| MQTT_USERNAME | | MQTT Username |
| MQTT_PASSWORD | | MQTT Password |
| MQTT_TOPIC | cec | MQTT Topic prefix |
| CEC_DEVICE | /dev/ttyACM0 | CEC device file |

## Topics

The prefix `/cec` is the default and can be switched by specifying the
`CEC_DEVICE` environment variable. Some operations require you to specify a
device logical address, it can be determined by reading from the `/cec/list`
topic. Usually the TV seems to have logical address `0`.

### `/cec/list`

The tool will continuously poll the list of HDMI devices and push it to this
topic. It should be treated as read-only for users. The list will also contain
the logical address of the connected devices that is required for some
operations.

### `/cec/transmit`

Publishing to this topic will transmit the given payload as arbitrary HDMI-CEC
command. See for example [cec-o-matic](http://www.cec-o-matic.com/) for details.

### `/cec/mute`

Publishing to this topic (payload is irrelevant) will toggle the mute status.

### `/cec/key/<addr>`

Publishing to this topic sends the key specified in the payload to the HDMI 
device with the logical address. See the [`cec`
library](https://github.com/chbmuc/cec/blob/master/cec.go#L34) for valid key
names.

### `/cec/power/<addr>`

Publishing the payload `on` to this topic will turn the device with the given
logical address on, publishing any other value will turn it off.

### `/cec/volume`

Publishing the payload `up` will turn the volume up, publishing any other
value will turn it down.
