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

    #[error("Metadata exceeds maximum length of {0} bytes")]
    MetadataTooLarge(usize),

    #[error("Credential exceeds maximum length of {0} bytes")]
    CredentialTooLarge(usize),

    #[error("Reason exceeds maximum length of {0} bytes")]
    ReasonTooLarge(usize),

    #[error("Policy ID exceeds maximum length of {0} bytes")]
    PolicyIdTooLarge(usize),

    #[error("VC type exceeds maximum length of {0} bytes")]
    VcTypeTooLarge(usize),
}
