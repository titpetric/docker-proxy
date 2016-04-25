docker-proxy
============

docker-proxy is a bridge between the network (a server listening on port 9999) and the local `docker.sock` unix socket. It exposes the HTTP
API on a network interface/port. This version was mostly tested with a web browser, as my use was directed towards simple GET JSON endpoints.
Hopefully it works for other people (POST, socket/websocket,...) but *this has not been tested*. The added timeouts/disconnects will most
likely mess with things, as it now behaves more strictly as a HTTP/1.0 proxy, disconnecting clients after responses from the upstream.

I've created this fork to fix some issues with the original code:

1. Better handling of EOF/timeouts (front and back-end connections)
2. Better logging (-v argument 0=none, 1=info, 2=all)
3. Also run the proxy from within docker (README)
4. Actual cleanup of accepted connections

And introduced some new issues:

1. Due to the timeouts some API endpoints (websocket, itd) will not be functional. Submit a PR if you need it.

Running with docker
-------------------

You can run docker-proxy with docker using the `golang` image:

~~~
#!/bin/bash
ARGS=$(cat docker.args | xargs echo -n)
docker run $ARGS --rm=true -it -v `pwd`:/go -w /go golang go run docker-proxy.go "$@"
~~~

The provided `docker.args` file lists additional arguments for the `docker run` command.
It forwards the `docker.sock` unix socket to the container and it also exposes the network port 9999 for access to the docker-proxy service.

The provided example above is included in the script `run`.

Example usage
-------------

Start a proxy service:

~~~
./run
~~~

Start a proxy service with minimal connection logging:

~~~
./run -v 1
~~~

Start a proxy service with full connection request/response logging:

~~~
./run -v 2
~~~

Run the service on a different port

~~~
./run -p 8080
~~~

> Note: This requires modifying `docker.args` to adjust the `-p` argument for docker.


Why?
----

Docker, by default, listens on the Unix socket file. I don't want to run it on tcp socket forever but I don't want to restart docker daemon whenever I want to access docker over network.

How do I use it?
----------------

This is a proxy. So one can just run it on the docker server as a member of docker group or as root (former is preferred) and the use docker to fire commands normally with an additional -H flag.

Example:

```
$ nohup go run docker-proxy.go --port 4321 &
$ docker -H tcp://docker_host:4321 ps
```

> Note: I'm pretty sure that with timeouts, using certain commands will be problematic. For example `logs` with the `-f` option will most likely fail unpredictably.
> As said above, please submit a PR if you need to gracefully handle a more feature-complete set of requests to the docker API.
