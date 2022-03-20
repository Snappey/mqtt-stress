# mqtt-stress
Used for generating fake data to stress test MQTT brokers.

## Configuration

```shell
  -delay string
        delay per message per worker(s), e.g. 500ms or 1s or 1m (default "500ms")
  -fields string
        <name>:<type> custom fields to generate structs from seperated by a "," where type is number, string, id, phone or email. e.g. customer:id,customer_email:email (overrides payload flag)
  -help
        print application options
  -increment
        payload is incremented for each message sent, each worker has an independent count (overrides payload flag)
  -password string
        password to authenticate with mqtt broker
  -payload string
        custom payload sent for each message, if left random payloads will be generated for each message
  -run string
        length of time to run application before exiting (default "15s")
  -url string
        scheme://host:port Where "scheme" is one of "tcp", "ssl", or "ws", "host" is the ip-address (or hostname) and "port" is the port on which the broker is accepting connections.
  -username string
        username to authenticate with mqtt broker
  -workers int
        amount of clients to connect to mqtt broker with (default 10)
```

## Example Configuration

### Generate Static Payload
```shell
docker run --rm snappey/mqtt-stress:latest -username admin -password admin123 -url tcp://mqtt.broker.com -payload mqtt-stress -workers 10 -delay 500ms -run 30s
```

### Increment Payload
```shell
docker run --rm snappey/mqtt-stress:latest -username admin -password admin123 -url tcp://mqtt.broker.com -increment -workers 10 -delay 500ms -run 30s
```

### Generate Random Struct Data
```shell
docker run --rm snappey/mqtt-stress:latest -username admin -password admin123 -url tcp://mqtt.broker.com -fields customer_id:id,customer_phone:phone,data:number -workers 10 -delay 500ms -run 30s
```