package pkg

import (
	"strings"
	"testing"
)

func TestParseVmName0(t *testing.T) {
	cmd0 := "/usr/bin/kvm -id 470 -name mz-pgpro-8796-ent-load,debug-threads=on -no-shutdown -chardev socket,id=qmp,path=/var/run/qemu-server/470.qmp,server=on,wait=off -mon chardev=qmp,mode=control -chardev socket,id=qmp-event,path=/var/run/qmeventd.sock,reconnect=5 -mon chardev=qmp-event,mode=control -pidfile /var/run/qemu-server/470.pid -daemonize -smbios type=1,uuid=e7b89993-7ca7-40de-897a-174d11ad641b -smp 8,sockets=1,cores=8,maxcpus=8 -nodefaults -boot menu=on,strict=on,reboot-timeout=1000,splash=/usr/share/qemu-server/bootsplash.jpg -vnc unix:/var/run/qemu-server/470.vnc,password=on -cpu host,+kvm_pv_eoi,+kvm_pv_unhalt -m 8192 -device pci-bridge,id=pci.1,chassis_nr=1,bus=pci.0,addr=0x1e -device pci-bridge,id=pci.2,chassis_nr=2,bus=pci.0,addr=0x1f -device vmgenid,guid=c87bcc7d-4ce8-4e37-9eea-330203655804 -device piix3-usb-uhci,id=uhci,bus=pci.0,addr=0x1.0x2 -device usb-tablet,id=tablet,bus=uhci.0,port=1 -chardev socket,id=serial0,path=/var/run/qemu-server/470.serial0,server=on,wait=off -device isa-serial,chardev=serial0 -device VGA,id=vga,bus=pci.0,addr=0x2 -chardev socket,path=/var/run/qemu-server/470.qga,server=on,wait=off,id=qga0 -device virtio-serial,id=qga0,bus=pci.0,addr=0x8 -device virtserialport,chardev=qga0,name=org.qemu.guest_agent.0 -device virtio-balloon-pci,id=balloon0,bus=pci.0,addr=0x3,free-page-reporting=on -iscsi initiator-name=iqn.1993-08.org.debian:01:e1e3b5a89c9 -drive file=iscsi://192.168.4.40/iqn.2012-06.ru.postgrespro.gigastorage5:target0/204,if=none,id=drive-ide0,media=cdrom,aio=io_uring -device ide-cd,bus=ide.0,unit=0,drive=drive-ide0,id=ide0 -drive if=none,id=drive-ide2,media=cdrom,aio=io_uring -device ide-cd,bus=ide.1,unit=0,drive=drive-ide2,id=ide2 -drive file=iscsi://192.168.4.40/iqn.2012-06.ru.postgrespro.gigastorage5:target0/205,if=none,id=drive-virtio0,format=raw,cache=none,aio=io_uring,detect-zeroes=on -device virtio-blk-pci,drive=drive-virtio0,id=virtio0,bus=pci.0,addr=0xa,bootindex=100 -netdev type=tap,id=net0,ifname=tap470i0,script=/var/lib/qemu-server/pve-bridge,downscript=/var/lib/qemu-server/pve-bridgedown,vhost=on,queues=4 -device virtio-net-pci,mac=00:18:59:00:3A:C0,netdev=net0,bus=pci.0,addr=0x12,id=net0,vectors=10,mq=on,packed=on,rx_queue_size=1024,tx_queue_size=1024 -machine type=pc+pve0"
	cmd1 := strings.Split(cmd0, " ")
	vmName, ok := parseVmName(cmd1)
	if !ok {
		t.FailNow()
	}
	if vmName != "mz-pgpro-8796-ent-load" {
		t.FailNow()
	}
}
