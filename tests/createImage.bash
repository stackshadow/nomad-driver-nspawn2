#!/usr/bin/env bash
set -euo pipefail

IMG="./debian.raw"
MNT="./debian-rootfs"


if [ ! -f "${IMG}" ]; then
  truncate -s 2G "${IMG}"
fi

if ! blkid -o value -s TYPE "${IMG}" 2>/dev/null | grep -qx 'ext4'; then
  mkfs.ext4 -F "${IMG}"
fi

if [ ! -d "${MNT}" ]; then
  sudo mkdir -p "${MNT}"
fi

if ! mountpoint -q "${MNT}"; then
  sudo mount -o loop "${IMG}" "${MNT}"
  echo "Mounted on ${MNT}"
  mounted=1
else
  echo "Already mounted on ${MNT}"
  mounted=1
fi

if [ ! -f ${MNT}/bin/systemctl ]; then
  debootstrap \
  --include=systemd-container,systemd,dbus \
  --arch arm64 \
  bookworm \
  ${MNT} \
  http://deb.debian.org/debian/
fi

# systemctl enable --now systemd-networkd

if [ "${mounted}" -eq 1 ]; then
  echo "Unmount ${MNT}" 
  sudo umount "${MNT}"
  rmdir ${MNT}
else
  echo "Not unmount ${MNT}" 
fi
