// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/governance/types"
)

// ============================
// PROPOSAL TEMPLATES (Feature 10)
// ============================

// GetProposalTemplate returns a proposal template by category
func (k *Keeper) GetProposalTemplate(category types.ProposalCategory) *types.ProposalTemplate {
	switch category {
	case types.CategoryText:
		return k.getTextProposalTemplate()
	case types.CategoryParameterChange:
		return k.getParameterChangeTemplate()
	case types.CategorySoftwareUpgrade:
		return k.getSoftwareUpgradeTemplate()
	case types.CategorySpending:
		return k.getCommunitySpendTemplate()
	default:
		return nil
	}
}

// getTextProposalTemplate returns template for text proposals
func (k *Keeper) getTextProposalTemplate() *types.ProposalTemplate {
	return &types.ProposalTemplate{
		Category:      types.CategoryText,
		Name:          "Text Proposal",
		Description:   "A text-based proposal for community discussion and signaling",
		TitleTemplate: "Proposal: [Your Title Here]",
		DescriptionTemplate: `## Summary
[Brief summary of the proposal]

## Background
[Context and motivation for the proposal]

## Proposal
[Detailed description of what is being proposed]

## Impact
[Expected impact and outcomes]

## Risks
[Potential risks and mitigation strategies]`,
		ContentTemplate: "",
		RequiredFields:  []string{"title", "description"},
		OptionalFields:  []string{"metadata"},
		Examples: []string{
			"Community direction proposal",
			"Feature request discussion",
			"Protocol improvement suggestion",
		},
	}
}

// getParameterChangeTemplate returns template for parameter change proposals
func (k *Keeper) getParameterChangeTemplate() *types.ProposalTemplate {
	return &types.ProposalTemplate{
		Category:      types.CategoryParameterChange,
		Name:          "Parameter Change Proposal",
		Description:   "Proposal to change governance or module parameters",
		TitleTemplate: "Parameter Change: [Module Name] - [Parameter Name]",
		DescriptionTemplate: `## Summary
Change parameter [parameter_name] from [old_value] to [new_value]

## Rationale
[Explanation of why this change is needed]

## Impact Analysis
- Economic impact: [analysis]
- Security impact: [analysis]
- User impact: [analysis]

## Implementation
Parameter changes will take effect immediately upon proposal execution.`,
		ContentTemplate: `{
  "changes": [
    {
      "module": "[module_name]",
      "parameter": "[parameter_name]",
      "current_value": "[current_value]",
      "new_value": "[new_value]"
    }
  ]
}`,
		RequiredFields: []string{"title", "description", "content"},
		OptionalFields: []string{"metadata"},
		Examples: []string{
			"Increase voting period from 3 days to 7 days",
			"Adjust minimum deposit requirement",
			"Update quorum threshold",
		},
	}
}

// getSoftwareUpgradeTemplate returns template for software upgrade proposals
func (k *Keeper) getSoftwareUpgradeTemplate() *types.ProposalTemplate {
	return &types.ProposalTemplate{
		Category:      types.CategorySoftwareUpgrade,
		Name:          "Software Upgrade Proposal",
		Description:   "Proposal to upgrade chain software to a new version",
		TitleTemplate: "Software Upgrade: [Version Number] - [Upgrade Name]",
		DescriptionTemplate: `## Upgrade Information
- Version: [version_number]
- Upgrade name: [upgrade_name]
- Upgrade height: [block_height]

## Changes
[List of major changes and improvements]

## Breaking Changes
[List any breaking changes]

## Migration Guide
[Instructions for validators and users]

## Timeline
- Proposal submission: [date]
- Voting period: [dates]
- Planned upgrade: [date/block]`,
		ContentTemplate: `{
  "name": "[upgrade_name]",
  "height": [block_height],
  "info": "[upgrade_info_url]",
  "version": "[version_number]"
}`,
		RequiredFields: []string{"title", "description", "content"},
		OptionalFields: []string{"metadata"},
		Examples: []string{
			"Upgrade to v2.0.0 with new features",
			"Security patch upgrade",
			"Performance optimization upgrade",
		},
	}
}

// getCommunitySpendTemplate returns template for community spend proposals
func (k *Keeper) getCommunitySpendTemplate() *types.ProposalTemplate {
	return &types.ProposalTemplate{
		Category:      types.CategorySpending,
		Name:          "Community Spend Proposal",
		Description:   "Proposal to spend from community pool",
		TitleTemplate: "Community Spend: [Project/Purpose Name]",
		DescriptionTemplate: `## Project Information
- Project name: [name]
- Recipient: [address]
- Amount requested: [amount]

## Purpose
[Detailed description of how funds will be used]

## Deliverables
[List of expected deliverables and timeline]

## Budget Breakdown
[Detailed budget allocation]

## Team
[Information about the team executing the project]

## Success Metrics
[How success will be measured]`,
		ContentTemplate: `{
  "recipient": "[recipient_address]",
  "amount": "[amount]",
  "denom": "[token_denom]",
  "purpose": "[purpose]"
}`,
		RequiredFields: []string{"title", "description", "content"},
		OptionalFields: []string{"metadata"},
		Examples: []string{
			"Fund development of new feature",
			"Marketing campaign funding",
			"Security audit funding",
		},
	}
}

// ValidateProposalAgainstTemplate validates a proposal against its template
func (k *Keeper) ValidateProposalAgainstTemplate(
	category types.ProposalCategory,
	title, description, content string,
) error {
	template := k.GetProposalTemplate(category)
	if template == nil {
		return fmt.Errorf("no template found for proposal category: %s", category)
	}

	// Validate required fields
	if title == "" {
		return fmt.Errorf("title is required")
	}

	if description == "" {
		return fmt.Errorf("description is required")
	}

	// Check if content is required
	for _, field := range template.RequiredFields {
		if field == "content" && content == "" {
			return fmt.Errorf("content is required for %s proposals", category)
		}
	}

	return nil
}

// GetAllTemplates returns all available proposal templates
func (k *Keeper) GetAllTemplates() []*types.ProposalTemplate {
	return []*types.ProposalTemplate{
		k.getTextProposalTemplate(),
		k.getParameterChangeTemplate(),
		k.getSoftwareUpgradeTemplate(),
		k.getCommunitySpendTemplate(),
	}
}

// CreateProposalFromTemplate creates a proposal using a template
func (k *Keeper) CreateProposalFromTemplate(
	ctx sdk.Context,
	category types.ProposalCategory,
	proposer string,
	templateData map[string]string,
) (uint64, error) {
	template := k.GetProposalTemplate(category)
	if template == nil {
		return 0, fmt.Errorf("template not found")
	}

	// Extract data from template
	title := templateData["title"]
	description := templateData["description"]
	content := templateData["content"]

	// Validate against template
	if err := k.ValidateProposalAgainstTemplate(category, title, description, content); err != nil {
		return 0, err
	}

	// Create proposal
	return k.CreateProposal(ctx, title, description, proposer, category, content)
}
