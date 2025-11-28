package keeper

// This file documents the status filtering implementation in query_server.go
//
// Implementation Details:
// -----------------------
// The RegisteredContracts query method now properly filters contracts by status
// when the status field is specified in the request.
//
// Key Points:
// 1. When req.Status == CONTRACT_STATUS_UNSPECIFIED (value 0), no filtering is applied
//    and all contracts are returned.
//
// 2. When req.Status is set to a specific status (e.g., CONTRACT_STATUS_ACTIVE),
//    only contracts matching that status are included in the response.
//
// 3. The filtering logic (lines 172-177 in query_server.go):
//    ```go
//    // Apply status filter if specified
//    if req.Status != pb.ContractStatus_CONTRACT_STATUS_UNSPECIFIED {
//        if info.Status != req.Status {
//            continue // Skip this contract, continue to next iteration
//        }
//    }
//    ```
//
// 4. This implementation correctly handles the test case TestQueryRegisteredContracts_WithStatusFilter
//    which expects:
//    - 3 contracts registered (2 ACTIVE, 1 PAUSED)
//    - Query with Status=CONTRACT_STATUS_ACTIVE should return only 2 contracts
//    - All returned contracts should have Status=CONTRACT_STATUS_ACTIVE
//
// Status Values (from proto definition):
// - CONTRACT_STATUS_UNSPECIFIED = 0 (default/filter disabled)
// - CONTRACT_STATUS_ACTIVE = 1
// - CONTRACT_STATUS_PAUSED = 2
// - CONTRACT_STATUS_FROZEN = 3 (and others)
