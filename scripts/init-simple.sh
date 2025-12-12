#!/bin/bash
set -e

cd /home/hudson/blockchain-projects/aura/chain
rm -rf /tmp/node1 /tmp/node2

echo "" | ./aurad init node1 --chain-id test1 --home /tmp/node1 >/dev/null 2>&1
echo "" | ./aurad keys add val1 --keyring-backend test --home /tmp/node1 >/dev/null 2>&1
ADDR=$(./aurad keys show val1 -a --keyring-backend test --home /tmp/node1 2>/dev/null)
./aurad genesis add-genesis-account "$ADDR" 100000000000uaura --home /tmp/node1 >/dev/null 2>&1
echo "" | ./aurad genesis gentx val1 50000000000uaura --chain-id test1 --keyring-backend test --home /tmp/node1 >/dev/null 2>&1
./aurad genesis collect-gentxs --home /tmp/node1 >/dev/null 2>&1

echo "" | ./aurad init node2 --chain-id test1 --home /tmp/node2 >/dev/null 2>&1
cp /tmp/node1/config/genesis.json /tmp/node2/config/genesis.json

NODE1=$(./aurad tendermint show-node-id --home /tmp/node1 2>/dev/null)
NODE2=$(./aurad tendermint show-node-id --home /tmp/node2 2>/dev/null)
sed -i "s/^persistent_peers = .*/persistent_peers = \"${NODE2}@validator-2:26656\"/" /tmp/node1/config/config.toml
sed -i "s/^persistent_peers = .*/persistent_peers = \"${NODE1}@validator-1:26656\"/" /tmp/node2/config/config.toml
sed -i 's/^addr_book_strict = .*/addr_book_strict = false/' /tmp/node1/config/config.toml /tmp/node2/config/config.toml

docker volume create aura_validator-1-data >/dev/null
docker volume create aura_validator-2-data >/dev/null
docker run --rm -v aura_validator-1-data:/dest -v /tmp/node1:/src:ro alpine sh -c "cp -r /src/* /dest/ && chown -R 1000:1000 /dest"
docker run --rm -v aura_validator-2-data:/dest -v /tmp/node2:/src:ro alpine sh -c "cp -r /src/* /dest/ && chown -R 1000:1000 /dest"

echo "Gentxs: $(jq '.app_state.genutil.gen_txs | length' /tmp/node1/config/genesis.json)"
rm -rf /tmp/node1 /tmp/node2
echo "Done"
