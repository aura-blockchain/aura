// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/governance/types"
)

func setupTemplateKeeper(t *testing.T) (*Keeper, sdk.Context) {
	input := keepertest.CreateTestInputWithKeys(t, "governance")
	mockStaking := NewMockStakingKeeper()
	mockBank := &MockBankKeeper{
		balances:       make(map[string]sdk.Coins),
		moduleBalances: make(map[string]sdk.Coins),
	}
	mockSecurity := &MockSecurityKeeper{}
	keeper := NewKeeper(input.Cdc, input.StoreKey, mockStaking, mockBank, mockSecurity)
	ctx := input.Ctx.WithKVGasConfig(storetypes.GasConfig{})
	keeper.SetParams(ctx, types.DefaultParams())
	return keeper, ctx
}

func TestGetProposalTemplate_TextProposal(t *testing.T) {
	keeper, _ := setupTemplateKeeper(t)

	template := keeper.GetProposalTemplate(types.CategoryText)

	require.NotNil(t, template)
	require.Equal(t, types.CategoryText, template.Category)
	require.Equal(t, "Text Proposal", template.Name)
	require.NotEmpty(t, template.Description)
	require.NotEmpty(t, template.TitleTemplate)
	require.NotEmpty(t, template.DescriptionTemplate)
	require.Contains(t, template.RequiredFields, "title")
	require.Contains(t, template.RequiredFields, "description")
	require.NotEmpty(t, template.Examples)
}

func TestGetProposalTemplate_ParameterChange(t *testing.T) {
	keeper, _ := setupTemplateKeeper(t)

	template := keeper.GetProposalTemplate(types.CategoryParameterChange)

	require.NotNil(t, template)
	require.Equal(t, types.CategoryParameterChange, template.Category)
	require.Equal(t, "Parameter Change Proposal", template.Name)
	require.NotEmpty(t, template.Description)
	require.NotEmpty(t, template.TitleTemplate)
	require.NotEmpty(t, template.DescriptionTemplate)
	require.NotEmpty(t, template.ContentTemplate)
	require.Contains(t, template.RequiredFields, "title")
	require.Contains(t, template.RequiredFields, "description")
	require.Contains(t, template.RequiredFields, "content")
	require.NotEmpty(t, template.Examples)
	require.Contains(t, template.ContentTemplate, "changes")
	require.Contains(t, template.ContentTemplate, "module")
	require.Contains(t, template.ContentTemplate, "parameter")
}

func TestGetProposalTemplate_SoftwareUpgrade(t *testing.T) {
	keeper, _ := setupTemplateKeeper(t)

	template := keeper.GetProposalTemplate(types.CategorySoftwareUpgrade)

	require.NotNil(t, template)
	require.Equal(t, types.CategorySoftwareUpgrade, template.Category)
	require.Equal(t, "Software Upgrade Proposal", template.Name)
	require.NotEmpty(t, template.Description)
	require.NotEmpty(t, template.TitleTemplate)
	require.NotEmpty(t, template.DescriptionTemplate)
	require.NotEmpty(t, template.ContentTemplate)
	require.Contains(t, template.RequiredFields, "title")
	require.Contains(t, template.RequiredFields, "description")
	require.Contains(t, template.RequiredFields, "content")
	require.NotEmpty(t, template.Examples)
	require.Contains(t, template.ContentTemplate, "name")
	require.Contains(t, template.ContentTemplate, "height")
	require.Contains(t, template.ContentTemplate, "version")
}

func TestGetProposalTemplate_CommunitySpend(t *testing.T) {
	keeper, _ := setupTemplateKeeper(t)

	template := keeper.GetProposalTemplate(types.CategorySpending)

	require.NotNil(t, template)
	require.Equal(t, types.CategorySpending, template.Category)
	require.Equal(t, "Community Spend Proposal", template.Name)
	require.NotEmpty(t, template.Description)
	require.NotEmpty(t, template.TitleTemplate)
	require.NotEmpty(t, template.DescriptionTemplate)
	require.NotEmpty(t, template.ContentTemplate)
	require.Contains(t, template.RequiredFields, "title")
	require.Contains(t, template.RequiredFields, "description")
	require.Contains(t, template.RequiredFields, "content")
	require.NotEmpty(t, template.Examples)
	require.Contains(t, template.ContentTemplate, "recipient")
	require.Contains(t, template.ContentTemplate, "amount")
	require.Contains(t, template.ContentTemplate, "denom")
}

func TestGetProposalTemplate_InvalidCategory(t *testing.T) {
	keeper, _ := setupTemplateKeeper(t)

	// Use an invalid category value
	template := keeper.GetProposalTemplate(types.ProposalCategory(999))

	require.Nil(t, template)
}

func TestGetTextProposalTemplate(t *testing.T) {
	keeper, _ := setupTemplateKeeper(t)

	template := keeper.getTextProposalTemplate()

	require.NotNil(t, template)
	require.Equal(t, types.CategoryText, template.Category)
	require.Equal(t, "Text Proposal", template.Name)
	require.Contains(t, template.TitleTemplate, "Proposal:")
	require.Contains(t, template.DescriptionTemplate, "## Summary")
	require.Contains(t, template.DescriptionTemplate, "## Background")
	require.Contains(t, template.DescriptionTemplate, "## Proposal")
	require.Contains(t, template.DescriptionTemplate, "## Impact")
	require.Contains(t, template.DescriptionTemplate, "## Risks")
	require.Empty(t, template.ContentTemplate)
	require.Len(t, template.RequiredFields, 2)
	require.Contains(t, template.RequiredFields, "title")
	require.Contains(t, template.RequiredFields, "description")
	require.Contains(t, template.OptionalFields, "metadata")
	require.Len(t, template.Examples, 3)
}

func TestGetParameterChangeTemplate(t *testing.T) {
	keeper, _ := setupTemplateKeeper(t)

	template := keeper.getParameterChangeTemplate()

	require.NotNil(t, template)
	require.Equal(t, types.CategoryParameterChange, template.Category)
	require.Equal(t, "Parameter Change Proposal", template.Name)
	require.Contains(t, template.TitleTemplate, "Parameter Change:")
	require.Contains(t, template.DescriptionTemplate, "## Summary")
	require.Contains(t, template.DescriptionTemplate, "## Rationale")
	require.Contains(t, template.DescriptionTemplate, "## Impact Analysis")
	require.Contains(t, template.DescriptionTemplate, "## Implementation")
	require.NotEmpty(t, template.ContentTemplate)
	require.Contains(t, template.ContentTemplate, "changes")
	require.Len(t, template.RequiredFields, 3)
	require.Contains(t, template.RequiredFields, "title")
	require.Contains(t, template.RequiredFields, "description")
	require.Contains(t, template.RequiredFields, "content")
	require.Len(t, template.Examples, 3)
}

func TestGetSoftwareUpgradeTemplate(t *testing.T) {
	keeper, _ := setupTemplateKeeper(t)

	template := keeper.getSoftwareUpgradeTemplate()

	require.NotNil(t, template)
	require.Equal(t, types.CategorySoftwareUpgrade, template.Category)
	require.Equal(t, "Software Upgrade Proposal", template.Name)
	require.Contains(t, template.TitleTemplate, "Software Upgrade:")
	require.Contains(t, template.DescriptionTemplate, "## Upgrade Information")
	require.Contains(t, template.DescriptionTemplate, "## Changes")
	require.Contains(t, template.DescriptionTemplate, "## Breaking Changes")
	require.Contains(t, template.DescriptionTemplate, "## Migration Guide")
	require.Contains(t, template.DescriptionTemplate, "## Timeline")
	require.NotEmpty(t, template.ContentTemplate)
	require.Contains(t, template.ContentTemplate, "name")
	require.Contains(t, template.ContentTemplate, "height")
	require.Len(t, template.RequiredFields, 3)
	require.Contains(t, template.RequiredFields, "title")
	require.Contains(t, template.RequiredFields, "description")
	require.Contains(t, template.RequiredFields, "content")
	require.Len(t, template.Examples, 3)
}

func TestGetCommunitySpendTemplate(t *testing.T) {
	keeper, _ := setupTemplateKeeper(t)

	template := keeper.getCommunitySpendTemplate()

	require.NotNil(t, template)
	require.Equal(t, types.CategorySpending, template.Category)
	require.Equal(t, "Community Spend Proposal", template.Name)
	require.Contains(t, template.TitleTemplate, "Community Spend:")
	require.Contains(t, template.DescriptionTemplate, "## Project Information")
	require.Contains(t, template.DescriptionTemplate, "## Purpose")
	require.Contains(t, template.DescriptionTemplate, "## Deliverables")
	require.Contains(t, template.DescriptionTemplate, "## Budget Breakdown")
	require.Contains(t, template.DescriptionTemplate, "## Team")
	require.Contains(t, template.DescriptionTemplate, "## Success Metrics")
	require.NotEmpty(t, template.ContentTemplate)
	require.Contains(t, template.ContentTemplate, "recipient")
	require.Contains(t, template.ContentTemplate, "amount")
	require.Len(t, template.RequiredFields, 3)
	require.Contains(t, template.RequiredFields, "title")
	require.Contains(t, template.RequiredFields, "description")
	require.Contains(t, template.RequiredFields, "content")
	require.Len(t, template.Examples, 3)
}

func TestValidateProposalAgainstTemplate_TextProposal(t *testing.T) {
	keeper, _ := setupTemplateKeeper(t)

	tests := []struct {
		name        string
		category    types.ProposalCategory
		title       string
		description string
		content     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid text proposal",
			category:    types.CategoryText,
			title:       "Test Proposal",
			description: "This is a description",
			content:     "",
			expectError: false,
		},
		{
			name:        "missing title",
			category:    types.CategoryText,
			title:       "",
			description: "This is a description",
			content:     "",
			expectError: true,
			errorMsg:    "title is required",
		},
		{
			name:        "missing description",
			category:    types.CategoryText,
			title:       "Test Proposal",
			description: "",
			content:     "",
			expectError: true,
			errorMsg:    "description is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := keeper.ValidateProposalAgainstTemplate(tt.category, tt.title, tt.description, tt.content)

			if tt.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateProposalAgainstTemplate_ParameterChange(t *testing.T) {
	keeper, _ := setupTemplateKeeper(t)

	tests := []struct {
		name        string
		title       string
		description string
		content     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid parameter change",
			title:       "Parameter Change",
			description: "Change voting period",
			content:     `{"changes": [{"module": "governance", "parameter": "voting_period"}]}`,
			expectError: false,
		},
		{
			name:        "missing content",
			title:       "Parameter Change",
			description: "Change voting period",
			content:     "",
			expectError: true,
			errorMsg:    "content is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := keeper.ValidateProposalAgainstTemplate(types.CategoryParameterChange, tt.title, tt.description, tt.content)

			if tt.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateProposalAgainstTemplate_SoftwareUpgrade(t *testing.T) {
	keeper, _ := setupTemplateKeeper(t)

	tests := []struct {
		name        string
		title       string
		description string
		content     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid software upgrade",
			title:       "Upgrade v2.0",
			description: "Upgrade to version 2.0",
			content:     `{"name": "v2.0", "height": 1000000}`,
			expectError: false,
		},
		{
			name:        "missing content",
			title:       "Upgrade v2.0",
			description: "Upgrade to version 2.0",
			content:     "",
			expectError: true,
			errorMsg:    "content is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := keeper.ValidateProposalAgainstTemplate(types.CategorySoftwareUpgrade, tt.title, tt.description, tt.content)

			if tt.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateProposalAgainstTemplate_CommunitySpend(t *testing.T) {
	keeper, _ := setupTemplateKeeper(t)

	tests := []struct {
		name        string
		title       string
		description string
		content     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid community spend",
			title:       "Fund Development",
			description: "Fund feature development",
			content:     `{"recipient": "aura1...", "amount": "10000", "denom": "uaura"}`,
			expectError: false,
		},
		{
			name:        "missing content",
			title:       "Fund Development",
			description: "Fund feature development",
			content:     "",
			expectError: true,
			errorMsg:    "content is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := keeper.ValidateProposalAgainstTemplate(types.CategorySpending, tt.title, tt.description, tt.content)

			if tt.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateProposalAgainstTemplate_InvalidCategory(t *testing.T) {
	keeper, _ := setupTemplateKeeper(t)

	err := keeper.ValidateProposalAgainstTemplate(types.ProposalCategory(999), "Title", "Description", "Content")

	require.Error(t, err)
	require.Contains(t, err.Error(), "no template found")
}

func TestGetAllTemplates(t *testing.T) {
	keeper, _ := setupTemplateKeeper(t)

	templates := keeper.GetAllTemplates()

	require.NotNil(t, templates)
	require.Len(t, templates, 4)

	// Verify all categories are present
	categories := make(map[types.ProposalCategory]bool)
	for _, template := range templates {
		categories[template.Category] = true
	}

	require.True(t, categories[types.CategoryText])
	require.True(t, categories[types.CategoryParameterChange])
	require.True(t, categories[types.CategorySoftwareUpgrade])
	require.True(t, categories[types.CategorySpending])
}

func TestGetAllTemplates_AllValid(t *testing.T) {
	keeper, _ := setupTemplateKeeper(t)

	templates := keeper.GetAllTemplates()

	for _, template := range templates {
		require.NotNil(t, template)
		require.NotEmpty(t, template.Name)
		require.NotEmpty(t, template.Description)
		require.NotEmpty(t, template.TitleTemplate)
		require.NotEmpty(t, template.DescriptionTemplate)
		require.NotEmpty(t, template.RequiredFields)
		require.NotEmpty(t, template.Examples)
	}
}

func TestCreateProposalFromTemplate_TextProposal(t *testing.T) {
	keeper, ctx := setupTemplateKeeper(t)

	templateData := map[string]string{
		"title":       "Test Text Proposal",
		"description": "This is a test text proposal",
		"content":     "",
	}

	proposalID, err := keeper.CreateProposalFromTemplate(
		ctx,
		types.CategoryText,
		testAddr("proposer1"),
		templateData,
	)

	require.NoError(t, err)
	require.Equal(t, uint64(1), proposalID)

	// Verify proposal was created
	proposal, err := keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	require.Equal(t, "Test Text Proposal", proposal.Title)
	require.Equal(t, "This is a test text proposal", proposal.Description)
	require.Equal(t, types.CategoryText, proposal.Category)
}

func TestCreateProposalFromTemplate_ParameterChange(t *testing.T) {
	keeper, ctx := setupTemplateKeeper(t)

	templateData := map[string]string{
		"title":       "Change Voting Period",
		"description": "Increase voting period to 7 days",
		"content":     `{"changes": [{"module": "governance", "parameter": "voting_period"}]}`,
	}

	proposalID, err := keeper.CreateProposalFromTemplate(
		ctx,
		types.CategoryParameterChange,
		testAddr("proposer1"),
		templateData,
	)

	require.NoError(t, err)
	require.Equal(t, uint64(1), proposalID)

	// Verify proposal was created
	proposal, err := keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	require.Equal(t, "Change Voting Period", proposal.Title)
	require.Equal(t, types.CategoryParameterChange, proposal.Category)
}

func TestCreateProposalFromTemplate_SoftwareUpgrade(t *testing.T) {
	keeper, ctx := setupTemplateKeeper(t)

	templateData := map[string]string{
		"title":       "Upgrade to v2.0",
		"description": "Upgrade chain to version 2.0",
		"content":     `{"name": "v2.0", "height": 1000000}`,
	}

	proposalID, err := keeper.CreateProposalFromTemplate(
		ctx,
		types.CategorySoftwareUpgrade,
		testAddr("proposer1"),
		templateData,
	)

	require.NoError(t, err)
	require.Equal(t, uint64(1), proposalID)

	// Verify proposal was created
	proposal, err := keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	require.Equal(t, "Upgrade to v2.0", proposal.Title)
	require.Equal(t, types.CategorySoftwareUpgrade, proposal.Category)
}

func TestCreateProposalFromTemplate_CommunitySpend(t *testing.T) {
	keeper, ctx := setupTemplateKeeper(t)

	templateData := map[string]string{
		"title":       "Fund Marketing Campaign",
		"description": "Fund Q1 marketing campaign",
		"content":     `{"recipient": "aura1...", "amount": "50000", "denom": "uaura"}`,
	}

	proposalID, err := keeper.CreateProposalFromTemplate(
		ctx,
		types.CategorySpending,
		testAddr("proposer1"),
		templateData,
	)

	require.NoError(t, err)
	require.Equal(t, uint64(1), proposalID)

	// Verify proposal was created
	proposal, err := keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	require.Equal(t, "Fund Marketing Campaign", proposal.Title)
	require.Equal(t, types.CategorySpending, proposal.Category)
}

func TestCreateProposalFromTemplate_MissingTitle(t *testing.T) {
	keeper, ctx := setupTemplateKeeper(t)

	templateData := map[string]string{
		"title":       "",
		"description": "This is a description",
		"content":     "",
	}

	proposalID, err := keeper.CreateProposalFromTemplate(
		ctx,
		types.CategoryText,
		testAddr("proposer1"),
		templateData,
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "title is required")
	require.Equal(t, uint64(0), proposalID)
}

func TestCreateProposalFromTemplate_MissingDescription(t *testing.T) {
	keeper, ctx := setupTemplateKeeper(t)

	templateData := map[string]string{
		"title":       "Test Proposal",
		"description": "",
		"content":     "",
	}

	proposalID, err := keeper.CreateProposalFromTemplate(
		ctx,
		types.CategoryText,
		testAddr("proposer1"),
		templateData,
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "description is required")
	require.Equal(t, uint64(0), proposalID)
}

func TestCreateProposalFromTemplate_MissingContent(t *testing.T) {
	keeper, ctx := setupTemplateKeeper(t)

	templateData := map[string]string{
		"title":       "Parameter Change",
		"description": "Change voting period",
		"content":     "",
	}

	proposalID, err := keeper.CreateProposalFromTemplate(
		ctx,
		types.CategoryParameterChange,
		testAddr("proposer1"),
		templateData,
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "content is required")
	require.Equal(t, uint64(0), proposalID)
}

func TestCreateProposalFromTemplate_InvalidCategory(t *testing.T) {
	keeper, ctx := setupTemplateKeeper(t)

	templateData := map[string]string{
		"title":       "Test Proposal",
		"description": "This is a description",
		"content":     "",
	}

	proposalID, err := keeper.CreateProposalFromTemplate(
		ctx,
		types.ProposalCategory(999),
		testAddr("proposer1"),
		templateData,
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "template not found")
	require.Equal(t, uint64(0), proposalID)
}

func TestCreateProposalFromTemplate_MultipleProposals(t *testing.T) {
	keeper, ctx := setupTemplateKeeper(t)

	// Create multiple proposals from templates
	for i := 0; i < 5; i++ {
		templateData := map[string]string{
			"title":       "Test Proposal",
			"description": "Test Description",
			"content":     "",
		}

		proposalID, err := keeper.CreateProposalFromTemplate(
			ctx,
			types.CategoryText,
			testAddr("proposer1"),
			templateData,
		)

		require.NoError(t, err)
		require.Equal(t, uint64(i+1), proposalID)
	}

	// Verify all proposals exist
	for i := uint64(1); i <= 5; i++ {
		proposal, err := keeper.GetProposal(ctx, i)
		require.NoError(t, err)
		require.NotNil(t, proposal)
	}
}

func TestTemplateContentFormat(t *testing.T) {
	keeper, _ := setupTemplateKeeper(t)

	tests := []struct {
		name     string
		category types.ProposalCategory
		jsonKeys []string
	}{
		{
			name:     "parameter change has correct JSON keys",
			category: types.CategoryParameterChange,
			jsonKeys: []string{"changes", "module", "parameter", "current_value", "new_value"},
		},
		{
			name:     "software upgrade has correct JSON keys",
			category: types.CategorySoftwareUpgrade,
			jsonKeys: []string{"name", "height", "info", "version"},
		},
		{
			name:     "community spend has correct JSON keys",
			category: types.CategorySpending,
			jsonKeys: []string{"recipient", "amount", "denom", "purpose"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := keeper.GetProposalTemplate(tt.category)
			require.NotNil(t, template)

			for _, key := range tt.jsonKeys {
				require.Contains(t, template.ContentTemplate, key)
			}
		})
	}
}

func TestTemplateDescriptionSections(t *testing.T) {
	keeper, _ := setupTemplateKeeper(t)

	tests := []struct {
		name     string
		category types.ProposalCategory
		sections []string
	}{
		{
			name:     "text proposal sections",
			category: types.CategoryText,
			sections: []string{"## Summary", "## Background", "## Proposal", "## Impact", "## Risks"},
		},
		{
			name:     "parameter change sections",
			category: types.CategoryParameterChange,
			sections: []string{"## Summary", "## Rationale", "## Impact Analysis", "## Implementation"},
		},
		{
			name:     "software upgrade sections",
			category: types.CategorySoftwareUpgrade,
			sections: []string{"## Upgrade Information", "## Changes", "## Breaking Changes", "## Migration Guide", "## Timeline"},
		},
		{
			name:     "community spend sections",
			category: types.CategorySpending,
			sections: []string{"## Project Information", "## Purpose", "## Deliverables", "## Budget Breakdown", "## Team", "## Success Metrics"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := keeper.GetProposalTemplate(tt.category)
			require.NotNil(t, template)

			for _, section := range tt.sections {
				require.Contains(t, template.DescriptionTemplate, section)
			}
		})
	}
}

func TestTemplateExamplesProvided(t *testing.T) {
	keeper, _ := setupTemplateKeeper(t)

	categories := []types.ProposalCategory{
		types.CategoryText,
		types.CategoryParameterChange,
		types.CategorySoftwareUpgrade,
		types.CategorySpending,
	}

	for _, category := range categories {
		template := keeper.GetProposalTemplate(category)
		require.NotNil(t, template)
		require.NotEmpty(t, template.Examples)
		require.GreaterOrEqual(t, len(template.Examples), 3)
	}
}
