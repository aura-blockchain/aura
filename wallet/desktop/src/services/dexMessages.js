import { BinaryReader, BinaryWriter } from 'cosmjs-types/binary';
import { Coin } from 'cosmjs-types/cosmos/base/v1beta1/coin';

const createBaseMsgCreatePool = () => ({
  creator: '',
  denomA: '',
  denomB: '',
  amountA: undefined,
  amountB: undefined
});

export const MsgCreatePool = {
  typeUrl: '/aura.dex.v1beta1.MsgCreatePool',
  encode(message, writer = BinaryWriter.create()) {
    if (message.creator !== '') {
      writer.uint32(10).string(message.creator);
    }
    if (message.denomA !== '') {
      writer.uint32(18).string(message.denomA);
    }
    if (message.denomB !== '') {
      writer.uint32(26).string(message.denomB);
    }
    if (message.amountA !== undefined) {
      Coin.encode(message.amountA, writer.uint32(34).fork()).ldelim();
    }
    if (message.amountB !== undefined) {
      Coin.encode(message.amountB, writer.uint32(42).fork()).ldelim();
    }
    return writer;
  },
  decode(input, length) {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    const end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseMsgCreatePool();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.denomA = reader.string();
          break;
        case 3:
          message.denomB = reader.string();
          break;
        case 4:
          message.amountA = Coin.decode(reader, reader.uint32());
          break;
        case 5:
          message.amountB = Coin.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },
  fromPartial(object) {
    const message = createBaseMsgCreatePool();
    message.creator = object.creator ?? '';
    message.denomA = object.denomA ?? '';
    message.denomB = object.denomB ?? '';
    message.amountA = object.amountA ? Coin.fromPartial(object.amountA) : undefined;
    message.amountB = object.amountB ? Coin.fromPartial(object.amountB) : undefined;
    return message;
  }
};

const createBaseMsgAddLiquidity = () => ({
  provider: '',
  poolId: '',
  amountA: undefined,
  amountB: undefined
});

export const MsgAddLiquidity = {
  typeUrl: '/aura.dex.v1beta1.MsgAddLiquidity',
  encode(message, writer = BinaryWriter.create()) {
    if (message.provider !== '') {
      writer.uint32(10).string(message.provider);
    }
    if (message.poolId !== '') {
      writer.uint32(18).string(message.poolId);
    }
    if (message.amountA !== undefined) {
      Coin.encode(message.amountA, writer.uint32(26).fork()).ldelim();
    }
    if (message.amountB !== undefined) {
      Coin.encode(message.amountB, writer.uint32(34).fork()).ldelim();
    }
    return writer;
  },
  decode(input, length) {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    const end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseMsgAddLiquidity();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.provider = reader.string();
          break;
        case 2:
          message.poolId = reader.string();
          break;
        case 3:
          message.amountA = Coin.decode(reader, reader.uint32());
          break;
        case 4:
          message.amountB = Coin.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },
  fromPartial(object) {
    const message = createBaseMsgAddLiquidity();
    message.provider = object.provider ?? '';
    message.poolId = object.poolId ?? '';
    message.amountA = object.amountA ? Coin.fromPartial(object.amountA) : undefined;
    message.amountB = object.amountB ? Coin.fromPartial(object.amountB) : undefined;
    return message;
  }
};

const createBaseMsgRemoveLiquidity = () => ({
  provider: '',
  poolId: '',
  lpTokens: ''
});

export const MsgRemoveLiquidity = {
  typeUrl: '/aura.dex.v1beta1.MsgRemoveLiquidity',
  encode(message, writer = BinaryWriter.create()) {
    if (message.provider !== '') {
      writer.uint32(10).string(message.provider);
    }
    if (message.poolId !== '') {
      writer.uint32(18).string(message.poolId);
    }
    if (message.lpTokens !== '') {
      writer.uint32(26).string(message.lpTokens);
    }
    return writer;
  },
  decode(input, length) {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    const end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseMsgRemoveLiquidity();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.provider = reader.string();
          break;
        case 2:
          message.poolId = reader.string();
          break;
        case 3:
          message.lpTokens = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },
  fromPartial(object) {
    const message = createBaseMsgRemoveLiquidity();
    message.provider = object.provider ?? '';
    message.poolId = object.poolId ?? '';
    message.lpTokens = object.lpTokens ?? '';
    return message;
  }
};

const createBaseMsgSwapExactIn = () => ({
  sender: '',
  poolId: '',
  coinIn: undefined,
  minAmountOut: '',
  maxSlippageBps: BigInt(0)
});

export const MsgSwapExactIn = {
  typeUrl: '/aura.dex.v1beta1.MsgSwapExactIn',
  encode(message, writer = BinaryWriter.create()) {
    if (message.sender !== '') {
      writer.uint32(10).string(message.sender);
    }
    if (message.poolId !== '') {
      writer.uint32(18).string(message.poolId);
    }
    if (message.coinIn !== undefined) {
      Coin.encode(message.coinIn, writer.uint32(26).fork()).ldelim();
    }
    if (message.minAmountOut !== '') {
      writer.uint32(34).string(message.minAmountOut);
    }
    if (message.maxSlippageBps !== BigInt(0)) {
      writer.uint32(40).uint64(message.maxSlippageBps);
    }
    return writer;
  },
  decode(input, length) {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    const end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseMsgSwapExactIn();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.sender = reader.string();
          break;
        case 2:
          message.poolId = reader.string();
          break;
        case 3:
          message.coinIn = Coin.decode(reader, reader.uint32());
          break;
        case 4:
          message.minAmountOut = reader.string();
          break;
        case 5:
          message.maxSlippageBps = reader.uint64();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },
  fromPartial(object) {
    const message = createBaseMsgSwapExactIn();
    message.sender = object.sender ?? '';
    message.poolId = object.poolId ?? '';
    message.coinIn = object.coinIn ? Coin.fromPartial(object.coinIn) : undefined;
    message.minAmountOut = object.minAmountOut ?? '';
    if (object.maxSlippageBps !== undefined && object.maxSlippageBps !== null) {
      message.maxSlippageBps = BigInt(object.maxSlippageBps.toString());
    } else {
      message.maxSlippageBps = BigInt(0);
    }
    return message;
  }
};

const createBaseMsgCreateOrder = () => ({
  creator: '',
  orderType: 0,
  auraAmount: '',
  otherCoin: '',
  otherAmount: ''
});

export const MsgCreateOrder = {
  typeUrl: '/aura.dex.v1beta1.MsgCreateOrder',
  encode(message, writer = BinaryWriter.create()) {
    if (message.creator !== '') {
      writer.uint32(10).string(message.creator);
    }
    if (message.orderType !== 0) {
      writer.uint32(16).int32(message.orderType);
    }
    if (message.auraAmount !== '') {
      writer.uint32(26).string(message.auraAmount);
    }
    if (message.otherCoin !== '') {
      writer.uint32(34).string(message.otherCoin);
    }
    if (message.otherAmount !== '') {
      writer.uint32(42).string(message.otherAmount);
    }
    return writer;
  },
  decode(input, length) {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    const end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseMsgCreateOrder();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.orderType = reader.int32();
          break;
        case 3:
          message.auraAmount = reader.string();
          break;
        case 4:
          message.otherCoin = reader.string();
          break;
        case 5:
          message.otherAmount = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },
  fromPartial(object) {
    const message = createBaseMsgCreateOrder();
    message.creator = object.creator ?? '';
    message.orderType = object.orderType ?? 0;
    message.auraAmount = object.auraAmount ?? '';
    message.otherCoin = object.otherCoin ?? '';
    message.otherAmount = object.otherAmount ?? '';
    return message;
  }
};

const createBaseMsgCancelOrder = () => ({
  creator: '',
  orderId: ''
});

export const MsgCancelOrder = {
  typeUrl: '/aura.dex.v1beta1.MsgCancelOrder',
  encode(message, writer = BinaryWriter.create()) {
    if (message.creator !== '') {
      writer.uint32(10).string(message.creator);
    }
    if (message.orderId !== '') {
      writer.uint32(18).string(message.orderId);
    }
    return writer;
  },
  decode(input, length) {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    const end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseMsgCancelOrder();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.orderId = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },
  fromPartial(object) {
    const message = createBaseMsgCancelOrder();
    message.creator = object.creator ?? '';
    message.orderId = object.orderId ?? '';
    return message;
  }
};

const createBaseMsgExecuteSwap = () => ({
  initiator: '',
  orderId: '',
  secret: ''
});

export const MsgExecuteSwap = {
  typeUrl: '/aura.dex.v1beta1.MsgExecuteSwap',
  encode(message, writer = BinaryWriter.create()) {
    if (message.initiator !== '') {
      writer.uint32(10).string(message.initiator);
    }
    if (message.orderId !== '') {
      writer.uint32(18).string(message.orderId);
    }
    if (message.secret !== '') {
      writer.uint32(26).string(message.secret);
    }
    return writer;
  },
  decode(input, length) {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    const end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseMsgExecuteSwap();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.initiator = reader.string();
          break;
        case 2:
          message.orderId = reader.string();
          break;
        case 3:
          message.secret = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },
  fromPartial(object) {
    const message = createBaseMsgExecuteSwap();
    message.initiator = object.initiator ?? '';
    message.orderId = object.orderId ?? '';
    message.secret = object.secret ?? '';
    return message;
  }
};

const createBaseMsgCreateHTLC = () => ({
  sender: '',
  recipient: '',
  amount: undefined,
  secretHash: '',
  timelockDuration: BigInt(0)
});

export const MsgCreateHTLC = {
  typeUrl: '/aura.dex.v1beta1.MsgCreateHTLC',
  encode(message, writer = BinaryWriter.create()) {
    if (message.sender !== '') {
      writer.uint32(10).string(message.sender);
    }
    if (message.recipient !== '') {
      writer.uint32(18).string(message.recipient);
    }
    if (message.amount !== undefined) {
      Coin.encode(message.amount, writer.uint32(26).fork()).ldelim();
    }
    if (message.secretHash !== '') {
      writer.uint32(34).string(message.secretHash);
    }
    if (message.timelockDuration !== BigInt(0)) {
      writer.uint32(40).uint64(message.timelockDuration);
    }
    return writer;
  },
  decode(input, length) {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    const end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseMsgCreateHTLC();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.sender = reader.string();
          break;
        case 2:
          message.recipient = reader.string();
          break;
        case 3:
          message.amount = Coin.decode(reader, reader.uint32());
          break;
        case 4:
          message.secretHash = reader.string();
          break;
        case 5:
          message.timelockDuration = reader.uint64();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },
  fromPartial(object) {
    const message = createBaseMsgCreateHTLC();
    message.sender = object.sender ?? '';
    message.recipient = object.recipient ?? '';
    message.amount = object.amount ? Coin.fromPartial(object.amount) : undefined;
    message.secretHash = object.secretHash ?? '';
    if (object.timelockDuration !== undefined && object.timelockDuration !== null) {
      message.timelockDuration = BigInt(object.timelockDuration.toString());
    } else {
      message.timelockDuration = BigInt(0);
    }
    return message;
  }
};

const createBaseMsgClaimHTLC = () => ({
  recipient: '',
  htlcId: '',
  secret: ''
});

export const MsgClaimHTLC = {
  typeUrl: '/aura.dex.v1beta1.MsgClaimHTLC',
  encode(message, writer = BinaryWriter.create()) {
    if (message.recipient !== '') {
      writer.uint32(10).string(message.recipient);
    }
    if (message.htlcId !== '') {
      writer.uint32(18).string(message.htlcId);
    }
    if (message.secret !== '') {
      writer.uint32(26).string(message.secret);
    }
    return writer;
  },
  decode(input, length) {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    const end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseMsgClaimHTLC();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.recipient = reader.string();
          break;
        case 2:
          message.htlcId = reader.string();
          break;
        case 3:
          message.secret = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },
  fromPartial(object) {
    const message = createBaseMsgClaimHTLC();
    message.recipient = object.recipient ?? '';
    message.htlcId = object.htlcId ?? '';
    message.secret = object.secret ?? '';
    return message;
  }
};

const createBaseMsgRefundHTLC = () => ({
  sender: '',
  htlcId: ''
});

export const MsgRefundHTLC = {
  typeUrl: '/aura.dex.v1beta1.MsgRefundHTLC',
  encode(message, writer = BinaryWriter.create()) {
    if (message.sender !== '') {
      writer.uint32(10).string(message.sender);
    }
    if (message.htlcId !== '') {
      writer.uint32(18).string(message.htlcId);
    }
    return writer;
  },
  decode(input, length) {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    const end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseMsgRefundHTLC();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.sender = reader.string();
          break;
        case 2:
          message.htlcId = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },
  fromPartial(object) {
    const message = createBaseMsgRefundHTLC();
    message.sender = object.sender ?? '';
    message.htlcId = object.htlcId ?? '';
    return message;
  }
};

const createBaseMsgCommitOrder = () => ({
  sender: '',
  commitHash: new Uint8Array()
});

export const MsgCommitOrder = {
  typeUrl: '/aura.dex.v1beta1.MsgCommitOrder',
  encode(message, writer = BinaryWriter.create()) {
    if (message.sender !== '') {
      writer.uint32(10).string(message.sender);
    }
    if (message.commitHash.length !== 0) {
      writer.uint32(18).bytes(message.commitHash);
    }
    return writer;
  },
  decode(input, length) {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    const end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseMsgCommitOrder();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.sender = reader.string();
          break;
        case 2:
          message.commitHash = reader.bytes();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },
  fromPartial(object) {
    const message = createBaseMsgCommitOrder();
    message.sender = object.sender ?? '';
    message.commitHash = object.commitHash ?? new Uint8Array();
    return message;
  }
};

const createBaseMsgRevealOrder = () => ({
  sender: '',
  commitId: '',
  orderType: 0,
  auraAmount: '',
  otherCoin: '',
  otherAmount: '',
  salt: new Uint8Array()
});

export const MsgRevealOrder = {
  typeUrl: '/aura.dex.v1beta1.MsgRevealOrder',
  encode(message, writer = BinaryWriter.create()) {
    if (message.sender !== '') {
      writer.uint32(10).string(message.sender);
    }
    if (message.commitId !== '') {
      writer.uint32(18).string(message.commitId);
    }
    if (message.orderType !== 0) {
      writer.uint32(24).int32(message.orderType);
    }
    if (message.auraAmount !== '') {
      writer.uint32(34).string(message.auraAmount);
    }
    if (message.otherCoin !== '') {
      writer.uint32(42).string(message.otherCoin);
    }
    if (message.otherAmount !== '') {
      writer.uint32(50).string(message.otherAmount);
    }
    if (message.salt.length !== 0) {
      writer.uint32(58).bytes(message.salt);
    }
    return writer;
  },
  decode(input, length) {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    const end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseMsgRevealOrder();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.sender = reader.string();
          break;
        case 2:
          message.commitId = reader.string();
          break;
        case 3:
          message.orderType = reader.int32();
          break;
        case 4:
          message.auraAmount = reader.string();
          break;
        case 5:
          message.otherCoin = reader.string();
          break;
        case 6:
          message.otherAmount = reader.string();
          break;
        case 7:
          message.salt = reader.bytes();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },
  fromPartial(object) {
    const message = createBaseMsgRevealOrder();
    message.sender = object.sender ?? '';
    message.commitId = object.commitId ?? '';
    message.orderType = object.orderType ?? 0;
    message.auraAmount = object.auraAmount ?? '';
    message.otherCoin = object.otherCoin ?? '';
    message.otherAmount = object.otherAmount ?? '';
    message.salt = object.salt ?? new Uint8Array();
    return message;
  }
};

export const DEX_TYPE_REGISTRY = [
  [MsgCreatePool.typeUrl, MsgCreatePool],
  [MsgAddLiquidity.typeUrl, MsgAddLiquidity],
  [MsgRemoveLiquidity.typeUrl, MsgRemoveLiquidity],
  [MsgSwapExactIn.typeUrl, MsgSwapExactIn],
  [MsgCreateOrder.typeUrl, MsgCreateOrder],
  [MsgCancelOrder.typeUrl, MsgCancelOrder],
  [MsgExecuteSwap.typeUrl, MsgExecuteSwap],
  [MsgCreateHTLC.typeUrl, MsgCreateHTLC],
  [MsgClaimHTLC.typeUrl, MsgClaimHTLC],
  [MsgRefundHTLC.typeUrl, MsgRefundHTLC],
  [MsgCommitOrder.typeUrl, MsgCommitOrder],
  [MsgRevealOrder.typeUrl, MsgRevealOrder]
];
