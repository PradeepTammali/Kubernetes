
1. Set env vars 

```
export DEFAULT_USERID=0
export DEFAULT_USERNAME=""
export DEFAULT_GROUPID=0
export DEFAULT_GROUPNAME=""
export ENABLE_WEBHOOKS=false


export MAPR_LICENSE_MODULES=
export MAPR_MEMORY=0
export MAPR_MOUNT_PATH=/mapr
export MAPR_SECURITY=
export MAPR_TZ=Europe/Stockholm
export MAPR_USER=
export MAPR_HS_HOST=
export MAPR_OT_HOSTS=
export MAPR_ZK_HOSTS=
export container=docker
export DEBIAN_FRONTEND=noninteractive
export MAPR_CLDB_HOSTS=
export MAPR_CLUSTER=
export MAPR_CONTAINER_GID=0
export MAPR_CONTAINER_GROUP=root
export MAPR_CONTAINER_UID=0
export MAPR_CONTAINER_USER=root
export TERM=xterm
```


2. Install dependencies 
```
apt-get update -qq && apt-get install --no-install-recommends -q -y vim net-tools curl sudo tzdata wget apt-utils dnsutils file iputils-ping net-tools nfs-common openssl syslinux sysv-rc-conf libssl1.0.0 openjdk-8-jdk && apt-get autoremove --purge -q -y && rm -rf /var/lib/apt/lists/* && apt-get clean -q
```


3. Download mapr-setup.sh and install.
```
wget https://package.mapr.com/releases/installer/mapr-setup.sh -P /tmp
chmod +x /tmp/mapr-setup.sh
mkdir -p /opt/mapr/installer/docker
mv /tmp/mapr-setup.sh /opt/mapr/installer/docker/
/opt/mapr/installer/docker/mapr-setup.sh -r http://package.mapr.com/releases container client 6.1.0 6.0.0 mapr-client mapr-posix-client-container mapr-hbase mapr-asynchbase mapr-hive mapr-pig mapr-spark mapr-kafka mapr-librdkafka
```

4. Copy trustore of mapr and set env vars like user, uid, group and gid
``` 
docker cp /home/pradeep/ssl_truststore f69551aa4141:/opt/mapr/conf/

export MAPR_CONTAINER_GID=0
export MAPR_CONTAINER_GROUP=root
export MAPR_CONTAINER_UID=0
export MAPR_CONTAINER_USER=root
export MAPR_TICKETFILE_LOCATION=/tmp/maprticket_$MAPR_CONTAINER_UID
```


5. Running the mapr client to generate ticket.
```
/opt/mapr/installer/docker/mapr-setup.sh container
maprlogin password -cluster $MAPR_CLUSTER -user $MAPR_CONTAINER_USER 
cat $MAPR_TICKETFILE_LOCATION | base64 -w 0
```

# Reference:
https://mapr.com/docs/home/MapRInstaller.html
https://mapr.com/docs/home/AdvancedInstallation/c_installer_how_it_works.html#concept_mt1_xzx_ft
