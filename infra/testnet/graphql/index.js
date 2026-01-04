import { createServer } from 'node:http';
import { createYoga, createSchema } from 'graphql-yoga';

const RPC_URL = process.env.RPC_URL || 'http://127.0.0.1:26657';
const API_URL = process.env.API_URL || 'http://127.0.0.1:1317';
const PORT = process.env.PORT || 4000;

async function fetchJson(url) {
  const response = await fetch(url);
  if (!response.ok) {
    const text = await response.text();
    throw new Error(`${response.status} ${response.statusText}: ${text}`);
  }
  return response.json();
}

const typeDefs = /* GraphQL */ `
  type SyncInfo {
    latestBlockHeight: String!
    latestBlockTime: String!
    catchingUp: Boolean!
  }

  type NodeInfo {
    network: String!
    moniker: String!
  }

  type Status {
    nodeInfo: NodeInfo!
    syncInfo: SyncInfo!
  }

  type ValidatorSummary {
    address: String!
    moniker: String
    tokens: String
    status: String
  }

  type Query {
    status: Status!
    validators(limit: Int = 50): [ValidatorSummary!]!
    latestBlockHeight: String!
  }
`;

const resolvers = {
  Query: {
    status: async () => {
      const data = await fetchJson(`${RPC_URL}/status`);
      return {
        nodeInfo: {
          network: data.result.node_info.network,
          moniker: data.result.node_info.moniker
        },
        syncInfo: {
          latestBlockHeight: data.result.sync_info.latest_block_height,
          latestBlockTime: data.result.sync_info.latest_block_time,
          catchingUp: data.result.sync_info.catching_up
        }
      };
    },
    validators: async (_, { limit }) => {
      const data = await fetchJson(
        `${API_URL}/cosmos/staking/v1beta1/validators?status=BOND_STATUS_BONDED&pagination.limit=${limit}`
      );
      return (data.validators || []).map((validator) => ({
        address: validator.operator_address,
        moniker: validator.description?.moniker,
        tokens: validator.tokens,
        status: validator.status
      }));
    },
    latestBlockHeight: async () => {
      const data = await fetchJson(`${RPC_URL}/status`);
      return data.result.sync_info.latest_block_height;
    }
  }
};

const yoga = createYoga({
  schema: createSchema({ typeDefs, resolvers }),
  graphqlEndpoint: '/graphql'
});

const server = createServer(yoga);
server.listen(PORT, () => {
  console.log(`GraphQL server running on http://0.0.0.0:${PORT}/graphql`);
});
