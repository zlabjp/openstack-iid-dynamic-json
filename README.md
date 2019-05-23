# openstack-iid-dynamic-json

The codes included in this repository are PoC codes for OpenStack IID using Dynamic JSON. We don't assume to use in production.

## Build

```
$ make build
```

## Build Docker image

```
$ make build-linux
$ docker build -t $(docker_repo_base)/openstack-iid-dynamic-json:$(tag) -f Dockerfile build
```

## Run

Use binary

```
$ ./openstack-iid-dynamic-json
```

Use Docker image

```
$  docker run \
   -v $PWD/data:/data \
   $(docker_repo_base)/openstack-iid-dynamic-json:$(tag) \
   --dataDir=/data \
```
## Options

| Key | Description | Default |
|:----|:------------|:--------|
| port | Port to listen to | 8080 |
| dataDir | Directory name for saving data | `.` |
| keyPath | Name of a key-pair file  | `keys.json` |
| keyRotationPeriod | Period for key-pair rotation | 720h |
| interval | Check key expiration with given interval | 5s |

## Endpoints

- `iid`
```
$ curl --silent -X POST -d @pkg/endpoint/iid/fixtures/request.json http://localhost:8080/iid
{
    "data": "eyJhbGciOiJFUzM4NCIsInR5cCI6IkpXVCJ9.eyJob3N0bmFtZSI6InRlc3QuZXhhbXBsZS5vcmciLCJpbWFnZUlEIjoiaW1hZ2UtMTIzIiwiaW5zdGFuY2VJRCI6Imluc3RhbmNlLWFiYyIsIm1ldGFkYXRhIjp7ImZvbyI6ImJhciJ9LCJwcm9qZWN0SUQiOiJwcm9qZWN0LTEyMyJ9.wZC3JcRmoA72We61qfFZbZ6C8i4HgYlyv9ajxBf2Skco5XHauxcut69MghyF_GdkO7-eqZvY4bz0ZbqvQ6XIGeJwCdaaPLUl0bzhgdnqupPoBPvvFvUoZunabcbfIVNq"
}
```
- `iid_keys`
```
$ curl --silent -X POST -d @pkg/endpoint/iid/fixtures/request.json http://localhost:8080/iid_keys
{
  "keys": [
    {
      "crv": "P-384",
      "kid": "eac414d5344d26c2a519afeb54baebf6cb74c14d",
      "kty": "EC",
      "x": "We3uQypVpiZO7i1cqaIhG-ZtYqIHj3Znghw-JNRO4TpHMchyss3ezG3OvbmenTec",
      "y": "ZJibjVnc6hjjRNn6JtHrRkPuQvdGTQkkX1EDIpp3LvivB57Dxi04lpq9IjEk9ivg"
    }
  ]
}
```
