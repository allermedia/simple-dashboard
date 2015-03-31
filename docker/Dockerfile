FROM busybox:ubuntu-14.04

MAINTAINER AdamB <adamb@aller.no>

ADD dashboard /usr/bin/dashboard

ENV DASHBOARD_HOST :3000

ADD html/ /etc/dashboard/
WORKDIR /etc/dashboard

ENTRYPOINT ["/usr/bin/dashboard"]