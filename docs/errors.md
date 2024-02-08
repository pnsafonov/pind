# Сообщения об ошибках

### Не хватает ядер
```
PinState, PinLoad pool.getNumaNodeForLoadAssign failed for cpuCount = 50, vmName = mz-poudriere-db
```
Для виртуальной машины с именем `mz-poudriere-db` не хватает 50 **отдельных**, свободных ядер на **одной** NUMA ноде.

### Idle pool перегружен
```
calcIdlePoolLoad, idle_overwork is high 83.87 >= 80.00 %
```
Набор ядер, на котором исполняются **ненагруженные** виртуальные машины, перегружен и потребляет много процессороного времени.


