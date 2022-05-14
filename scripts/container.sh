#!/usr/bin/env bash
# Starts development docker container

docker --context "${DOCKER_CONTEXT?:is missing}" run --rm -it \
  -v /dev/bus:/dev/bus:ro \
  -v /dev/serial:/dev/serial:ro \
  -v "${SCAN_DIR?:is missing}:/work/scans" \
  -p 8090:8090 \
  --cap-add SYS_PTRACE \
  --device-cgroup-rule "c ${DEVICE_MAJOR?:is missing}:* rwm" \
  "${DOCKER_IMAGE?:is missing}" \
  scandoc -dir /work/scans
