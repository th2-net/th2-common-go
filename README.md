# th2 common library GO (0.3.0)

## Usage

1. Import factory package

```sh
factory "github.com/th2-net/th2-common-go/schema/factory"
```

2. Create factory with configs from the default path (`var/th2/config/*`) located on same level as caller file:

```sh
NewCommonFactory()
```

3. Create factory with configs from the passed path of directory where config files are located and their extension (default is .json):

```sh
NewCommonFactory()
```

```sh
go run *.go -config-file-path="path" -config-file-extension="ext"
```

### Configuration formats

The `CommonFactory` reads a RabbitMQ configuration from the rabbitMQ.json file.

* host - the required setting defines the RabbitMQ host.
* vHost - the required setting defines the virtual host that will be used for connecting to RabbitMQ.
   Please see more details about the virtual host in RabbitMQ via [link](https://www.rabbitmq.com/vhosts.html)
* port - the required setting defines the RabbitMQ port.
* username - the required setting defines the RabbitMQ username.
   The user must have permission to publish messages via routing keys and subscribe to message queues.
* password - the required setting defines the password that will be used for connecting to RabbitMQ.
* exchangeName - the required setting defines the exchange that will be used for sending/subscribing operation in MQ routers.
   Please see more details about the exchanges in RabbitMQ via [link](https://www.rabbitmq.com/tutorials/amqp-concepts.html#exchanges)
* connectionTimeout - the connection TCP establishment timeout in milliseconds with its default value set to 60000. Use zero for infinite waiting.
* connectionCloseTimeout - the timeout in milliseconds for completing all the close-related operations, use -1 for infinity, the default value is set to 10000.
* maxRecoveryAttempts - this option defines the number of reconnection attempts to RabbitMQ, with its default value set to 5.
   The `th2_readiness` probe is set to false and publishers are blocked after a lost connection to RabbitMQ. The `th2_readiness` probe is reverted to true if the connection will be recovered during specified attempts otherwise the `th2_liveness` probe will be set to false.
* minConnectionRecoveryTimeout - this option defines a minimal interval in milliseconds between reconnect attempts, with its default value set to 10000. Common factory increases the reconnect interval values from minConnectionRecoveryTimeout to maxConnectionRecoveryTimeout.
* maxConnectionRecoveryTimeout - this option defines a maximum interval in milliseconds between reconnect attempts, with its default value set to 60000. Common factory increases the reconnect interval values from minConnectionRecoveryTimeout to maxConnectionRecoveryTimeout.
* prefetchCount - this option is the maximum number of messages that the server will deliver, with its value set to 0 if unlimited, the default value is set to 10.
* messageRecursionLimit - an integer number denotes how deep nested protobuf message might be, default = 100

```json
{
  "host": "<host>",
  "vHost": "<virtual host>",
  "port": 5672,
  "username": "<user name>",
  "password": "<password>",
  "exchangeName": "<exchange name>",
  "connectionTimeout": 60000,
  "connectionCloseTimeout": 10000,
  "maxRecoveryAttempts": 5,
  "minConnectionRecoveryTimeout": 10000,
  "maxConnectionRecoveryTimeout": 60000,
  "prefetchCount": 10,
  "messageRecursionLimit": 100
}
```

The `CommonFactory` reads a message's router configuration from the `mq.json` file.

* queues - the required settings defines all pins for an application
   * name - routing key in RabbitMQ for sending
   * queue - queue's name in RabbitMQ for subscribe
   * exchange - exchange in RabbitMQ
   * attributes - pin's attribute for mark. Default attributes:
      * subscribe
      * publish
      * parsed
      * raw
      * event
      * store

```json
{
  "queues": {
    "pin1": {
      "name": "routing_key_1",
      "queue": "queue_1",
      "exchange": "exchange",
      "attributes": [
        "publish",
        "subscribe"
      ]
    }
  }
}
```

## Release notes

### 0.3.1

* Use int type for Prometheus Port config option
* Refactored gRPC module logging

### 0.3.0

+ Migrated to github.com/th2-net/th2-grpc-common-go

### 0.2.0

#### Dependencies

+ Deprecated `streadway/amqp` dependency was replaced with `rabbitmq/amqp091-go`.
   NOTE: this should not be visible to end users as the MQ is only used internally.

### 0.1.0

#### Implemented:

+ Module architecture
+ Zerolog configuration on CommonFactory creation
+ Read and resolve environment variables during configuration reading
+ Base version of gRPC module

   + Start server / Start service features

+ Base version of MQ module

   + EventBatch / MessageGroupBatch router
   + SendAll / SubscribeAll / SubscribeAllWithManualAck features

+ Prometheus module

   + Liveness / Readiness probe
   + MQ module metrics
