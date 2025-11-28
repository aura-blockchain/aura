use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};
use aura_bindings::msg::RegisterVCMsg;
use aura_bindings::query::GetVCQuery;
use aura_bindings::{AuraMsg, AuraQuery, VerifiableCredential};
use cosmwasm_std::{
    entry_point, to_json_binary, Binary, CosmosMsg, Deps, DepsMut, Env, MessageInfo, QueryRequest,
    Response, StdResult,
};

#[entry_point]
pub fn instantiate(
    _deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    _msg: InstantiateMsg,
) -> StdResult<Response> {
    Ok(Response::new())
}

#[entry_point]
pub fn execute(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> StdResult<Response<AuraMsg>> {
    match msg {
        ExecuteMsg::RegisterVc { address, vc_base64 } => {
            register_vc(deps, info, address, vc_base64)
        }
    }
}

pub fn register_vc(
    _deps: DepsMut,
    _info: MessageInfo,
    address: String,
    vc_base64: String,
) -> StdResult<Response<AuraMsg>> {
    let msg = AuraMsg::VCRegistry {
        register_vc: RegisterVCMsg { address, vc_base64 },
    };

    let register_vc_msg: CosmosMsg<AuraMsg> = CosmosMsg::Custom(msg);
    let response = Response::<AuraMsg>::new().add_message(register_vc_msg);
    Ok(response)
}

#[entry_point]
pub fn query(deps: Deps<AuraQuery>, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetVc { address } => to_json_binary(&get_vc(deps, address)?),
    }
}

pub fn get_vc(deps: Deps<AuraQuery>, address: String) -> StdResult<VerifiableCredential> {
    let query = AuraQuery::VCRegistry {
        get_vc: GetVCQuery { address },
    };
    let request: QueryRequest<AuraQuery> = QueryRequest::Custom(query);
    let res: VerifiableCredential = deps.querier.query(&request)?;
    Ok(res)
}
