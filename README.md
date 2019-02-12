# 404-server

`Copy of https://github.com/kubernetes/ingress-gce/tree/master/cmd/404-server`

404-server is a simple webserver that satisfies the ingress, which means it has to do two things:

Serves a 404 page at /
Serves a 200 at /healthcheck
