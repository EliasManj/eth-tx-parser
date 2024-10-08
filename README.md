## Tx Parser

### Goal
Implement Ethereum blockchain parser that will allow to query transactions for subscribed addresses.

### Problem
Users not able to receive push notifications for incoming/outgoing transactions. By Implementing Parser interface we would be able to hook this up to notifications service to notify about any incoming/outgoing transactions.

### Limitations
* Use Go Language
* Avoid usage of external libraries
* Use Ethereum JSONRPC to interact with Ethereum Blockchain
* Use memory storage for storing any data (should be easily extendable to
support any storage in the future)

Expose public interface for external usage either via command line or http api that
will include supported list of operations defined in the Parser interface

```go
type Parser interface {
// last parsed block
GetCurrentBlock() int
// add address to observer
Subscribe(address string) bool
// list of inbound or outbound transactions for an address
GetTransactions(address string) []Transaction
}
```

### Endpoint
URL: https://ethereum-rpc.publicnode.com

Request Example:
```bash
curl -X POST 'https://ethereum-rpc.publicnode.com' --data \
'{
"jsonrpc": "2.0",
"method": "eth_blockNumber",
"params": [],
"id": 83
}'
// Result
{
"id":83,
"jsonrpc": "2.0",
"result": "0x4b7" // 1207
}
```

### Parser parameters

Start parser with specific RPC URl 
```bash
go run main.go -url=[rpc url]
```

Start parser in a specific block number
```bash
go run main.go -startblock=[block number]
```

### Running locally with Anvil

Start Anvil service
```bash
anvil
```

Start service using local RPC url
```bash
go run main.go -url="http://localhost:8545"
```

### Running with Sepolia Testnet

Start service using a Sepolia RPC url
```bash
go run main.go -url="https://ethereum-sepolia-rpc.publicnode.com/"
```

### Running local unit tests

Start Anvil service
```bash
anvil
```

Run package tests
```
go test ./rpcclient
```

```
go test ./api
```

### API endpoints


**Get Current Block**

```
/getCurrentBlock
```

**Subscribe to Address**

This endpoint is used to subscribe to a specific address.
```
/subscribe?address=[addres]
```
Request Parameters

* address ((string, required)): The address to subscribe to.


**List Subscribed Addresses**

List all address subscriptions.
```
/subscribe?address=[addres]
```

**Get Inbound and Outbound transactions**

List all address subscriptions for the specified address.
```
/getTransactions?address=[addres]
```

Parameters

* address (string, required): The address for which transactions are to be retrieved.

The response is a JSON array containing transaction details. The schema for the response is as follows:

```
[
    {
        "txhash": "",
        "blockhash": "",
        "blocknumber": "",
        "from": "",
        "to": "",
        "txtype": "",
        "gasUsed": "",
        "blobGasPrice": "",
        "contractAddress": "",
        "nonce": ""
    }
]
```