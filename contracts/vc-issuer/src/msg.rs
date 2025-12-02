use cosmwasm_schema::{cw_serde, QueryResponses};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use crate::state::{IssueRecord, IssueRequest};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct InstantiateMsg {
    pub admin: Option<String>,
}

#[cw_serde]
pub enum ExecuteMsg {
    RegisterIssuer {
        issuer: String,
        policy_id: String,
        daily_limit: u64,
    },
    UpdateIssuerStatus {
        issuer: String,
        active: bool,
    },
    RequestVc {
        issuer: String,
        subject: String,
        vc_type: String,
        metadata: String,
    },
    FulfillRequest {
        request_id: String,
        credential_base64: String,
    },
    RevokeVc {
        vc_id: String,
        reason: String,
    },
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    #[returns(IssuerProfile)]
    Issuer { address: String },
    #[returns(PendingRequestsResponse)]
    PendingRequests { issuer: String },
    #[returns(IssuedResponse)]
    IssuedBySubject { subject: String },
}

#[cw_serde]
pub struct PendingRequestsResponse {
    pub requests: Vec<IssueRequest>,
}

#[cw_serde]
pub struct IssuedResponse {
    pub credentials: Vec<IssueRecord>,
}
