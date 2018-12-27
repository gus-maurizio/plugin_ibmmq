# IBM MQ plugin

## Start MQ
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
  -v $GOPATH/src/github.com/gus-maurizio/plugin_ibmmq:/goutils \
  ibmcom/mq:9.1.1.0


docker exec --tty --interactive ibmqm1 /bin/bash
dspmq
runmqsc

docker run --rm -it ubuntu:16.04 /bin/bash


open https://localhost:9443/ibmmq/console/login.html

```

### Observations
```
Channel 'SYSTEM.DEF.SVRCONN' to host '127.0.0.1' ended abnormally.
export MQSAMP_USER_ID=admin
./amqsconn QM1 "SYSTEM.DEF.SVRCONN" "localhost(1414)"

display qmgr CONNAUTH CHLAUTH
# QMNAME(QM1) CHLAUTH(ENABLED)  CONNAUTH(DEV.AUTHINFO)
display authinfo(DEV.AUTHINFO)

./amqsconn QM1 DEV.ADMIN.SVRCONN "localhost(1414)"

export MQSAMP_USER_ID=app
export MQSAMP_PASSWORD=passw0rd
./amqsconn QM1 DEV.APP.SVRCONN "ibmqm1(1414)"

export MQSERVER='DEV.APP.SVRCONN/TCP/ibmqm1(1414)'

```

Compile all
```
for i in *.go; do  echo $i;j=$(echo $i|sed 's/.go//g');echo $j;echo go build -o bin/$j $i;go build -o bin/$j $i; done
```

## Preparation for MQ GO
```
docker build -f Dockerfile-build-mqgosdk -t gmaurizio/mqgosdk . 
Use this to compile
docker run --rm -it \
--env GOROOT=/usr/local/go \
--env GOPATH=/mnt \
--env PATH=/usr/local/go/bin:$PATH \
--network  test \
--name     mqgosdk \
--hostname mqgosdk \
-v $GOPATH:/mnt \
gmaurizio/mqgosdk


docker run --rm -it \
  --network test \
  --name     test1 \
  --hostname test1 \
  -v $GOPATH/src/github.com/gus-maurizio/plugin_ibmmq:/goutils \
  ubuntu:18.10 /bin/bash

export LD_LIBRARY_PATH=/goutils/metrics/:$LD_LIBRARY_PATH
export MQSAMP_USER_ID=app 
./amqsconn QM1 DEV.APP.SVRCONN "ibmqm1(1414)"




```

you might need to authorize channel `SET CHLAUTH('SYSTEM.DEF.SVRCONN') TYPE(BLOCKUSER) USERLIST(ALLOWANY).`
Contact the systems administrator, who should examine the channel
authentication records to ensure that the correct settings have been
configured. The ALTER QMGR CHLAUTH switch is used to control whether channel
authentication records are used. The command DISPLAY CHLAUTH can be used to
query the channel authentication records.
```

app:x:1001:1002::/home/app:
(mq:9.1.1.0)root@ibmqm1:/# id admin
uid=1000(admin) gid=1000(admin) groups=1000(admin),999(mqm)
(mq:9.1.1.0)root@ibmqm1:/# id app  
uid=1001(app) gid=1002(app) groups=1002(app),1001(mqclient)
(mq:9.1.1.0)root@ibmqm1:/# 


 DISPLAY CHLAUTH('SYSTEM.DEF.SVRCONN')

display qmgr CONNAUTH
display authinfo(DEV.AUTHINFO)

ALTER AUTHINFO(DEV.AUTHINFO) AUTHTYPE(IDPWOS) CHCKCLNT(NONE)
REFRESH SECURITY TYPE(CONNAUTH)

---------- USE ----------------------------

display qmgr CONNAUTH CHLAUTH

# Disable AUTH on channels
ALTER QMGR CHLAUTH(DISABLED)
ALTER QMGR CHLAUTH(ENABLED)

### try CONNAUTH(DEV.AUTHINFO) replacing with more relaxed one

DEFINE AUTHINFO(DONT.USE.PW) AUTHTYPE(IDPWOS) FAILDLAY(1) CHCKLOCL(OPTIONAL) CHCKCLNT(OPTIONAL)
ALTER QMGR CONNAUTH(DONT.USE.PW)
REFRESH SECURITY TYPE(CONNAUTH)

SET CHLAUTH(‘*’) TYPE(ADDRESSMAP) ADDRESS(‘*’)
USERSRC(CHANNEL) CHCKCLNT(REQUIRED)
SET CHLAUTH(‘*’) TYPE(SSLPEERMAP)
SSLPEER(‘CN=*’) USERSRC(CHANNEL)
CHCKCLNT(ASQMGR)

---------- END USE ----------------------------

set CHLAUTH(*) TYPE(BLOCKUSER) USERLIST('nobody','*MQADMIN')
set CHLAUTH(SYSTEM.*) TYPE(BLOCKUSER) USERLIST('nobody')

SET CHLAUTH('*') TYPE(ADDRESSMAP) ADDRESS('*') USERSRC(CHANNEL) CHCKCLNT(ASQMGR) DESCR('BackStop rule') ACTION(REPLACE)

ALTER AUTHINFO(SYSTEM.DEFAULT.AUTHINFO.IDPWOS) AUTHTYPE(IDPWOS) CHCKCLNT(OPTIONAL)
REFRESH SECURITY TYPE(CONNAUTH)


ALTER QMGR CHLAUTH(DISABLED)


ALTER QMGR CONNAUTH(USE.PW)
DEFINE AUTHINFO(USE.PW) AUTHTYPE(xxxxxx)
FAILDLAY(1) CHCKLOCL(OPTIONAL)
CHCKCLNT(REQUIRED)
REFRESH SECURITY TYPE(CONNAUTH)



SET CHLAUTH(*) TYPE(BLOCKUSER) USERLIST('*NOACCESS')
DEFINE CHANNEL(SYSTEM.DEF.SRVCONN) CHLTYPE(SVRCONN) MCAUSER('mqm') REPLACE
ALTER AUTHINFO(SYSTEM.DEFAULT.AUTHINFO.IDPWOS) AUTHTYPE(IDPWOS) CHCKCLNT(OPTIONAL)
REFRESH SECURITY TYPE(CONNAUTH)



setmqaut -m QM1 -t q -n SYSTEM.ADMIN.STATISTICS.QUEUE -g mqm +inq +browse +get +dsp
setmqaut -m QM1 -t q -n SYSTEM.ADMIN.STATISTICS.QUEUE -p app +inq +browse +get +dsp

DISPLAY QMGR STATINT
ALTER QMGR STATINT(30)

DISPLAY QMGR STATQ STATCHL STATMQI
ALTER QMGR STATQ(ON) STATCHL(MEDIUM) STATMQI(ON) ACCTINT(60) ACCTMQI(ON) ACCTQ(ON) STATINT(30)

RESET QMGR TYPE(STATISTICS)

DISPLAY QUEUE STATISTICS


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
