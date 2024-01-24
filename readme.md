# Pind 

**Pind** - daemon for pinning cpu cores to processes

## Run

Run as a service:
```
pind -c example/pind0.yml -s
```

## Install

```
cd pind
go build

sudo su

cp ./pind /usr/bin

mkdir /etc/pind
cp ./example/pind0.yml /etc/pind/pind.yml

cp ./conf/pind.service /lib/systemd/system/pind.service
systemctl daemon-reload

systemctl start pind
systemctl enable pind

```
