SET POSTGRES_EXPOSED_PORT=5432& SET NATS_EXPOSED_PORT=4222& SET BLOCKCHAIN_APP_API_PORT=1317& SET TENDERMINT_NODE_GRPC_PORT=26657& SET TENDERMINT_NODE_PORT=26655& SET PROXY_APP_PORT=8081& SET ORGANIZATION_ID=f4ccf7bc-dd31-4fab-b233-959f9d52ebbb

docker-compose -p baseledger-node up -d