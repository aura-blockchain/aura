// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import pb "github.com/aequitas/aura/proto/aura/dataregistry/v1beta1"

// Re-export all proto types
type (
	// Enums
	DataItemType      = pb.DataItemType
	DataItemStatus    = pb.DataItemStatus
	VerificationLevel = pb.VerificationLevel
	AccessMode        = pb.AccessMode

	// Core types
	DataItem                = pb.DataItem
	AccessPolicy            = pb.AccessPolicy
	GeoLocation             = pb.GeoLocation
	VehicleRegistrationData = pb.VehicleRegistrationData
	PhotoData               = pb.PhotoData
	GolfScoreData           = pb.GolfScoreData
	Verification            = pb.Verification
	Params                  = pb.Params

	// Message types
	MsgStoreDataItem          = pb.MsgStoreDataItem
	MsgStoreDataItemResponse  = pb.MsgStoreDataItemResponse
	MsgUpdateDataItem         = pb.MsgUpdateDataItem
	MsgUpdateDataItemResponse = pb.MsgUpdateDataItemResponse
	MsgDeleteDataItem         = pb.MsgDeleteDataItem
	MsgDeleteDataItemResponse = pb.MsgDeleteDataItemResponse
	MsgVerifyDataItem         = pb.MsgVerifyDataItem
	MsgVerifyDataItemResponse = pb.MsgVerifyDataItemResponse
	MsgRevokeDataItem         = pb.MsgRevokeDataItem
	MsgRevokeDataItemResponse = pb.MsgRevokeDataItemResponse

	// Query types
	QueryDataItemRequest               = pb.QueryDataItemRequest
	QueryDataItemResponse              = pb.QueryDataItemResponse
	QueryUserDataItemsRequest          = pb.QueryUserDataItemsRequest
	QueryUserDataItemsResponse         = pb.QueryUserDataItemsResponse
	QuerySearchDataItemsRequest        = pb.QuerySearchDataItemsRequest
	QuerySearchDataItemsResponse       = pb.QuerySearchDataItemsResponse
	QueryDataItemVerificationsRequest  = pb.QueryDataItemVerificationsRequest
	QueryDataItemVerificationsResponse = pb.QueryDataItemVerificationsResponse
	QueryStatsRequest                  = pb.QueryStatsRequest
	QueryStatsResponse                 = pb.QueryStatsResponse
	QueryParamsRequest                 = pb.QueryParamsRequest
	QueryParamsResponse                = pb.QueryParamsResponse

	// Genesis types
	GenesisState = pb.GenesisState

	// Event types
	EventDataItemStored   = pb.EventDataItemStored
	EventDataItemUpdated  = pb.EventDataItemUpdated
	EventDataItemDeleted  = pb.EventDataItemDeleted
	EventDataItemVerified = pb.EventDataItemVerified
	EventDataItemRevoked  = pb.EventDataItemRevoked
)

// Re-export enum values for DataItemType
const (
	DataItemType_DATA_ITEM_TYPE_UNSPECIFIED          = pb.DataItemType_DATA_ITEM_TYPE_UNSPECIFIED
	DataItemType_DATA_ITEM_TYPE_VEHICLE_REGISTRATION = pb.DataItemType_DATA_ITEM_TYPE_VEHICLE_REGISTRATION
	DataItemType_DATA_ITEM_TYPE_VEHICLE_INSURANCE    = pb.DataItemType_DATA_ITEM_TYPE_VEHICLE_INSURANCE
	DataItemType_DATA_ITEM_TYPE_PROPERTY_DEED        = pb.DataItemType_DATA_ITEM_TYPE_PROPERTY_DEED
	DataItemType_DATA_ITEM_TYPE_LEASE_AGREEMENT      = pb.DataItemType_DATA_ITEM_TYPE_LEASE_AGREEMENT
	DataItemType_DATA_ITEM_TYPE_CONTRACT             = pb.DataItemType_DATA_ITEM_TYPE_CONTRACT
	DataItemType_DATA_ITEM_TYPE_RECEIPT              = pb.DataItemType_DATA_ITEM_TYPE_RECEIPT
	DataItemType_DATA_ITEM_TYPE_WARRANTY             = pb.DataItemType_DATA_ITEM_TYPE_WARRANTY
	DataItemType_DATA_ITEM_TYPE_PHOTO                = pb.DataItemType_DATA_ITEM_TYPE_PHOTO
	DataItemType_DATA_ITEM_TYPE_VIDEO                = pb.DataItemType_DATA_ITEM_TYPE_VIDEO
	DataItemType_DATA_ITEM_TYPE_AUDIO                = pb.DataItemType_DATA_ITEM_TYPE_AUDIO
	DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF         = pb.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF
	DataItemType_DATA_ITEM_TYPE_GOLF_SCORE           = pb.DataItemType_DATA_ITEM_TYPE_GOLF_SCORE
	DataItemType_DATA_ITEM_TYPE_TEST_SCORE           = pb.DataItemType_DATA_ITEM_TYPE_TEST_SCORE
	DataItemType_DATA_ITEM_TYPE_CERTIFICATION        = pb.DataItemType_DATA_ITEM_TYPE_CERTIFICATION
	DataItemType_DATA_ITEM_TYPE_ACHIEVEMENT          = pb.DataItemType_DATA_ITEM_TYPE_ACHIEVEMENT
	DataItemType_DATA_ITEM_TYPE_NFT                  = pb.DataItemType_DATA_ITEM_TYPE_NFT
	DataItemType_DATA_ITEM_TYPE_DIGITAL_ART          = pb.DataItemType_DATA_ITEM_TYPE_DIGITAL_ART
	DataItemType_DATA_ITEM_TYPE_MUSIC_LICENSE        = pb.DataItemType_DATA_ITEM_TYPE_MUSIC_LICENSE
	DataItemType_DATA_ITEM_TYPE_VACCINATION_RECORD   = pb.DataItemType_DATA_ITEM_TYPE_VACCINATION_RECORD
	DataItemType_DATA_ITEM_TYPE_MEDICAL_RECORD       = pb.DataItemType_DATA_ITEM_TYPE_MEDICAL_RECORD
	DataItemType_DATA_ITEM_TYPE_PRESCRIPTION         = pb.DataItemType_DATA_ITEM_TYPE_PRESCRIPTION
	DataItemType_DATA_ITEM_TYPE_CUSTOM               = pb.DataItemType_DATA_ITEM_TYPE_CUSTOM
)

// Re-export enum values for DataItemStatus
const (
	DataItemStatus_DATA_ITEM_STATUS_UNSPECIFIED          = pb.DataItemStatus_DATA_ITEM_STATUS_UNSPECIFIED
	DataItemStatus_DATA_ITEM_STATUS_PENDING_VERIFICATION = pb.DataItemStatus_DATA_ITEM_STATUS_PENDING_VERIFICATION
	DataItemStatus_DATA_ITEM_STATUS_VERIFIED             = pb.DataItemStatus_DATA_ITEM_STATUS_VERIFIED
	DataItemStatus_DATA_ITEM_STATUS_REJECTED             = pb.DataItemStatus_DATA_ITEM_STATUS_REJECTED
	DataItemStatus_DATA_ITEM_STATUS_EXPIRED              = pb.DataItemStatus_DATA_ITEM_STATUS_EXPIRED
	DataItemStatus_DATA_ITEM_STATUS_REVOKED              = pb.DataItemStatus_DATA_ITEM_STATUS_REVOKED
)

// Re-export enum values for VerificationLevel
const (
	VerificationLevel_VERIFICATION_LEVEL_UNSPECIFIED         = pb.VerificationLevel_VERIFICATION_LEVEL_UNSPECIFIED
	VerificationLevel_VERIFICATION_LEVEL_SELF_ATTESTED       = pb.VerificationLevel_VERIFICATION_LEVEL_SELF_ATTESTED
	VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED       = pb.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED
	VerificationLevel_VERIFICATION_LEVEL_AI_VERIFIED         = pb.VerificationLevel_VERIFICATION_LEVEL_AI_VERIFIED
	VerificationLevel_VERIFICATION_LEVEL_AUTHORITY_VERIFIED  = pb.VerificationLevel_VERIFICATION_LEVEL_AUTHORITY_VERIFIED
	VerificationLevel_VERIFICATION_LEVEL_BLOCKCHAIN_ANCHORED = pb.VerificationLevel_VERIFICATION_LEVEL_BLOCKCHAIN_ANCHORED
)

// Re-export enum values for AccessMode
const (
	AccessMode_ACCESS_MODE_PRIVATE        = pb.AccessMode_ACCESS_MODE_PRIVATE
	AccessMode_ACCESS_MODE_WHITELIST      = pb.AccessMode_ACCESS_MODE_WHITELIST
	AccessMode_ACCESS_MODE_PUBLIC         = pb.AccessMode_ACCESS_MODE_PUBLIC
	AccessMode_ACCESS_MODE_VERIFIED_USERS = pb.AccessMode_ACCESS_MODE_VERIFIED_USERS
)
