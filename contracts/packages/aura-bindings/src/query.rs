use cosmwasm_std::CustomQuery;
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum AuraQuery {
    #[serde(rename = "vc_registry")]
    VCRegistry { get_vc: GetVCQuery },
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct GetVCQuery {
    pub address: String,
}

impl CustomQuery for AuraQuery {}
