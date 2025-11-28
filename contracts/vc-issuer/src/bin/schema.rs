use std::fs::create_dir_all;

use cosmwasm_schema::{export_schema, remove_schemas, schema_for};

use vc_issuer::msg::{
    ExecuteMsg, InstantiateMsg, IssuedResponse, PendingRequestsResponse, QueryMsg,
};

fn main() {
    let out_dir = std::path::Path::new("schema");
    create_dir_all(&out_dir).unwrap();
    remove_schemas(&out_dir).unwrap();

    export_schema(&schema_for!(InstantiateMsg), &out_dir);
    export_schema(&schema_for!(ExecuteMsg), &out_dir);
    export_schema(&schema_for!(QueryMsg), &out_dir);
    export_schema(&schema_for!(PendingRequestsResponse), &out_dir);
    export_schema(&schema_for!(IssuedResponse), &out_dir);
}
