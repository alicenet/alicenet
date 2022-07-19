# Layer1 package

The layer 1 package contains all the logic for interacting with all the layer1 blockchains that AliceNet supports.

The package is mainly composed of 3 services:

* **Monitor**: responsible for monitoring and reacting to layer 1 events.
* **Executor**: responsible for scheduling and executing tasks and actions (e.g snapshots, ethdkg, accusations) against the layer 1 blockchains.
* **Transaction**: responsible for watching for layer 1 transactions done by the AliceNet node and retrieve its receipts.

## Creating new Tasks

The main way to interact with the layer 1 blockchains is via **Tasks**. So, if you have a service that needs to send a transaction to the layer 1 blockchains or need to check data from time to time in a smart contract create a new task.

The new task should be a new file in its own folder inside the **tasks folder**. Check the [example task](./executor/tasks/examples/simple_task.go) to get more information on how to create a new task. Once the task file is created, its type need to be registered into the `GetTaskRegistry()` function in the task [scheduler file](./executor/task_scheduler.go). Once all these requirements all fulfilled, use the task scheduler to schedule the task when needed.

## Registering new layer 1 events

The main way to react to changes in the layer1 blockchains is via events. The layer1 smart contracts will emit events based on the actions performed by the users and the Alicenet node can listen to these events and react accordingly. For instance, the node will create deposits UTXO when someone deposits tokens in the layer1 smart contracts, it will commit the snapshot to the consensus database when someone commits a snapshot that was validated by the smart contracts, etc. For more examples, check the **monitor/events/** folder.

If you have an event in which the alicenet node should be listening and reacting to, you will need to create a `callback` function. Some, examples of `callback` functions can be seen in the folder `./monitor/events`. Before creating a new file, check if the `callback` function cannot be grouped with the files that currently exists.

Once the `callback` is created, it needs to be registered in jointly with the event ABI data into the `EventMap`. Although the parameters of the `callback` function are not fixed (anything can be passed as long we have access to the data at registration), the `callback` function will need to be wrapped on the type [EventProcessor](./monitor/objects/event_map.go#L11) which has fixed inputs. Check the [events/setup](./monitor/events/setup.go) for examples on how to get the event ABI data for a layer1 smart contract and how to properly register a `callback` function.

Finally, with all requirements fulfilled, Alicenet will be watching for the specified event and it will be calling the `callback` function once the event is observed.
