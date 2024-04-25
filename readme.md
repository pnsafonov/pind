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

## Test
To run go tests:
```
cd pind
go test ./...
```

## Вывод топологии процессора
```
pind --print-numa-phys

numa 0
phys core 0, siblings = [0 8]
phys core 1, siblings = [1 9]
phys core 2, siblings = [2 10]
phys core 3, siblings = [3 11]
phys core 4, siblings = [4 12]
phys core 5, siblings = [5 13]
phys core 6, siblings = [6 14]
phys core 7, siblings = [7 15]
```
8 физических ядер, 16 логических.   

## load_type
В случае `load_type: logical` следует указаывать ядра из диапазона `0-15`.   
В случае `load_type: phys` следует указаывать ядра из диапазона `0-8`.

## Документация
* [Описание](docs/description.md)
* [Конфигурационный файл](docs/pind.yml)
* [Ошибки в логах](docs/errors.md)
