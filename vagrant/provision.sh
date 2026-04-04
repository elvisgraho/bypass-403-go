#!/usr/bin/env bash
set -euo pipefail

export DEBIAN_FRONTEND=noninteractive

apt-get update
apt-get install -y nginx golang-go

mkdir -p /opt/bypass-403-playground

cd /vagrant/playground
/usr/bin/go build -o /opt/bypass-403-playground/playground

install -m 0644 /vagrant/vagrant/403-playground.service /etc/systemd/system/403-playground.service
install -m 0644 /vagrant/vagrant/nginx-playground.conf /etc/nginx/sites-available/bypass-403-playground

rm -f /etc/nginx/sites-enabled/default
ln -sf /etc/nginx/sites-available/bypass-403-playground /etc/nginx/sites-enabled/bypass-403-playground

systemctl daemon-reload
systemctl enable 403-playground.service
systemctl restart 403-playground.service
nginx -t
systemctl enable nginx
systemctl restart nginx
