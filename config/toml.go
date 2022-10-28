package config

import (
	"bytes"
	"text/template"
)

func CreateTOML(config *RootConfiguration) ([]byte, error) {
	var err error
	tmpl := template.New("configFileTemplate")
	configTemplate, err := tmpl.Parse(defaultConfigTemplate)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	if err := configTemplate.Execute(&buffer, config); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

const defaultConfigTemplate = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml


#######################################################################
###                   Main Base Config Options                      ###
#######################################################################

[chain]

# AliceNet ChainID that corresponds to the AliceNet Network that you are trying
# to connect.
id = {{ .Chain.ID }}

# Path to the location where to save the monitor database. The monitor database
# is responsible for storing the (events, tasks, receipts) coming from layer 1
# blockchains that AliceNet is anchored with. Store this database in a safe
# location. If this database is deleted, the node will replay all events by
# querying the layer1 chains using the information provided below.
monitorDB = "{{ .Chain.MonitorDbPath }}"

# Flag to save the monitor database only on memory. USE ONLY RECOMMENDED TO SET
# TRUE FOR TESTING PURPOSES.
monitorDBInMemory = {{ .Chain.MonitorDbInMemory }}

# Path to the location where to save the state database. The state database is
# responsible for storing the AliceNet blockchain data (blocks, validator's
# data). Store this database in a safe location. If this database is deleted,
# the node will re-sync with its peers from scratch. DON'T DELETE THIS DATABASE
# IF YOU ARE RUNNING A VALIDATOR NODE!!! If this database is deleted and you
# are running a validator node, the validator's data will be permanently
# deleted and the node will not be able to proceed with its work as a
# validator, even after a re-sync. Therefore, you may be susceptible to a
# slashing event
stateDB = "{{ .Chain.StateDbPath }}"

# Flag to save the state database only on memory. USE ONLY RECOMMENDED TO SET
# TRUE FOR TESTING PURPOSES.
stateDBInMemory = {{ .Chain.StateDbInMemory }}

# Path to the location where to save the transaction database. The transaction
# database is responsible for storing the AliceNet blockchain data
# (transactions). Store this database in a safe location. If this database is
# deleted, the node will re-sync all transactions with its peers.
transactionDB = "{{ .Chain.TransactionDbPath }}"

# Flag to save the transaction database only on memory. USE ONLY RECOMMENDED TO
# SET TRUE FOR TESTING PURPOSES.
transactionDBInMemory = {{ .Chain.TransactionDbInMemory }}

[ethereum]

# Ethereum address that will be used to sign transactions and connect to the
# AliceNet services on ethereum.
defaultAccount = "{{ .Ethereum.DefaultAccount }}"

# Ethereum endpoint url to the ethereum chain where the AliceNet network
# infra-structure that you are trying to connect lives. Ethereum mainnet for
# AliceNet mainnet and Goerli for AliceNet Testnet. Infura and Alchemy services
# can be used, but if you are running your own validator node, we recommend to
# use a more secure node.
endpoint = "{{ .Ethereum.Endpoint }}"

# Minimum number of peers connected to your ethereum node that you wish to
# reach before trying to process ethereum blocks to retrieve the AliceNet
# events.
endpointMinimumPeers = {{ .Ethereum.EndpointMinimumPeers }}

# Path to the encrypted private key used on the address above.
keystore = "{{ .Ethereum.Keystore }}"

# Path to the file containing the password to unlock the account private key.
passCodes = "{{ .Ethereum.PassCodes }}"

# The ethereum block where the AliceNet contract factory was deployed. This
# block is used as starting block to retrieve all events (e.g snapshots,
# deposits) necessary to initialize and validate your AliceNet node.
startingBlock = {{ .Ethereum.StartingBlock }}

# Ethereum address of the AliceNet factory of smart contracts. The factory is
# responsible for registering and deploying every contract used by the AliceNet
# infra-structure.
factoryAddress = "{{ .Ethereum.FactoryAddress }}"

# Batch size of blocks that will be downloaded and processed from your endpoint
# address. Every ethereum block starting from the startingBlock until the
# latest ethereum block will be downloaded and all events (e.g snapshots,
# deposits) coming from AliceNet smart contracts will be processed in a
# chronological order. If this value is too large, your endpoint may end up
# being overloaded with API requests.
processingBlockBatchSize = {{ .Ethereum.ProcessingBlockBatchSize }}

# The maximum gas price that you are willing to pay (in GWEI) for a transaction
# done by your node. If you are validator, putting this value too low can
# result in your node failing to fulfill the validators duty, hence, being
# passive for a slashing.
txMaxGasFeeAllowedInGwei = {{ .Ethereum.TxMaxGasFeeAllowedInGwei }}

# Flag to decide if the ethereum transactions information will be shown on the
# logs.
txMetricsDisplay = {{ .Ethereum.TxMetricsDisplay }}


#######################################################################
###                   Logging Config Options                        ###
#######################################################################

[loglevel]

# Log level per service. See all possible services in
# https://github.com/alicenet/alicenet/blob/main/constants/shared.go
consensus = "{{ .Logging.Consensus }}"

[utils]

# Flag to decide if the status will be shown on the logs. Maybe be a little
# noisy.
status = {{ .Utils.Status }}


#######################################################
###       Network Configuration Options             ###
#######################################################

[transport]

# Address to a bootnode running on the desired AliceNet network that you are
# trying to connect with. A bootnode is a software client responsible for
# sharing information about aliceNet peers. Your node will connect to a
# bootnode to retrieve an initial list of peers to try to connect with.
bootNodeAddresses = "{{ .Transport.BootNodeAddresses }}"

# Address and port where your node will be listening for rpc requests.
localStateListeningAddress = "{{ .Transport.LocalStateListeningAddress }}"

# Maximum number of peers that we can connect from a same ip.
originLimit = {{ .Transport.OriginLimit }}

# Address and port where you node will be listening for requests coming from
# other peers. The address should be publicly reachable.
p2pListeningAddress = "{{ .Transport.P2PListeningAddress }}"

# Maximum number of peers that you wish to be connected with. Upper bound to
# limit bandwidth shared with the peers.
peerLimitMax = {{ .Transport.PeerLimitMax }}

# Minimum number of peers that you wish to be connect with, before trying to
# attempt to download blockchain data and participate in consensus.
peerLimitMin = {{ .Transport.PeerLimitMin }}

# 16 Byte private key that is used to encrypt and decrypt information shared
# with peers. Generate this with a safe random generator.
privateKey = "{{ .Transport.PrivateKey }}"

# If UPNP should be used to discover opened ports to connect with the peers.
upnp = {{ .Transport.UPnP }}


#######################################################
###       Validator Configuration Options           ###
#######################################################

# OPTIONAL: Only necessary if you plan to run a validator node.
[validator]

# Passphrase that will be used to encrypt private keys in the database.
symmetricKey = "{{ .Validator.SymmetricKey }}"

`
