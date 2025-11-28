use cosmwasm_std::StdError;
use thiserror::Error;

#[derive(Error, Debug, PartialEq)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("unauthorized")]
    Unauthorized,

    #[error("issuer already registered")]
    IssuerExists,

    #[error("issuer not found")]
    IssuerNotFound,

    #[error("request not found")]
    RequestNotFound,

    #[error("request already fulfilled")]
    RequestAlreadyFulfilled,

    #[error("issuer inactive")]
    IssuerInactive,

    #[error("request already exists")]
    RequestExists,

    #[error("mint limit reached for issuer")]
    MintLimitReached,

    #[error("credential not found")]
    CredentialNotFound,
}
