use cosmwasm_schema::cw_serde;
use cosmwasm_std::Addr;
use cw_storage_plus::{Item, Map};

#[cw_serde]
pub struct Config {
    pub admin: Addr,
}

#[cw_serde]
pub struct IssuerProfile {
    pub address: Addr,
    pub policy_id: String,
    pub daily_limit: u64,
    pub active: bool,
    pub registered_at: u64,
}

#[cw_serde]
pub enum RequestStatus {
    Pending,
    Fulfilled { vc_id: String },
}

#[cw_serde]
pub struct IssueRequest {
    pub id: String,
    pub issuer: Addr,
    pub subject: Addr,
    pub vc_type: String,
    pub metadata: String,
    pub created_at: u64,
    pub status: RequestStatus,
}

#[cw_serde]
pub struct IssueRecord {
    pub vc_id: String,
    pub issuer: Addr,
    pub subject: Addr,
    pub vc_type: String,
    pub metadata: String,
    pub issued_at: u64,
}

pub const CONFIG: Item<Config> = Item::new("config");
pub const ISSUERS: Map<&Addr, IssuerProfile> = Map::new("issuers");
pub const REQUESTS: Map<&str, IssueRequest> = Map::new("requests");
pub const ISSUED: Map<&str, IssueRecord> = Map::new("issued");
pub const PENDING_BY_ISSUER: Map<&Addr, Vec<String>> = Map::new("pending_by_issuer");
pub const ISSUED_BY_SUBJECT: Map<&Addr, Vec<String>> = Map::new("issued_by_subject");
pub const MINT_COUNTERS: Map<(&Addr, u64), u32> = Map::new("mint_counters");
