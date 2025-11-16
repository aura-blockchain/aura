package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	dataregistrypb "github.com/aequitas/aura/proto/aura/dataregistry/v1beta1"
)

// GetQueryCmd returns the query commands for this module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "dataregistry",
		Short:                      "Querying commands for the Data Registry module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdShowDataItem(),
		CmdListDataItems(),
		CmdSearchDataItems(),
		CmdGetStats(),
		CmdGetParams(),
	)

	return cmd
}

// CmdShowDataItem shows a specific data item
func CmdShowDataItem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-data-item [data-id]",
		Short: "Show details of a specific data item",
		Long: `Show full details of a data item including metadata, verifications, and access policy.

Examples:
  aurad query dataregistry show-data-item data:abc123...
  aurad query dataregistry show-data-item data:abc123... --requester aura1...
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := dataregistrypb.NewQueryClient(clientCtx)

			// Get requester address from flag or use client address
			requester, _ := cmd.Flags().GetString("requester")
			if requester == "" && clientCtx.GetFromAddress() != nil {
				requester = clientCtx.GetFromAddress().String()
			}

			req := &dataregistrypb.QueryDataItemRequest{
				DataId:    args[0],
				Requester: requester,
			}

			res, err := queryClient.DataItem(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String("requester", "", "Requester address (optional)")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdListDataItems lists user's data items
func CmdListDataItems() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-data-items [owner-address]",
		Short: "List data items owned by a user",
		Long: `List all data items owned by a specific address, with optional filters.

Examples:
  aurad query dataregistry list-data-items aura1abc...
  aurad query dataregistry list-data-items aura1abc... --type photo
  aurad query dataregistry list-data-items aura1abc... --status verified
  aurad query dataregistry list-data-items $(aurad keys show alice -a) --type document --status verified
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := dataregistrypb.NewQueryClient(clientCtx)

			// Parse optional filters
			typeFilter, _ := cmd.Flags().GetString("type")
			statusFilter, _ := cmd.Flags().GetString("status")

			var dataType dataregistrypb.DataItemType
			if typeFilter != "" {
				dataType = parseDataTypeProto(typeFilter)
				if dataType == dataregistrypb.DataItemType_DATA_ITEM_TYPE_UNSPECIFIED {
					return fmt.Errorf("invalid data type: %s", typeFilter)
				}
			}

			var status dataregistrypb.DataItemStatus
			if statusFilter != "" {
				status = parseDataItemStatusProto(statusFilter)
				if status == dataregistrypb.DataItemStatus_DATA_ITEM_STATUS_UNSPECIFIED {
					return fmt.Errorf("invalid status: %s", statusFilter)
				}
			}

			req := &dataregistrypb.QueryUserDataItemsRequest{
				OwnerAddress: args[0],
				TypeFilter:   dataType,
				StatusFilter: status,
			}

			res, err := queryClient.UserDataItems(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String("type", "", "Filter by data type (photo, video, document, etc.)")
	cmd.Flags().String("status", "", "Filter by status (pending_verification, verified, revoked)")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdSearchDataItems searches for data items
func CmdSearchDataItems() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search-data-items [query-json]",
		Short: "Search for data items with filters",
		Long: `Search for data items using various filters including tags, type, and location.

Examples:
  aurad query dataregistry search-data-items '{"tags":["vacation","2024"]}'
  aurad query dataregistry search-data-items '{"type":"photo","tags":["beach"]}' --requester aura1...
  aurad query dataregistry search-data-items '{"geo_location":{"latitude":37.7749,"longitude":-122.4194},"radius_km":10}' --requester aura1...

Query JSON fields:
  - query: text query (string)
  - tags: array of tags to match
  - type: data item type filter
  - geo_location: {latitude, longitude} for geo search
  - radius_km: radius for geo search (in kilometers)
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := dataregistrypb.NewQueryClient(clientCtx)

			// Parse query JSON
			var queryMap map[string]interface{}
			if err := json.Unmarshal([]byte(args[0]), &queryMap); err != nil {
				return fmt.Errorf("invalid query JSON: %w", err)
			}

			// Get requester address from flag or use client address
			requester, _ := cmd.Flags().GetString("requester")
			if requester == "" && clientCtx.GetFromAddress() != nil {
				requester = clientCtx.GetFromAddress().String()
			}

			// Extract query fields
			query := ""
			if q, ok := queryMap["query"].(string); ok {
				query = q
			}

			var tags []string
			if t, ok := queryMap["tags"].([]interface{}); ok {
				for _, tag := range t {
					if tagStr, ok := tag.(string); ok {
						tags = append(tags, tagStr)
					}
				}
			}

			var typeFilter dataregistrypb.DataItemType
			if t, ok := queryMap["type"].(string); ok {
				typeFilter = parseDataTypeProto(t)
			}

			var geoLocation *dataregistrypb.GeoLocation
			var radiusKM float64
			if gl, ok := queryMap["geo_location"].(map[string]interface{}); ok {
				geoLocation = &dataregistrypb.GeoLocation{
					Latitude:  gl["latitude"].(float64),
					Longitude: gl["longitude"].(float64),
				}
				if r, ok := queryMap["radius_km"].(float64); ok {
					radiusKM = r
				}
			}

			req := &dataregistrypb.QuerySearchDataItemsRequest{
				SearchQuery:  query,
				Tags:         tags,
				TypeFilter:   typeFilter,
				NearLocation: geoLocation,
				RadiusKm:     radiusKM,
				Requester:    requester,
			}

			res, err := queryClient.SearchDataItems(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String("requester", "", "Requester address (optional)")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdGetStats returns registry statistics
func CmdGetStats() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Get Data Registry statistics",
		Long: `Get statistics about the Data Registry including total items, verifications, and breakdown by type.

Examples:
  aurad query dataregistry stats
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := dataregistrypb.NewQueryClient(clientCtx)

			req := &dataregistrypb.QueryStatsRequest{}

			res, err := queryClient.Stats(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdGetParams returns the module parameters
func CmdGetParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Get Data Registry module parameters",
		Long: `Get the Data Registry module parameters including limits and rewards.

Examples:
  aurad query dataregistry params
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := dataregistrypb.NewQueryClient(clientCtx)

			req := &dataregistrypb.QueryParamsRequest{}

			res, err := queryClient.Params(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// parseDataTypeProto converts a string to DataItemType enum from proto
func parseDataTypeProto(s string) dataregistrypb.DataItemType {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "_")

	// Map to proto-generated enum values
	typeMap := map[string]dataregistrypb.DataItemType{
		"photo":                dataregistrypb.DataItemType_DATA_ITEM_TYPE_PHOTO,
		"video":                dataregistrypb.DataItemType_DATA_ITEM_TYPE_VIDEO,
		"audio":                dataregistrypb.DataItemType_DATA_ITEM_TYPE_AUDIO,
		"document":             dataregistrypb.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
		"document_pdf":         dataregistrypb.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
		"contract":             dataregistrypb.DataItemType_DATA_ITEM_TYPE_CONTRACT,
		"receipt":              dataregistrypb.DataItemType_DATA_ITEM_TYPE_RECEIPT,
		"vehicle_registration": dataregistrypb.DataItemType_DATA_ITEM_TYPE_VEHICLE_REGISTRATION,
		"vehicle_insurance":    dataregistrypb.DataItemType_DATA_ITEM_TYPE_VEHICLE_INSURANCE,
		"property_deed":        dataregistrypb.DataItemType_DATA_ITEM_TYPE_PROPERTY_DEED,
		"lease_agreement":      dataregistrypb.DataItemType_DATA_ITEM_TYPE_LEASE_AGREEMENT,
		"warranty":             dataregistrypb.DataItemType_DATA_ITEM_TYPE_WARRANTY,
		"golf_score":           dataregistrypb.DataItemType_DATA_ITEM_TYPE_GOLF_SCORE,
		"test_score":           dataregistrypb.DataItemType_DATA_ITEM_TYPE_TEST_SCORE,
		"certification":        dataregistrypb.DataItemType_DATA_ITEM_TYPE_CERTIFICATION,
		"achievement":          dataregistrypb.DataItemType_DATA_ITEM_TYPE_ACHIEVEMENT,
		"nft":                  dataregistrypb.DataItemType_DATA_ITEM_TYPE_NFT,
		"digital_art":          dataregistrypb.DataItemType_DATA_ITEM_TYPE_DIGITAL_ART,
		"music_license":        dataregistrypb.DataItemType_DATA_ITEM_TYPE_MUSIC_LICENSE,
		"vaccination_record":   dataregistrypb.DataItemType_DATA_ITEM_TYPE_VACCINATION_RECORD,
		"medical_record":       dataregistrypb.DataItemType_DATA_ITEM_TYPE_MEDICAL_RECORD,
		"prescription":         dataregistrypb.DataItemType_DATA_ITEM_TYPE_PRESCRIPTION,
		"custom":               dataregistrypb.DataItemType_DATA_ITEM_TYPE_CUSTOM,
	}

	if dataType, ok := typeMap[s]; ok {
		return dataType
	}

	return dataregistrypb.DataItemType_DATA_ITEM_TYPE_UNSPECIFIED
}

// parseDataItemStatusProto converts a string to DataItemStatus enum from proto
func parseDataItemStatusProto(s string) dataregistrypb.DataItemStatus {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "_")

	statusMap := map[string]dataregistrypb.DataItemStatus{
		"pending_verification": dataregistrypb.DataItemStatus_DATA_ITEM_STATUS_PENDING_VERIFICATION,
		"verified":             dataregistrypb.DataItemStatus_DATA_ITEM_STATUS_VERIFIED,
		"rejected":             dataregistrypb.DataItemStatus_DATA_ITEM_STATUS_REJECTED,
		"expired":              dataregistrypb.DataItemStatus_DATA_ITEM_STATUS_EXPIRED,
		"revoked":              dataregistrypb.DataItemStatus_DATA_ITEM_STATUS_REVOKED,
	}

	if status, ok := statusMap[s]; ok {
		return status
	}

	return dataregistrypb.DataItemStatus_DATA_ITEM_STATUS_UNSPECIFIED
}
