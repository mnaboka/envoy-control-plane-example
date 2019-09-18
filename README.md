# envoy-control-plane-example

Example of the control plane server for envoy.

This is an example implementation of envoy control plane written in golang.
Envoy is extremely powerful, but learning curve might be steep.

In this example we are dynamically updating Clusters, Endpoints and Routes.
More details [here](https://www.envoyproxy.io/docs/envoy/latest/configuration/overview/v2_overview)

docker-compose contains the following containers:
  - `control-plane` - implementation of xDS server. It listens for xDS DiscoveryRequests on port `:5678`
    and exposes REST API on port `8000` to dynamically add/remove clusters and endpoints(upstreams).
    CLI script contains examples how to interact with REST API [cli.sh](https://github.com/mnaboka/envoy-control-plane-example/blob/master/cli.sh)

  - `envoy` - is an official image of envoy v1.11.1, it exposes admin interface on port `:9901` and HTTP1 listener on `:3000`
  - `test-server-dev` and `test-server-prod` simple HTTP servers listens on `:8080` and returns info about server (hostname, ip address etc.)

### How to use CLI
```bash
Usage:
./cli.sh cluster add <name> <url_prefix>
./cli.sh cluster remove <name>
./cli.sh endpoint add <cluster_name> <ip_address> <port>
./cli.sh endpoint remove <cluster_name> <ip_address> <port>
./cli.sh commit
```

Examples:

To add a new cluster with a URL prefix `/api/v1` and 2 upstream hosts. When changes are made, they must be committed with
`commit` subcommand.

```bash
cli.sh cluster add backend-cluster /api/v1
cli.sh endpoint add backend-cluster 10.10.0.1 8080
cli.sh endpoint add backend-cluster 10.10.0.2 8080
cli.sh commit
```

To verify changes, you can navigate to envoy admin interface:
  - [config_dump](http://127.0.0.1:9901/config_dump)
  - [clusters](http://127.0.0.1:9901/clusters)



### How to use this repo
  - start with cloning this repo on a local machine

```bash
git clone git@github.com:mnaboka/envoy-control-plane-example.git
```

  - start up docker-compose and scale the `test-server-dev` and `test-server-prod` to 3 instances

```bash
docker-compose up -d --scale test-server-prod=3 --scale test-server-dev=3 --no-recreate
```

  - run `docker-compose ps` to see all the running instances, we should have 8 containers running
  
```bash
$ docker-compose ps
                     Name                                   Command               State                             Ports
---------------------------------------------------------------------------------------------------------------------------------------------------
envoy-control-plane-example_control-plane_1      /go/bin/control-plane            Up      0.0.0.0:5678->5678/tcp, 0.0.0.0:8000->8000/tcp, 8080/tcp
envoy-control-plane-example_envoy_1              /docker-entrypoint.sh /usr ...   Up      10000/tcp, 0.0.0.0:3000->3000/tcp, 0.0.0.0:9901->9901/tcp
envoy-control-plane-example_test-server-dev_1    /go/bin/server                   Up      5678/tcp, 8000/tcp, 8080/tcp
envoy-control-plane-example_test-server-dev_2    /go/bin/server                   Up      5678/tcp, 8000/tcp, 8080/tcp
envoy-control-plane-example_test-server-dev_3    /go/bin/server                   Up      5678/tcp, 8000/tcp, 8080/tcp
envoy-control-plane-example_test-server-prod_1   /go/bin/server                   Up      5678/tcp, 8000/tcp, 8080/tcp
envoy-control-plane-example_test-server-prod_2   /go/bin/server                   Up      5678/tcp, 8000/tcp, 8080/tcp
envoy-control-plane-example_test-server-prod_3   /go/bin/server                   Up      5678/tcp, 8000/tcp, 8080/tcp

```
  
  - find ip addresses for our `test-server-dev` instances
  
```bash
$ docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' envoy-control-plane-example_test-server-dev_1 envoy-control-plane-example_test-server-dev_2 envoy-control-plane-example_test-server-dev_3
192.168.240.7
192.168.240.4
192.168.240.8
```

  - create a new envoy cluster called "dev" with provided `cli.sh`
```bash
$ ./cli.sh cluster add dev /dev

HTTP/1.1 200 OK
Date: Wed, 18 Sep 2019 00:46:09 GMT
Content-Length: 0

$ ./cli.sh endpoint add dev 192.168.240.7 8080

HTTP/1.1 200 OK
Date: Wed, 18 Sep 2019 00:46:27 GMT
Content-Length: 0

$ ./cli.sh endpoint add dev 192.168.240.4 8080

HTTP/1.1 200 OK
Date: Wed, 18 Sep 2019 00:46:32 GMT
Content-Length: 0

$ ./cli.sh endpoint add dev 192.168.240.8 8080

HTTP/1.1 200 OK
Date: Wed, 18 Sep 2019 00:46:38 GMT
Content-Length: 0

$ ./cli.sh commit

HTTP/1.1 200 OK
Date: Wed, 18 Sep 2019 00:47:59 GMT
Content-Length: 0
```

  - curl envoy to see if it proxies the requests
  
```bash
17:49 $ curl 127.0.0.1:3000/dev 2>/dev/null| jq '.'
{
  "env": "development",
  "hostname": "16a1355951b2",
  "ips": [
    "127.0.0.1/8",
    "192.168.240.8/20"
  ]
}
```

  - verify changes, you can navigate to envoy admin interface:
    - [config_dump](http://127.0.0.1:9901/config_dump)
    - [clusters](http://127.0.0.1:9901/clusters)
  - optionally repeat steps for test-server-prod
  - great! looks like envoy proxies to the correct upstream host, congrats!