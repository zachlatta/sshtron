#!/bin/sh

cd "$SNAP_DATA"

[ -f id_rsa ] || $SNAP/usr/bin/ssh-keygen -t rsa -N '' -f id_rsa

sshtron > "$SNAP_DATA/sshtron.log" 2>&1
