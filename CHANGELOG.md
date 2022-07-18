# Changelog

## alpha-1 (unreleased)

### Bug Fixes

* Fixed race conditions around the event getter workers.
* Fixed persistence of state for the task manager and task scheduler. In case of a node exit, all running and future tasks will be saved and resume under node initialization.

### Features

* Refactored the ethereum layer1 interface for security protection.
* Simplified the monitoring system.
* Removed all calls to smart contracts during event processing.
* Unified ethdkg and snapshot tasks under a unique interface called **Task**.
* Task scheduler was decoupled from monitoring and now it's its own service under the executor sub-package.
* Task manager was completely refactored to work with the new Task interface and scheduler.
* Added new features to the deployment scripts. Now it's possible to schedule a maintenance (required to change validators) and unregister validators during the local tests.
* Replaced the old transaction Queue for the new transaction Watcher. This service is now responsible for retrieving transaction receipts, and retrying stale transactions.
* Added persistence of state for the transaction Watcher. In case of a node exit, all transaction waiting for receipt will be saved and resume under node initialization.

### Deprecated

* Removed the field `timeout` from [monitor] section in the user configuration.
* Removed the fields `timeout`, `testEther`, `finalityDelay`, `retryCount`, `retryDelay`, `txFeePercentageToIncrease`, `txCheckFrequency`, `txTimeoutForReplacement` from [ethereum] section in the user configuration. Most of these fields were replaced by constants.
* Renamed the fields `endpointPeers`, `passcodes`, `registryAddress`, `txMaxFeeThresholdInGwei` to `endpointMinimumPeers`, `passCodes`, `factoryAddress` and `txMaxGasFeeAllowedInGwei` accordingly.
