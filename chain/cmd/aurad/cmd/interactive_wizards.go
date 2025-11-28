package cmd

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

// runVCRegistryWizard runs an interactive wizard for VC Registry operations
func runVCRegistryWizard(reader *bufio.Reader) error {
	fmt.Println("\n=== VC Registry Wizard ===")
	fmt.Println("What would you like to do?")

	fmt.Println("  1. Mint a new Verifiable Credential")
	fmt.Println("  2. Revoke a Verifiable Credential")
	fmt.Println("  3. Register a new DID")
	fmt.Println("  4. Update DID document")
	fmt.Println("  5. Query VC by ID")
	fmt.Println("  6. Query DIDs by controller")
	fmt.Print("\nEnter number: ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		return runMintVCWizard(reader)
	case "2":
		return runRevokeVCWizard(reader)
	case "3":
		return runRegisterDIDWizard(reader)
	case "4":
		return runUpdateDIDWizard(reader)
	case "5":
		return runQueryVCWizard(reader)
	case "6":
		return runQueryDIDsWizard(reader)
	default:
		return fmt.Errorf("invalid choice: %s", choice)
	}
}

func runMintVCWizard(reader *bufio.Reader) error {
	fmt.Println("\n=== Mint Verifiable Credential ===")

	fmt.Print("Holder DID (e.g., did:aura:mainnet:user123): ")
	holderDID, _ := reader.ReadString('\n')
	holderDID = strings.TrimSpace(holderDID)

	fmt.Println("\nVC Type Options:")
	fmt.Println("  1. VC_TYPE_VERIFIED_HUMAN")
	fmt.Println("  2. VC_TYPE_AGE_OVER_18")
	fmt.Println("  3. VC_TYPE_AGE_OVER_21")
	fmt.Println("  4. VC_TYPE_KYC_VERIFICATION")
	fmt.Println("  5. VC_TYPE_BIOMETRIC_AUTH")
	fmt.Println("  6. VC_TYPE_CUSTOM")
	fmt.Print("Select VC type: ")

	vcTypeChoice, _ := reader.ReadString('\n')
	vcTypeChoice = strings.TrimSpace(vcTypeChoice)

	vcTypes := []string{"", "VC_TYPE_VERIFIED_HUMAN", "VC_TYPE_AGE_OVER_18",
		"VC_TYPE_AGE_OVER_21", "VC_TYPE_KYC_VERIFICATION", "VC_TYPE_BIOMETRIC_AUTH", "VC_TYPE_CUSTOM"}

	vcTypeIdx, _ := strconv.Atoi(vcTypeChoice)
	if vcTypeIdx < 1 || vcTypeIdx >= len(vcTypes) {
		return fmt.Errorf("invalid VC type selection")
	}
	vcType := vcTypes[vcTypeIdx]

	var customType string
	if vcType == "VC_TYPE_CUSTOM" {
		fmt.Print("Enter custom type name: ")
		customType, _ = reader.ReadString('\n')
		customType = strings.TrimSpace(customType)
	}

	fmt.Print("Metadata (optional, format: key1=val1,key2=val2): ")
	metadata, _ := reader.ReadString('\n')
	metadata = strings.TrimSpace(metadata)

	fmt.Print("From address/key name: ")
	from, _ := reader.ReadString('\n')
	from = strings.TrimSpace(from)

	fmt.Println("\n=== Transaction Summary ===")
	fmt.Printf("Holder DID:  %s\n", holderDID)
	fmt.Printf("VC Type:     %s\n", vcType)
	if customType != "" {
		fmt.Printf("Custom Type: %s\n", customType)
	}
	if metadata != "" {
		fmt.Printf("Metadata:    %s\n", metadata)
	}
	fmt.Printf("From:        %s\n", from)

	fmt.Print("\nExecute transaction? (y/n): ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm == "y" || confirm == "yes" {
		// Build command
		cmdStr := fmt.Sprintf("aurad tx vcregistry mint-vc %s %s", holderDID, vcType)
		if customType != "" {
			cmdStr += fmt.Sprintf(" --custom-type %s", customType)
		}
		if metadata != "" {
			cmdStr += fmt.Sprintf(" --metadata \"%s\"", metadata)
		}
		cmdStr += fmt.Sprintf(" --from %s", from)

		fmt.Println("\nGenerated command:")
		fmt.Println(cmdStr)
		fmt.Println("\nCopy and execute this command to proceed.")
	} else {
		fmt.Println("Transaction cancelled.")
	}

	return nil
}

func runRevokeVCWizard(reader *bufio.Reader) error {
	fmt.Println("\n=== Revoke Verifiable Credential ===")

	fmt.Print("VC ID to revoke: ")
	vcID, _ := reader.ReadString('\n')
	vcID = strings.TrimSpace(vcID)

	fmt.Print("Reason for revocation (optional): ")
	reason, _ := reader.ReadString('\n')
	reason = strings.TrimSpace(reason)

	fmt.Print("From address/key name: ")
	from, _ := reader.ReadString('\n')
	from = strings.TrimSpace(from)

	fmt.Println("\n=== Transaction Summary ===")
	fmt.Printf("VC ID:  %s\n", vcID)
	if reason != "" {
		fmt.Printf("Reason: %s\n", reason)
	}
	fmt.Printf("From:   %s\n", from)

	fmt.Print("\nExecute transaction? (y/n): ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm == "y" || confirm == "yes" {
		cmdStr := fmt.Sprintf("aurad tx vcregistry revoke-vc %s", vcID)
		if reason != "" {
			cmdStr += fmt.Sprintf(" --reason \"%s\"", reason)
		}
		cmdStr += fmt.Sprintf(" --from %s", from)

		fmt.Println("\nGenerated command:")
		fmt.Println(cmdStr)
		fmt.Println("\nCopy and execute this command to proceed.")
	} else {
		fmt.Println("Transaction cancelled.")
	}

	return nil
}

func runRegisterDIDWizard(reader *bufio.Reader) error {
	fmt.Println("\n=== Register DID ===")

	fmt.Print("DID to register (e.g., did:aura:mainnet:user123): ")
	did, _ := reader.ReadString('\n')
	did = strings.TrimSpace(did)

	fmt.Print("Metadata URI (optional, e.g., ipfs://Qm...): ")
	metadataURI, _ := reader.ReadString('\n')
	metadataURI = strings.TrimSpace(metadataURI)

	fmt.Print("Verification method (optional, format: id:type:pubkey_hex): ")
	verificationMethod, _ := reader.ReadString('\n')
	verificationMethod = strings.TrimSpace(verificationMethod)

	fmt.Print("From address/key name: ")
	from, _ := reader.ReadString('\n')
	from = strings.TrimSpace(from)

	fmt.Println("\n=== Transaction Summary ===")
	fmt.Printf("DID:      %s\n", did)
	if metadataURI != "" {
		fmt.Printf("Metadata: %s\n", metadataURI)
	}
	if verificationMethod != "" {
		fmt.Printf("VM:       %s\n", verificationMethod)
	}
	fmt.Printf("From:     %s\n", from)

	fmt.Print("\nExecute transaction? (y/n): ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm == "y" || confirm == "yes" {
		cmdStr := fmt.Sprintf("aurad tx vcregistry register-did %s", did)
		if metadataURI != "" {
			cmdStr += fmt.Sprintf(" --metadata-uri %s", metadataURI)
		}
		if verificationMethod != "" {
			cmdStr += fmt.Sprintf(" --verification-method %s", verificationMethod)
		}
		cmdStr += fmt.Sprintf(" --from %s", from)

		fmt.Println("\nGenerated command:")
		fmt.Println(cmdStr)
		fmt.Println("\nCopy and execute this command to proceed.")
	} else {
		fmt.Println("Transaction cancelled.")
	}

	return nil
}

func runUpdateDIDWizard(reader *bufio.Reader) error {
	fmt.Println("\n=== Update DID Document ===")

	fmt.Print("DID to update: ")
	did, _ := reader.ReadString('\n')
	did = strings.TrimSpace(did)

	fmt.Print("New metadata URI (optional): ")
	metadataURI, _ := reader.ReadString('\n')
	metadataURI = strings.TrimSpace(metadataURI)

	fmt.Print("New verification method (optional): ")
	verificationMethod, _ := reader.ReadString('\n')
	verificationMethod = strings.TrimSpace(verificationMethod)

	fmt.Print("From address/key name: ")
	from, _ := reader.ReadString('\n')
	from = strings.TrimSpace(from)

	cmdStr := fmt.Sprintf("aurad tx vcregistry update-did %s", did)
	if metadataURI != "" {
		cmdStr += fmt.Sprintf(" --metadata-uri %s", metadataURI)
	}
	if verificationMethod != "" {
		cmdStr += fmt.Sprintf(" --verification-method %s", verificationMethod)
	}
	cmdStr += fmt.Sprintf(" --from %s", from)

	fmt.Println("\nGenerated command:")
	fmt.Println(cmdStr)

	return nil
}

func runQueryVCWizard(reader *bufio.Reader) error {
	fmt.Println("\n=== Query Verifiable Credential ===")

	fmt.Print("VC ID: ")
	vcID, _ := reader.ReadString('\n')
	vcID = strings.TrimSpace(vcID)

	cmdStr := fmt.Sprintf("aurad query vcregistry vc %s", vcID)
	fmt.Println("\nGenerated command:")
	fmt.Println(cmdStr)

	return nil
}

func runQueryDIDsWizard(reader *bufio.Reader) error {
	fmt.Println("\n=== Query DIDs by Controller ===")

	fmt.Print("Controller address: ")
	controller, _ := reader.ReadString('\n')
	controller = strings.TrimSpace(controller)

	cmdStr := fmt.Sprintf("aurad query vcregistry dids-by-controller %s", controller)
	fmt.Println("\nGenerated command:")
	fmt.Println(cmdStr)

	return nil
}

// runBridgeWizard runs an interactive wizard for Bridge operations
func runBridgeWizard(reader *bufio.Reader) error {
	fmt.Println("\n=== Bridge Wizard ===")
	fmt.Println("What would you like to do?")

	fmt.Println("  1. Link addresses (AURA/PAW/XAI)")
	fmt.Println("  2. Lock tokens for cross-chain transfer")
	fmt.Println("  3. Unlock tokens")
	fmt.Println("  4. Query linked addresses")
	fmt.Print("\nEnter number: ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		return runLinkAddressWizard(reader)
	case "2":
		return runLockTokensWizard(reader)
	case "3":
		return runUnlockTokensWizard(reader)
	case "4":
		return runQueryLinkedAddressesWizard(reader)
	default:
		return fmt.Errorf("invalid choice: %s", choice)
	}
}

func runLinkAddressWizard(reader *bufio.Reader) error {
	fmt.Println("\n=== Link Addresses ===")

	fmt.Print("AURA address: ")
	auraAddr, _ := reader.ReadString('\n')
	auraAddr = strings.TrimSpace(auraAddr)

	fmt.Print("PAW address (optional, press Enter to skip): ")
	pawAddr, _ := reader.ReadString('\n')
	pawAddr = strings.TrimSpace(pawAddr)

	fmt.Print("XAI address (optional, press Enter to skip): ")
	xaiAddr, _ := reader.ReadString('\n')
	xaiAddr = strings.TrimSpace(xaiAddr)

	fmt.Print("From address/key name: ")
	from, _ := reader.ReadString('\n')
	from = strings.TrimSpace(from)

	cmdStr := fmt.Sprintf("aurad tx bridge link-address %s %s %s --from %s",
		auraAddr, pawAddr, xaiAddr, from)

	fmt.Println("\nGenerated command:")
	fmt.Println(cmdStr)

	return nil
}

func runLockTokensWizard(reader *bufio.Reader) error {
	fmt.Println("\n=== Lock Tokens ===")

	fmt.Print("Target chain (PAW/XAI): ")
	targetChain, _ := reader.ReadString('\n')
	targetChain = strings.TrimSpace(strings.ToUpper(targetChain))

	fmt.Print("Amount to lock (e.g., 1000uaura): ")
	amount, _ := reader.ReadString('\n')
	amount = strings.TrimSpace(amount)

	fmt.Print("Recipient address on target chain: ")
	recipient, _ := reader.ReadString('\n')
	recipient = strings.TrimSpace(recipient)

	fmt.Print("From address/key name: ")
	from, _ := reader.ReadString('\n')
	from = strings.TrimSpace(from)

	cmdStr := fmt.Sprintf("aurad tx bridge lock-tokens %s %s %s --from %s",
		targetChain, amount, recipient, from)

	fmt.Println("\nGenerated command:")
	fmt.Println(cmdStr)

	return nil
}

func runUnlockTokensWizard(reader *bufio.Reader) error {
	fmt.Println("\n=== Unlock Tokens ===")

	fmt.Print("Lock ID: ")
	lockID, _ := reader.ReadString('\n')
	lockID = strings.TrimSpace(lockID)

	fmt.Print("Proof (hex): ")
	proof, _ := reader.ReadString('\n')
	proof = strings.TrimSpace(proof)

	fmt.Print("From address/key name: ")
	from, _ := reader.ReadString('\n')
	from = strings.TrimSpace(from)

	cmdStr := fmt.Sprintf("aurad tx bridge unlock-tokens %s --proof %s --from %s",
		lockID, proof, from)

	fmt.Println("\nGenerated command:")
	fmt.Println(cmdStr)

	return nil
}

func runQueryLinkedAddressesWizard(reader *bufio.Reader) error {
	fmt.Println("\n=== Query Linked Addresses ===")

	fmt.Print("AURA address: ")
	auraAddr, _ := reader.ReadString('\n')
	auraAddr = strings.TrimSpace(auraAddr)

	cmdStr := fmt.Sprintf("aurad query bridge linked-addresses %s", auraAddr)

	fmt.Println("\nGenerated command:")
	fmt.Println(cmdStr)

	return nil
}

// runInclusionRoutinesWizard runs an interactive wizard for Inclusion Routines
func runInclusionRoutinesWizard(reader *bufio.Reader) error {
	fmt.Println("\n=== Inclusion Routines Wizard ===")
	fmt.Println("What would you like to do?")

	fmt.Println("  1. Complete an Inclusion Routine")
	fmt.Println("  2. Query IR by ID")
	fmt.Println("  3. Query user's completed IRs")
	fmt.Println("  4. List available IRs")
	fmt.Print("\nEnter number: ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		return runCompleteIRWizard(reader)
	case "2":
		return runQueryIRWizard(reader)
	case "3":
		return runQueryUserIRsWizard(reader)
	case "4":
		fmt.Println("\nGenerated command:")
		fmt.Println("aurad query inclusionroutines list")
	default:
		return fmt.Errorf("invalid choice: %s", choice)
	}

	return nil
}

func runCompleteIRWizard(reader *bufio.Reader) error {
	fmt.Println("\n=== Complete Inclusion Routine ===")

	fmt.Print("IR ID: ")
	irID, _ := reader.ReadString('\n')
	irID = strings.TrimSpace(irID)

	fmt.Print("Proof data (optional): ")
	proof, _ := reader.ReadString('\n')
	proof = strings.TrimSpace(proof)

	fmt.Print("From address/key name: ")
	from, _ := reader.ReadString('\n')
	from = strings.TrimSpace(from)

	cmdStr := fmt.Sprintf("aurad tx inclusionroutines complete %s", irID)
	if proof != "" {
		cmdStr += fmt.Sprintf(" --proof \"%s\"", proof)
	}
	cmdStr += fmt.Sprintf(" --from %s", from)

	fmt.Println("\nGenerated command:")
	fmt.Println(cmdStr)

	return nil
}

func runQueryIRWizard(reader *bufio.Reader) error {
	fmt.Println("\n=== Query Inclusion Routine ===")

	fmt.Print("IR ID: ")
	irID, _ := reader.ReadString('\n')
	irID = strings.TrimSpace(irID)

	cmdStr := fmt.Sprintf("aurad query inclusionroutines ir %s", irID)
	fmt.Println("\nGenerated command:")
	fmt.Println(cmdStr)

	return nil
}

func runQueryUserIRsWizard(reader *bufio.Reader) error {
	fmt.Println("\n=== Query User's Completed IRs ===")

	fmt.Print("User address: ")
	userAddr, _ := reader.ReadString('\n')
	userAddr = strings.TrimSpace(userAddr)

	cmdStr := fmt.Sprintf("aurad query inclusionroutines user-irs %s", userAddr)
	fmt.Println("\nGenerated command:")
	fmt.Println(cmdStr)

	return nil
}
