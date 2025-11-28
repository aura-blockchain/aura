use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum AuraMsg {
    #[serde(rename = "vc_registry")]
    VCRegistry { register_vc: RegisterVCMsg },
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct RegisterVCMsg {
    pub address: String,
    pub vc_base64: String,
}
