# Running the test net blockchain in a node-per-server setup

## Setup a node infrastructure on a server

1. Copy docker-compose.yml to the server
2. Copy *setup-node-infrastructure-for-testnet.sh* to the same folder on the server
3. Generate a new UUID for ORGANIZATION_ID and add it to the script. It will be used later to setup the organization and workgroups via rest endpoints
3. Run *setup-node-infrastructure-for-testnet.sh*
      This one sets up all components necessary for a node to function.
      Run it on as many servers as nodes are needed.
4. Open all ports in the firewall that are necessary for external communication


## SETUP INITIAL VALIDATOR NODE

NODE 1: docker exec baseledger-node_blockchain_app_1 baseledgerd init node1 --chain-id baseledger
NODE 1: docker exec baseledger-node_blockchain_app_1 baseledgerd keys add node1_validator --keyring-backend test

- name: node1_validator
  type: local
  address: baseledger1clfvetf3qzsghyquj7hz02v0juyap504wt9rww
  pubkey: baseledgerpub1addwnpepq2ulja9xav85vrgdkej7x67d3vlykl7cu0whe2sjrjmwldgtct5syts4e3y
  mnemonic: ""
  threshold: 0
  pubkeys: []

tendermint node id fa9e205e2591309888374fe6f1e931e92a7352e1

### Add first node validator account as genesis
NODE 1: node1_validator_address=$(docker exec baseledger-node_blockchain_app_1 baseledgerd keys show node1_validator -a --keyring-backend test)
           docker exec baseledger-node_blockchain_app_1 baseledgerd add-genesis-account baseledger1clfvetf3qzsghyquj7hz02v0juyap504wt9rww 100000000000stake,100000000000token

### Generate genensis transaction on the first node
NODE 1: docker exec baseledger-node_blockchain_app_1 baseledgerd gentx node1_validator 100000000stake --chain-id baseledger --keyring-backend test

### Collect genesis transactions
NODE 1: docker exec baseledger-node_blockchain_app_1 baseledgerd collect-gentxs

### Enable rest api, it is only enable = false entry, maybe we can make it a bit more precise?
NODE 1: docker exec baseledger-node_blockchain_app_1 sed -i 's/enable = false/enable = true/' ~/.baseledger/config/app.toml

### Enables grpc
NODE 1:docker exec baseledger-node_blockchain_app_1 sed -i 's@laddr = "tcp://127.0.0.1:'26657'"@laddr = "tcp://0.0.0.0:'26657'"@' ~/.baseledger/config/config.toml

### Allow connecting peers not in the address book
NODE 1: docker exec baseledger-node_blockchain_app_1 sed -i 's/addr_book_strict = true/addr_book_strict = false/' ~/.baseledger/config/config.toml

### Allow connections from localhost to tendermint API
NODE 1: docker exec baseledger-node_blockchain_app_1 sed -i 's/allow_duplicate_ip = false/allow_duplicate_ip = true/' ~/.baseledger/config/config.toml

### Increase the timeout between blocks to 9s
NODE 1: docker exec baseledger-node_blockchain_app_1 sed -i 's/timeout_commit = "5s"/timeout_commit = "9s"/' ~/.baseledger/config/config.toml

### Run the node
NODE 1: docker exec baseledger-node_blockchain_app_1 baseledgerd start


## SETUP ADDITIONAL REPLICATOR NODE

NODE 2: repeat first five steps or running the *setup-node-infrastructure-for-testnet.sh* on the machine
NODE 2: copy *add-node-to-running-blokchain.sh* to the same folder on the machine
NODE 2: make sure node1_id and node1_ip adress in the script are correct
NODE 2: copy genesis.json from node 1 to the same folder on the machine
NODE 2: run *add-node-to-running-blokchain.sh*

*****************************************
- name: node55_replicator
  type: local
  address: baseledger1quz8telhz7tt3sv4m7fdh6ueu6lpn0ypt6w2ff
  pubkey: baseledgerpub1addwnpepqgtdm7laxgq6fltf0ujfx0cp6y3ksctfxqfwvr072zg3u7m07wt5vxrqdmp
  mnemonic: ""
  threshold: 0
  pubkeys: []

tendermint node id 1228ef8024fa89d8d9ab756689856813ce69682b
tendermint validator id baseledgervalconspub1zcjduepq3sp0443lqkltzxwe2vfs09psjmw6ps9umssqz7n0c3n5ampggh0q6df0m7
validator operator address baseledgervaloper1quz8telhz7tt3sv4m7fdh6ueu6lpn0yp5wll36

*****************************************
name: node42_replicator
  type: local
  address: baseledger14q4zlymjfa39swywtk3vtcy3fgjjgtszt2vfs4
  pubkey: baseledgerpub1addwnpepqf3ukpzazk4f98wdcp5p8a5dz4thm0p46vgafz4xf2znp0ay8faljac5dgz
  mnemonic: ""
  threshold: 0
  pubkeys: []



### How to copy a genesis from initial node to new node

NODE 1: docker cp baseledger-node_blockchain_app_1:/root/.baseledger/config/genesis.json .
NODE 1: Copy genesis.json to clipboard
NODE 2: copy cliboard to genesis.json
NODE 2: docker cp ./genesis.json baseledger-node_blockchain_app_1:/root/.baseledger/config/genesis.json


### Run the node
NODE 2: docker exec baseledger-node_blockchain_app_1 baseledgerd start
if it fails for any reason, try to run *docker exec baseledger-node_starport_1 baseledgerd unsafe-reset-all* before start command

### Send TOKENS to the node
NODE2: node2_adress = docker exec baseledger-node_blockchain_app_1 baseledgerd keys show node2_validator -a
NODE1: docker exec baseledger-node_blockchain_app_1 baseledgerd tx bank send node1_validator baseledger1quz8telhz7tt3sv4m7fdh6ueu6lpn0ypt6w2ff 100000token --yes

## ADD REPLICATOR NODE AS VALIDATOR

### Send a minimal amount of STAKE tokens from Node1 to the Node_to_become_validator

NODE1: docker exec baseledger-node_blockchain_app_1 baseledgerd tx bank send node1_validator baseledger1quz8telhz7tt3sv4m7fdh6ueu6lpn0ypt6w2ff 1stake --yes

Here baseledger1xax2e85vqn4n26wxk0qfcy9jgmwlgvnw750hzm is the receiver address obtained from baseledgerd keys list command

### Node_To_become_Validator now take the minimal amount of STAKE tokens received and stakes them to make him a validator:

NODE 2: docker exec baseledger-node_blockchain_app_1 baseledgerd tx staking create-validator  --amount=1stake  --pubkey=baseledgervalconspub1zcjduepq3sp0443lqkltzxwe2vfs09psjmw6ps9umssqz7n0c3n5ampggh0q6df0m7 --moniker="node55"  --commission-rate="0.10" --commission-max-rate="0.20" --commission-max-change-rate="0.01" --min-self-delegation="1" --from=node55_replicator --yes 

In the command above i removed (-gas="200000" --gas-prices="0.025stake" ) as we assume to have 0 gas costs that way
Params explanation:
--from = <name of the node to become validator>
--pubkey <output of tendermint show-validator on node_to_become_validator>
--moniker= <unique name for the validator>


### Now the new validator should be in the validator set in status UNBONDED (he has to few tokens staked to participate). 
We stake the right amount from Node1 (our token controlling node):

NODE1: docker exec baseledger-node_blockchain_app_1 baseledgerd tx staking delegate baseledgervaloper1quz8telhz7tt3sv4m7fdh6ueu6lpn0yp5wll36 100000000stake --from=node1_validator --yes 

Params explanation:
--baseledgervaloper-address from the new validator node, can be seen in "docker exec first_node_blockchain_app_1 baseledgerd query staking validators"
--from=<our token controlling node1>