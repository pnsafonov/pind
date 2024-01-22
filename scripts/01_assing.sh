#!/usr/bin/env bash

# error, only first pid
ls -1 /proc/$(pgrep -f /usr/bin/kvm)/task | xargs -I$ taskset -cp 124-127 $
# ok
ls -1 /proc/$(pgrep -f mm-ent-161-db)/task | xargs -I$ taskset -cp 32-47 $
# ok
ls -1 /proc/$(pgrep -f mm-ent-161-load)/task | xargs -I$ taskset -cp 48-63 $
