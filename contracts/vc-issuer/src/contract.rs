use aura_bindings::msg::RegisterVCMsg;
use aura_bindings::AuraMsg;
use cosmwasm_std::{
    entry_point, to_json_binary, Addr, Binary, CosmosMsg, Deps, DepsMut, Env, MessageInfo,
    Response, StdResult,
};
use cw_storage_plus::Map;
use sha2::{Digest, Sha256};

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg, IssuedResponse, PendingRequestsResponse, QueryMsg};
use crate::state::{
    Config, IssueRecord, IssueRequest, IssuerProfile, RequestStatus, CONFIG, ISSUED,
    ISSUED_BY_SUBJECT, ISSUERS, MINT_COUNTERS, PENDING_BY_ISSUER, REQUESTS,
};

const MAX_REQUESTS_RETURNED: usize = 25;
const SECONDS_PER_DAY: u64 = 86_400;

#[entry_point]
pub fn instantiate(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response<AuraMsg>, ContractError> {
    let admin = msg
        .admin
        .map(|a| deps.api.addr_validate(&a))
        .transpose()?
        .unwrap_or_else(|| info.sender.clone());

    CONFIG.save(
        deps.storage,
        &Config {
            admin: admin.clone(),
        },
    )?;

    Ok(Response::new()
        .add_attribute("action", "instantiate")
        .add_attribute("admin", admin)
        .add_attribute("contract", env.contract.address))
}

#[entry_point]
pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response<AuraMsg>, ContractError> {
    match msg {
        ExecuteMsg::RegisterIssuer {
            issuer,
            policy_id,
            daily_limit,
        } => execute_register_issuer(deps, env.clone(), info, issuer, policy_id, daily_limit),
        ExecuteMsg::UpdateIssuerStatus { issuer, active } => {
            execute_update_issuer_status(deps, info, issuer, active)
        }
        ExecuteMsg::RequestVc {
            issuer,
            subject,
            vc_type,
            metadata,
        } => execute_request_vc(deps, env.clone(), issuer, subject, vc_type, metadata),
        ExecuteMsg::FulfillRequest {
            request_id,
            credential_base64,
        } => execute_fulfill_request(deps, info, env, request_id, credential_base64),
        ExecuteMsg::RevokeVc { vc_id, reason } => execute_revoke_vc(deps, info, vc_id, reason),
    }
}

#[entry_point]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::Issuer { address } => {
            let addr = deps.api.addr_validate(&address)?;
            let issuer = ISSUERS.load(deps.storage, &addr)?;
            to_json_binary(&issuer)
        }
        QueryMsg::PendingRequests { issuer } => {
            let addr = deps.api.addr_validate(&issuer)?;
            let mut requests = Vec::new();
            if let Some(ids) = PENDING_BY_ISSUER.may_load(deps.storage, &addr)? {
                for id in ids.iter().rev() {
                    if requests.len() >= MAX_REQUESTS_RETURNED {
                        break;
                    }
                    if let Some(request) = REQUESTS.may_load(deps.storage, id)? {
                        if matches!(request.status, RequestStatus::Pending) {
                            requests.push(request);
                        }
                    }
                }
            }
            to_json_binary(&PendingRequestsResponse { requests })
        }
        QueryMsg::IssuedBySubject { subject } => {
            let addr = deps.api.addr_validate(&subject)?;
            let mut creds = Vec::new();
            if let Some(ids) = ISSUED_BY_SUBJECT.may_load(deps.storage, &addr)? {
                for id in ids.iter().rev() {
                    if creds.len() >= MAX_REQUESTS_RETURNED {
                        break;
                    }
                    if let Some(record) = ISSUED.may_load(deps.storage, id)? {
                        creds.push(record);
                    }
                }
            }
            to_json_binary(&IssuedResponse { credentials: creds })
        }
    }
}

fn execute_register_issuer(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    issuer: String,
    policy_id: String,
    daily_limit: u64,
) -> Result<Response<AuraMsg>, ContractError> {
    let config = CONFIG.load(deps.storage)?;
    if info.sender != config.admin {
        return Err(ContractError::Unauthorized);
    }
    let issuer_addr = deps.api.addr_validate(&issuer)?;
    if ISSUERS.may_load(deps.storage, &issuer_addr)?.is_some() {
        return Err(ContractError::IssuerExists);
    }
    let profile = IssuerProfile {
        address: issuer_addr.clone(),
        policy_id,
        daily_limit,
        active: true,
        registered_at: env.block.time.seconds(),
    };
    ISSUERS.save(deps.storage, &issuer_addr, &profile)?;
    Ok(Response::new()
        .add_attribute("action", "register_issuer")
        .add_attribute("issuer", issuer_addr))
}

fn execute_update_issuer_status(
    deps: DepsMut,
    info: MessageInfo,
    issuer: String,
    active: bool,
) -> Result<Response<AuraMsg>, ContractError> {
    let config = CONFIG.load(deps.storage)?;
    if info.sender != config.admin {
        return Err(ContractError::Unauthorized);
    }
    let issuer_addr = deps.api.addr_validate(&issuer)?;
    ISSUERS.update(
        deps.storage,
        &issuer_addr,
        |maybe| -> Result<_, ContractError> {
            let mut profile = maybe.ok_or(ContractError::IssuerNotFound)?;
            profile.active = active;
            Ok(profile)
        },
    )?;
    Ok(Response::new()
        .add_attribute("action", "update_issuer")
        .add_attribute("issuer", issuer_addr)
        .add_attribute("active", active.to_string()))
}

fn execute_request_vc(
    deps: DepsMut,
    env: Env,
    issuer: String,
    subject: String,
    vc_type: String,
    metadata: String,
) -> Result<Response<AuraMsg>, ContractError> {
    let issuer_addr = deps.api.addr_validate(&issuer)?;
    let subject_addr = deps.api.addr_validate(&subject)?;
    let profile = ISSUERS
        .may_load(deps.storage, &issuer_addr)?
        .ok_or(ContractError::IssuerNotFound)?;
    if !profile.active {
        return Err(ContractError::IssuerInactive);
    }

    let request_id = build_request_id(
        &issuer_addr,
        &subject_addr,
        &vc_type,
        env.block.time.seconds(),
    );
    if REQUESTS.has(deps.storage, &request_id) {
        return Err(ContractError::RequestExists);
    }

    let request = IssueRequest {
        id: request_id.clone(),
        issuer: issuer_addr.clone(),
        subject: subject_addr.clone(),
        vc_type,
        metadata,
        created_at: env.block.time.seconds(),
        status: RequestStatus::Pending,
    };

    REQUESTS.save(deps.storage, &request_id, &request)?;
    push_identifier(
        deps.storage,
        &PENDING_BY_ISSUER,
        &issuer_addr,
        request_id.clone(),
    )?;

    Ok(Response::new()
        .add_attribute("action", "request_vc")
        .add_attribute("issuer", issuer_addr)
        .add_attribute("subject", subject_addr)
        .add_attribute("request_id", request_id))
}

fn execute_fulfill_request(
    deps: DepsMut,
    info: MessageInfo,
    env: Env,
    request_id: String,
    credential_base64: String,
) -> Result<Response<AuraMsg>, ContractError> {
    let mut request = REQUESTS
        .may_load(deps.storage, &request_id)?
        .ok_or(ContractError::RequestNotFound)?;

    match request.status {
        RequestStatus::Pending => {}
        RequestStatus::Fulfilled { .. } => return Err(ContractError::RequestAlreadyFulfilled),
    }

    if info.sender != request.issuer {
        return Err(ContractError::Unauthorized);
    }

    let issuer_profile = ISSUERS.load(deps.storage, &request.issuer)?;
    enforce_mint_limit(
        deps.storage,
        &request.issuer,
        env.block.time.seconds(),
        issuer_profile.daily_limit,
    )?;

    let vc_msg = AuraMsg::VCRegistry {
        register_vc: RegisterVCMsg {
            address: request.subject.to_string(),
            vc_base64: credential_base64.clone(),
        },
    };

    let issue_record = IssueRecord {
        vc_id: format!("vc-{}", &request_id),
        issuer: request.issuer.clone(),
        subject: request.subject.clone(),
        vc_type: request.vc_type.clone(),
        metadata: request.metadata.clone(),
        issued_at: env.block.time.seconds(),
    };

    request.status = RequestStatus::Fulfilled {
        vc_id: issue_record.vc_id.clone(),
    };

    REQUESTS.save(deps.storage, &request_id, &request)?;
    ISSUED.save(deps.storage, &issue_record.vc_id, &issue_record)?;
    remove_identifier(
        deps.storage,
        &PENDING_BY_ISSUER,
        &request.issuer,
        &request_id,
    )?;
    push_identifier(
        deps.storage,
        &ISSUED_BY_SUBJECT,
        &request.subject,
        issue_record.vc_id.clone(),
    )?;

    Ok(Response::new()
        .add_message(CosmosMsg::Custom(vc_msg))
        .add_attribute("action", "fulfill_request")
        .add_attribute("request_id", request_id)
        .add_attribute("vc_id", issue_record.vc_id))
}

fn execute_revoke_vc(
    deps: DepsMut,
    info: MessageInfo,
    vc_id: String,
    reason: String,
) -> Result<Response<AuraMsg>, ContractError> {
    let record = ISSUED
        .may_load(deps.storage, &vc_id)?
        .ok_or(ContractError::CredentialNotFound)?;
    let config = CONFIG.load(deps.storage)?;
    if info.sender != record.issuer && info.sender != config.admin {
        return Err(ContractError::Unauthorized);
    }
    Ok(Response::new()
        .add_attribute("action", "revoke_vc")
        .add_attribute("issuer", info.sender)
        .add_attribute("vc_id", vc_id)
        .add_attribute("reason", reason))
}

fn build_request_id(issuer: &Addr, subject: &Addr, vc_type: &str, timestamp: u64) -> String {
    let mut hasher = Sha256::new();
    hasher.update(issuer.as_bytes());
    hasher.update(subject.as_bytes());
    hasher.update(vc_type.as_bytes());
    hasher.update(timestamp.to_be_bytes());
    format!("req-{}", hex::encode(hasher.finalize())[..32].to_string())
}

fn push_identifier(
    storage: &mut dyn cosmwasm_std::Storage,
    map: &Map<&Addr, Vec<String>>,
    owner: &Addr,
    id: String,
) -> StdResult<()> {
    let mut entries = map.may_load(storage, owner)?.unwrap_or_default();
    if !entries.iter().any(|existing| existing == &id) {
        entries.push(id);
    }
    map.save(storage, owner, &entries)
}

fn remove_identifier(
    storage: &mut dyn cosmwasm_std::Storage,
    map: &Map<&Addr, Vec<String>>,
    owner: &Addr,
    id: &str,
) -> StdResult<()> {
    if let Some(mut entries) = map.may_load(storage, owner)? {
        entries.retain(|existing| existing != id);
        if entries.is_empty() {
            map.remove(storage, owner);
        } else {
            map.save(storage, owner, &entries)?;
        }
    }
    Ok(())
}

fn enforce_mint_limit(
    storage: &mut dyn cosmwasm_std::Storage,
    issuer: &Addr,
    timestamp: u64,
    limit: u64,
) -> Result<(), ContractError> {
    if limit == 0 {
        return Ok(());
    }
    let day = timestamp / SECONDS_PER_DAY;
    let count = MINT_COUNTERS.may_load(storage, (issuer, day))?.unwrap_or(0);
    if (count as u64) >= limit {
        return Err(ContractError::MintLimitReached);
    }
    MINT_COUNTERS.save(storage, (issuer, day), &(count + 1))?;
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    use cosmwasm_std::testing::{mock_dependencies, mock_env, mock_info};

    fn instantiate_with_admin(deps: DepsMut, admin: &str) {
        instantiate(
            deps,
            mock_env(),
            mock_info(admin, &[]),
            InstantiateMsg {
                admin: Some(admin.to_string()),
            },
        )
        .unwrap();
    }

    fn register_default_issuer(deps: DepsMut, env: Env, admin: &str, limit: u64) {
        execute(
            deps,
            env,
            mock_info(admin, &[]),
            ExecuteMsg::RegisterIssuer {
                issuer: "issuer".into(),
                policy_id: "policy".into(),
                daily_limit: limit,
            },
        )
        .unwrap();
    }

    #[test]
    fn register_enforces_admin() {
        let mut deps = mock_dependencies();
        instantiate_with_admin(deps.as_mut(), "admin");

        let err = execute(
            deps.as_mut(),
            mock_env(),
            mock_info("intruder", &[]),
            ExecuteMsg::RegisterIssuer {
                issuer: "issuer".into(),
                policy_id: "policy".into(),
                daily_limit: 5,
            },
        )
        .unwrap_err();
        assert_eq!(err, ContractError::Unauthorized);

        register_default_issuer(deps.as_mut(), mock_env(), "admin", 5);
        let issuer = ISSUERS
            .load(
                deps.as_ref().storage,
                &deps.api.addr_validate("issuer").unwrap(),
            )
            .unwrap();
        assert_eq!(issuer.daily_limit, 5);
    }

    #[test]
    fn request_records_pending() {
        let mut deps = mock_dependencies();
        instantiate_with_admin(deps.as_mut(), "admin");
        register_default_issuer(deps.as_mut(), mock_env(), "admin", 5);

        let resp = execute(
            deps.as_mut(),
            mock_env(),
            mock_info("user", &[]),
            ExecuteMsg::RequestVc {
                issuer: "issuer".into(),
                subject: "holder".into(),
                vc_type: "kyc".into(),
                metadata: "{}".into(),
            },
        )
        .unwrap();
        let req_id = resp
            .attributes
            .iter()
            .find(|a| a.key == "request_id")
            .unwrap()
            .value
            .clone();
        let pending = PENDING_BY_ISSUER
            .may_load(
                deps.as_ref().storage,
                &deps.api.addr_validate("issuer").unwrap(),
            )
            .unwrap()
            .unwrap();
        assert!(pending.contains(&req_id));

        let query_resp: PendingRequestsResponse = cosmwasm_std::from_binary(
            &query(
                deps.as_ref(),
                mock_env(),
                QueryMsg::PendingRequests {
                    issuer: "issuer".into(),
                },
            )
            .unwrap(),
        )
        .unwrap();
        assert_eq!(query_resp.requests.len(), 1);
        assert_eq!(query_resp.requests[0].id, req_id);
    }

    #[test]
    fn fulfill_respects_daily_limit() {
        let mut deps = mock_dependencies();
        instantiate_with_admin(deps.as_mut(), "admin");
        register_default_issuer(deps.as_mut(), mock_env(), "admin", 1);

        let mut env = mock_env();
        let first = execute(
            deps.as_mut(),
            env.clone(),
            mock_info("req1", &[]),
            ExecuteMsg::RequestVc {
                issuer: "issuer".into(),
                subject: "holder1".into(),
                vc_type: "kyc".into(),
                metadata: "{}".into(),
            },
        )
        .unwrap()
        .attributes
        .iter()
        .find(|a| a.key == "request_id")
        .unwrap()
        .value
        .clone();

        env.block.time = env.block.time.plus_seconds(5);
        let second = execute(
            deps.as_mut(),
            env.clone(),
            mock_info("req2", &[]),
            ExecuteMsg::RequestVc {
                issuer: "issuer".into(),
                subject: "holder2".into(),
                vc_type: "kyc".into(),
                metadata: "{}".into(),
            },
        )
        .unwrap()
        .attributes
        .iter()
        .find(|a| a.key == "request_id")
        .unwrap()
        .value
        .clone();

        execute(
            deps.as_mut(),
            env.clone(),
            mock_info("issuer", &[]),
            ExecuteMsg::FulfillRequest {
                request_id: first,
                credential_base64: "ZGF0YQ==".into(),
            },
        )
        .unwrap();

        let err = execute(
            deps.as_mut(),
            env,
            mock_info("issuer", &[]),
            ExecuteMsg::FulfillRequest {
                request_id: second,
                credential_base64: "ZGF0YQ==".into(),
            },
        )
        .unwrap_err();
        assert_eq!(err, ContractError::MintLimitReached);
    }
}
