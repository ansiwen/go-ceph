#!/bin/bash -xe

docker network create ceph-net || true

docker run --device /dev/fuse --cap-add SYS_ADMIN \
  --security-opt apparmor:unconfined --rm -d \
  --name cluster_a --net ceph-net -v cluster_a_data:/tmp/ceph \
  go-ceph-ci:octopus --test-run=NONE --pause

docker run --device /dev/fuse --cap-add SYS_ADMIN \
  --security-opt apparmor:unconfined --rm -d \
  --name cluster_b --net ceph-net -v cluster_b_data:/tmp/ceph \
  go-ceph-ci:octopus --test-run=NONE --pause


docker run --device /dev/fuse --cap-add SYS_ADMIN \
  --security-opt apparmor:unconfined --rm -ti --net ceph-net \
  -v /Users/svanders/github/go-ceph:/go/src/github.com/ceph/go-ceph \
  -v cluster_a_data:/cluster_a_data -v cluster_b_data:/cluster_b_data \
  --entrypoint /bin/bash go-ceph-ci:octopus -c "./mirror_config.sh ; bash -i" 

docker kill cluster_a cluster_b
