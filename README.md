# Data Denormalizer

0. Runs as a TCP server.
0. Requests can be made from any other application which can open TCP ports.
0. Uses ZeroMQ 'ROUTER' socket. Clients should use ZeroMQ 'REQ', 'DEALER'
   or 'ROUTER'.
0. Work is distributed between workers pool.
0. Workers run on separate threads (goroutines).
0. Each worker has a DB connection.
0. Workers have work buffers.
0. The server distributes requests between workers by selecting the worker
   which has the least items in the buffer.

## Available Commands

Commands are sent as multi-part messages. The following shows the parts space
separated:

0. Question: `Q [ID]`
0. Question joins: `QJ [ID] [COUNT] [PAGE]`
0. Question latest comments: `QLC [ID] [COUNT] [PAGE]`
0. Answer: `A [ID]`
0. Question top answers: `QTA [ID] [COUNT] [PAGE]`
0. Question latest answers: `QLA [ID] [COUNT] [PAGE]`

## Response format

Reponses are sent as a 3-part message:

0. Success? true / false string
0. Empty? true / false string
0. Payload - JSON encoded string
