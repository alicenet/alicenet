# Layer1 package

The layer 1 package contains all the logic for interacting with all the layer1 blockchains that AliceNet supports.

The package is mainly composed of 3 services:

* **Monitor**: responsible for monitoring and reacting to layer 1 events.
* **Executor**: responsible for scheduling and executing tasks and actions (e.g snapshots, ethdkg, accusations) against the layer 1 blockchains.
* **Transaction**: responsible for watching for layer 1 transactions done by the AliceNet node and retrieve its receipts.

## Interacting with the layer 1 blockchains

The main way to interact with the layer 1 blockchains is via **Tasks**. So, if you have a service that needs to send a transaction to the layer 1 blockchains, create a new task following


