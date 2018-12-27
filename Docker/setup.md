# IBM MQ plugin

## Preparation for MQ GO
```
docker build -f Dockerfile-build-mqgosdk -t gmaurizio/mqgosdk . 

docker run --rm -it \
--env GOROOT=/usr/local/go \
--env GOPATH=/mnt \
--env PATH=/usr/local/go/bin:$PATH \
--network  test \
--name     mqgosdk \
--hostname mqgosdk \
-v $GOPATH:/mnt \
gmaurizio/mqgosdk


```

This is based on IBM MQ 9.1.1.0 image in docker.
See https://github.com/ibm-messaging/mq-container/blob/master/docs/usage.md
To test you can use:
```
docker run --rm -it \
  --env LICENSE=accept \
  --env MQ_QMGR_NAME=QM1 \
  --env MQ_ENABLE_METRICS=true \
  --env MQ_ADMIN_PASSWORD=passw0rd \
  --env MQ_APP_PASSWORD=passw0rd \
  --env LOG_FORMAT=basic \
  --publish 1414:1414 \
  --publish 9443:9443 \
  --publish 9157:9157 \
  --network test \
  --name     ibmqm1 \
  --hostname ibmqm1 \
  ibmcom/mq:9.1.1.0

docker run --rm -it ubuntu:16.04 /bin/bash


open https://localhost:9443/ibmmq/console/login.html

docker build -f Dockerfile-build-mqsdk -t gmaurizio/go-mqsdk .

docker exec --tty --interactive ibmqm1 dspmq
docker exec --tty --interactive gmaurizio/go-mqsdk
docker run --rm -it \
--env PATH=/usr/local/go/bin:$PATH \
--env GOROOT=/usr/local/go \
--env GOPATH=$HOME/go \
gmaurizio/go-mqsdk
```

Here is an example corresponding 20-config.mqsc script from the mqdev blog, which allows users with passwords to connect on the PASSWORD.SVRCONN channel:
```
DEFINE CHANNEL(PASSWORD.SVRCONN) CHLTYPE(SVRCONN) REPLACE
SET CHLAUTH(PASSWORD.SVRCONN) TYPE(BLOCKUSER) USERLIST('nobody') DESCR('Allow privileged users on this channel')
SET CHLAUTH('*') TYPE(ADDRESSMAP) ADDRESS('*') USERSRC(NOACCESS) DESCR('BackStop rule')
SET CHLAUTH(PASSWORD.SVRCONN) TYPE(ADDRESSMAP) ADDRESS('*') USERSRC(CHANNEL) CHCKCLNT(REQUIRED)
ALTER AUTHINFO(SYSTEM.DEFAULT.AUTHINFO.IDPWOS) AUTHTYPE(IDPWOS) ADOPTCTX(YES)
REFRESH SECURITY TYPE(CONNAUTH)
```
--volume myqm1data:/mnt/mqm 

### Examples of API REST
```
curl -i -k -u admin:passw0rd https://localhost:9443/ibmmq/rest/v1/admin/qmgr/QM1/channel/MYCHANNEL
curl -i -k -u admin:passw0rd https://localhost:9444/ibmmq/rest/v1/admin/qmgr/QM2/channel/MYCHANNEL

curl -i -k -u admin:passw0rd https://localhost:9443/ibmmq/rest/v1/admin/qmgr/QM1/queue?name=DEV.XMIT*\&status=*
curl -i -k -u admin:passw0rd https://localhost:9444/ibmmq/rest/v1/admin/qmgr/QM2/queue?name=DEV.*\&status=*

runmqsc QM1
  
display qlocal(XQ1)
display qlocal(LQ1)



curl    -i -k -u admin:passw0rd \
                https://localhost:9443/ibmmq/rest/v1/admin/qmgr

curl    -i -k -u admin:passw0rd \
                https://localhost:9443/ibmmq/rest/v1/admin/qmgr/QM1/queue/LQ1?status=*


curl    -i -k -u admin:passw0rd \
                https://localhost:9443/ibmmq/rest/v1/admin/qmgr/QM1/queue/LQ1?status=status.currentDepth


curl    -i -k -u admin:passw0rd \
                https://localhost:9443/ibmmq/rest/v1/admin/qmgr/QM1/queue?name=SYSTEM.CHANNEL*\&status=status.currentDepth

curl    -i -k -u admin:passw0rd \
                https://localhost:9443/ibmmq/rest/v1/admin/qmgr/QM1/queue?name=DEV*\&status=status.currentDepth

curl    -i -k -u admin:passw0rd \
                https://localhost:9443/ibmmq/rest/v1/admin/qmgr/QM1/queue?name=*\&status=status.currentDepth

The following example sends a persistent message to queue LQ1 on queue manager QM1, with an expiry of 2 minutes.
The message contains the text "Hello World!":


curl    -i -k -u app:passw0rd \
                -H "ibm-mq-rest-csrf-token: none" \
                -H "Content-Type: text/plain;charset=utf-8" \
                -H "ibm-mq-md-persistence: nonPersistent" \
                -H "ibm-mq-md-messageId: 414d5120514d4144455620202020202067d8ce5923582f07" \
                -H "ibm-mq-md-correlationId: 414d5120514d4144455620202020202067d8ce5923582f07" \
                -H "ibm-mq-md-expiry: 120000" \
                https://localhost:9443/ibmmq/rest/v1/messaging/qmgr/QM1/queue/DEV.QUEUE.1/message \
                -X POST --data "Hello World $(date)"

curl    -k -u app:passw0rd \
--data-binary \
-X DELETE \
--header 'ibm-mq-rest-csrf-token: noneed' \
--header 'Accept: text/plain' \
'https://localhost:9443/ibmmq/rest/v1/messaging/qmgr/QM1/queue/SYSTEM.ADMIN.STATISTICS.QUEUE/message?wait=30000' \
-o stats.record

```
