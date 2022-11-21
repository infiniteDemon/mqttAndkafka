# 下载文件 启动程序之前需要把插件放好到对应的位置
这里是个confluent上提供的一个插件下载地址
https://www.confluent.io/hub/

放好插件之后
再启动 docker-compose.yaml

使用配套包中的docker-compose.yaml文件进行启动
```yaml
version: '3.3'

services:
  mosquitto:
    image: eclipse-mosquitto:1.5.5
    hostname: mosquitto
    container_name: mosquitto
    expose:
      - "1883"
    ports:
      - "1883:1883"
  zookeeper:
    image: zookeeper:3.4.9
    restart: unless-stopped
    hostname: zookeeper
    container_name: zookeeper
    ports:
      - "2181:2181"
    environment:
      ZOO_MY_ID: 1
      ZOO_PORT: 2181
      ZOO_SERVERS: server.1=zookeeper:2888:3888
    volumes:
      - ./zookeeper/data:/data
      - ./zookeeper/datalog:/datalog
  kafka:
    image: confluentinc/cp-kafka:5.1.0
    hostname: kafka
    container_name: kafka
    ports:
      - "9092:9092"
    environment:
      KAFKA_LISTENER_SECURITY_PROTOCOL_MAP: PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://kafka:9092,PLAINTEXT_HOST://localhost:29092
      KAFKA_ZOOKEEPER_CONNECT: "zookeeper:2181"
      KAFKA_BROKER_ID: 1
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
    volumes:
      - ./kafka/data:/var/lib/kafka/data
    depends_on:
      - zookeeper
  kafka-connect:
    image: confluentinc/cp-kafka-connect:5.1.0
    hostname: kafka-connect
    container_name: kafka-connect
    ports:
      - "8083:8083"
    environment:
      CONNECT_BOOTSTRAP_SERVERS: "kafka:9092"
      CONNECT_REST_ADVERTISED_HOST_NAME: connect
      CONNECT_REST_PORT: 8083
      CONNECT_GROUP_ID: compose-connect-group
      CONNECT_CONFIG_STORAGE_TOPIC: docker-connect-configs
      CONNECT_OFFSET_STORAGE_TOPIC: docker-connect-offsets
      CONNECT_STATUS_STORAGE_TOPIC: docker-connect-status
      CONNECT_KEY_CONVERTER: org.apache.kafka.connect.json.JsonConverter
      CONNECT_VALUE_CONVERTER: org.apache.kafka.connect.json.JsonConverter
      CONNECT_INTERNAL_KEY_CONVERTER: "org.apache.kafka.connect.json.JsonConverter"
      CONNECT_INTERNAL_VALUE_CONVERTER: "org.apache.kafka.connect.json.JsonConverter"
      CONNECT_CONFIG_STORAGE_REPLICATION_FACTOR: "1"
      CONNECT_OFFSET_STORAGE_REPLICATION_FACTOR: "1"
      CONNECT_STATUS_STORAGE_REPLICATION_FACTOR: "1"
      CONNECT_PLUGIN_PATH: '/usr/share/java,/etc/kafka-connect/jars'
      CONNECT_CONFLUENT_TOPIC_REPLICATION_FACTOR: 1
    volumes:
      - ./jars:/etc/kafka-connect/jars
    depends_on:
      - zookeeper
      - kafka
      - mosquitto
  mongo-db:
    image: mongo:4.0.5
    hostname: mongo-db
    container_name: mongo-db
    expose:
      - "27017"
    ports:
      - "27017:27017"
    command: --bind_ip_all --smallfiles
    volumes:
      - ./mongo-db:/data
  mongoclient:
    image: mongoclient/mongoclient:2.2.0
    container_name: mongoclient
    hostname: mongoclient
    depends_on:
      - mongo-db
    ports:
      - 3000:3000
    environment:
      MONGO_URL: "mongodb://mongo-db:27017"
      PORT: 3000
    expose:
      - "3000"
```
输入 docker-compose up -d

等待一会 等待程序启动完成之后
```shell
vim connect-mqtt-source.json
```
```json
{
  "name": "mqtt-source",
  "config": {
    "connector.class": "io.confluent.connect.mqtt.MqttSourceConnector",
    "tasks.max": 1,
    "mqtt.server.uri": "tcp://mosquitto:1883",
    "mqtt.topics": "baeldung",
    "kafka.topic": "connect-custom",
    "value.converter": "org.apache.kafka.connect.converters.ByteArrayConverter",
    "confluent.topic.bootstrap.servers": "kafka:9092",
    "confluent.topic.replication.factor": 1
  }
}
```

docker-compose ps查看程序是否正常运行，都正常运行之后，键入
```shell
curl -d @connect-mqtt-source.json -H "Content-Type: application/json" -X POST http://localhost:8083/connectors
```

这时候另外起一个客户端，我们启动一个对kafka数据的读取
输入
```shell
docker run --rm --network test_default confluentinc/cp-kafka:5.1.0 kafka-console-consumer --bootstrap-server kafka:9092 --topic connect-custom --from-beginning
```

然后 我们回到原来的shell通道中，输入
```shell
docker run -it --rm --name mqtt-publisher --network test_default efrecon/mqtt-client pub -h mosquitto  -t "baeldung" -m  "{\"schema\":{\"type\":\"struct\",\"fields\":[{\"type\":\"int64\",\"optional\":false,\"field\":\"registertime\"},{\"type\":\"string\",\"optional\":false,\"field\":\"userid\"},{\"type\":\"string\",\"optional\":false,\"field\":\"regionid\"},{\"type\":\"string\",\"optional\":false,\"field\":\"gender\"}],\"optional\":false,\"name\":\"ksql.users\"},\"payload\":{\"registertime\":1493819497170,\"userid\":\"User_1\",\"regionid\":\"Region_5\",\"gender\":\"MALE\"}}"
```

如果在kafka读取的客户端中能够正常读取，那就说明我们的kafka-connect-mqtt成功
接下来进行kafka to mongodb的落盘
```shell
vim connect-mongodb-sink.json
```
输入
```json
{
  "name": "mongo-sink",
  "config": {
    "connector.class": "com.mongodb.kafka.connect.MongoSinkConnector",
    "tasks.max": 1,
    "topics": "connect-custom",
    "connection.uri": "mongodb://mongo-db:27017/quickstart?retryWrites=true",
    "database": "quickstart",
    "collection": "MyCollection",
    "key.converter": "org.apache.kafka.connect.json.JsonConverter",
    "key.converter.schemas.enable": false,
    "value.converter": "org.apache.kafka.connect.json.JsonConverter",
    "value.converter.schemas.enable": false
  }
}

```

然后我们再进行http请求进行配置
```shell
curl -d @connect-mongodb-sink.json -H "Content-Type: application/json" -X POST http://localhost:8083/connectors
```

这时已经配置成功，我们进行再一次的往mqtt中写入数据
```shell
docker run -it --rm --name mqtt-publisher --network test_default efrecon/mqtt-client pub -h mosquitto  -t "baeldung" -m  "{\"schema\":{\"type\":\"struct\",\"fields\":[{\"type\":\"int64\",\"optional\":false,\"field\":\"registertime\"},{\"type\":\"string\",\"optional\":false,\"field\":\"userid\"},{\"type\":\"string\",\"optional\":false,\"field\":\"regionid\"},{\"type\":\"string\",\"optional\":false,\"field\":\"gender\"}],\"optional\":false,\"name\":\"ksql.users\"},\"payload\":{\"registertime\":1493819497170,\"userid\":\"User_1\",\"regionid\":\"Region_5\",\"gender\":\"MALE\"}}"
```

然后打开一个网页 访问http://{{ip}}:3000/
访问我们的mongodb，可以看到MyCollection的collection中已经存在数据了