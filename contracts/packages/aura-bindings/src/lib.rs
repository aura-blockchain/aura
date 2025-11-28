pub mod msg;
pub mod query;

use cosmwasm_std::CustomMsg;

impl CustomMsg for msg::AuraMsg {}
pub mod types;

pub use msg::AuraMsg;
pub use query::{AuraQuery, GetVCQuery};
pub use types::VerifiableCredential;

#[cfg(test)]
mod tests {
    use super::*;
    use cosmwasm_std::to_json_vec;

    #[test]
    fn vc_registry_message_serializes_as_expected() {
        let msg = AuraMsg::VCRegistry {
            register_vc: msg::RegisterVCMsg {
                address: "aura1holder".into(),
                vc_base64: "dGVzdA==".into(),
            },
        };

        let serialized = to_json_vec(&msg).expect("json");
        assert_eq!(
            serde_json::from_slice::<serde_json::Value>(&serialized).unwrap(),
            serde_json::json!({
                "vc_registry": {
                    "register_vc": {
                        "address": "aura1holder",
                        "vc_base64": "dGVzdA=="
                    }
                }
            })
        );
    }

    #[test]
    fn vc_registry_query_serializes_as_expected() {
        let query = AuraQuery::VCRegistry {
            get_vc: GetVCQuery {
                address: "aura1holder".into(),
            },
        };
        let serialized = to_json_vec(&query).expect("json");
        assert_eq!(
            serde_json::from_slice::<serde_json::Value>(&serialized).unwrap(),
            serde_json::json!({
                "vc_registry": {
                    "get_vc": {
                        "address": "aura1holder"
                    }
                }
            })
        );
    }
}
