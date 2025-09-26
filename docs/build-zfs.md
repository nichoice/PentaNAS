# 从源码构建 zfs

## 下载

```bash
https://github.com/openzfs/zfs/releases/download/zfs-2.3.4/zfs-2.3.4.tar.gz
```

## 安装依赖
```bash
sudo dnf install --skip-broken epel-release gcc make autoconf automake libtool rpm-build libtirpc-devel libblkid-devel libuuid-devel libudev-devel openssl-devel zlib-devel libaio-devel libattr-devel elfutils-libelf-devel kernel-devel-$(uname -r) python3 python3-devel python3-setuptools python3-cffi libffi-devel git ncompress libcurl-devel

sudo dnf install --skip-broken --enablerepo=epel --enablerepo=powertools python3-packaging dkms
```

## 构建
```bash
[root@localhost zfs-2.3.4]# ./autogen.sh
[root@localhost zfs-2.3.4]# ./configure --prefix=/usr/openzfs
[root@localhost zfs-2.3.4]# make -s -j
[root@localhost zfs-2.3.4]# make install; sudo ldconfig; sudo depmod; make -C modules/ install
```


## Running zloop.sh and zfs-tests.sh

```bash
sudo dnf install --skip-broken ksh bc bzip2 fio acl sysstat mdadm lsscsi parted attr nfs-utils samba rng-tools pax perf popt-devel
```

### dbbench install
```bash
https://dl.fedoraproject.org/pub/epel/10/Everything/x86_64/Packages/d/dbench-4.0-32.el10_0.x86_64.rpm
dnf localinstall dbench-4.0-32.el10_0.x86_64.rpm
```
