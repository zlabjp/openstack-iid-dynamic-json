FROM ubuntu:bionic

COPY out/bin/openstack-iid-dynamic-json /openstack-iid-dynamic-json

ENTRYPOINT ["/openstack-iid-dynamic-json"]
