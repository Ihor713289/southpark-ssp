# Assignment
```
Welcome to the quiet mountain town of South Park… where the residents now have their own messaging system!
Your mission is to build a small distributed microservice(s) application that lets South Park characters send and receive funny messages asynchronously.
You’ll implement:
•	A Go HTTP API using Hexagonal Architecture (Ports and Adapters)
•	A RabbitMQ message broker
•	A Python consumer service that reads messages from RabbitMQ and displays them

1.	Project components:
Go API Service
•	Exposes a POST /messages endpoint.
Accepts JSON like:
{
  "author": "Cartman",
  "body": "Respect my authoritah!"
}
•	Go API project has RabbitMQ injected as a port and adapter.
•	Publishes this message to a RabbitMQ queue named southpark_messages.
RabbitMQ Broker
•	Acts as the “mailbox” of South Park — the queue stores messages until someone consumes them.
Python Consumer
•	Listens to southpark_messages queue.
•	When a message arrives, prints something like: “Respect my authoritah!”

2.	Setup and steps:

Set up the Go project following the given structure (app, domain, ports, adapters).
Create the HTTP endpoint that accepts a message and calls the publisher port.
Implement the RabbitMQ publisher adapter (publishes JSON messages).
Run RabbitMQ (use Docker)
Handle invalid JSON or missing fields gracefully in the Go API.
Python microservice consumes the messages from RabbitMQ
Python microservice displayes consumed messages to the console

```
