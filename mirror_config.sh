#!/bin/bash -xe

#/entrypoint.sh --test-run=NONE
export CONF_A=/cluster_a_data/ceph.conf
export CONF_B=/cluster_b_data/ceph.conf

while ! [[ -f /cluster_a_data/.ready && -f /cluster_b_data/.ready ]] ; do
  sleep 1
done
ceph -c $CONF_A osd pool create rbd 8
ceph -c $CONF_B osd pool create rbd 8
rbd -c $CONF_A pool init
rbd -c $CONF_B pool init
rbd -c $CONF_A mirror pool enable rbd image
rbd -c $CONF_B mirror pool enable rbd image
rbd -c $CONF_A mirror pool peer bootstrap create --site-name cluster_a rbd > token
rbd -c $CONF_B mirror pool peer bootstrap import --site-name cluster_b rbd token
rbd -c $CONF_A create alice --size 1G
rbd -c $CONF_A mirror image enable alice snapshot
#rbd -c $CONF_A mirror image resync
yum install -y rbd-fuse
while ! rbd -c /cluster_a_data/ceph.conf mirror image status alice \
  | grep -q "state: \+up+replaying" ; do
  sleep 1
done
rbd -c $CONF_A mirror pool info rbd --all
rbd -c $CONF_B mirror pool info rbd --all
rbd -c $CONF_A mirror image status alice
rbd -c $CONF_B mirror image status alice
mkdir /tmp/ceph /mnt/images_a /mnt/images_b
rbd-fuse -c $CONF_A /mnt/images_a
rbd-fuse -c $CONF_B /mnt/images_b
echo "Mexican Funeral" | dd conv=notrunc bs=1 of=/mnt/images_a/alice
rbd -c $CONF_A mirror image snapshot alice
while ! dd bs=1 count=16 if=/mnt/images_b/alice | grep -q "Mexican Funeral" ; do
  sleep 1
done
echo "Mirroring functional"
