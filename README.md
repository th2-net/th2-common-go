# th2 common library (GO)

## Usage

1. Import factory package
```
factory "github.com/th2-net/th2-common-go/schema/factory"
```

2. Create factory with configs from the default path (`/var/th2/config/*`):
```
NewCommonFactory()
```

2. Create factory with configs from the specified arguments:
```
NewCommonFactoryFromArgs(rabbit_cfg, rabbit_router_cfg)
```

Usage example: `./example` folder

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
    * attributes - pin's attribute for filtering. Default attributes:
        * first
        * second
        * subscribe
        * publish        
        * parsed
        * raw
        * store
        * event
      
    * filters - pin's message's filters
        * metadata - a metadata filters
        * message - a message's fields filters
    
Filters format: 
* fieldName - a field's name
* expectedValue - expected field's value (not used for all operations)
* operation - operation's type
    * `EQUAL` - the filter passes if the field is equal to the exact value
    * `NOT_EQUAL` - the filter passes if the field does not equal the exact value
    * `EMPTY` - the filter passes if the field is empty
    * `NOT_EMPTY` - the filter passes if the field is not empty
    * `WILDCARD` - filters the field by wildcard expression

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
      ],
      "filters": {
        "metadata": [
          {
            "fieldName": "session-alias",
            "expectedValue": "connection1",
            "operation": "EQUAL"
          }
        ],
        "message": [
          {
            "fieldName": "checkField",
            "expectedValue": "wil?card*",
            "operation": "WILDCARD"
          }
        ]
      }
    }
  }
}
```