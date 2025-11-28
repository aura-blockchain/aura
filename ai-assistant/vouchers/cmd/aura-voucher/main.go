package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/spf13/cobra"

	"github.com/aequitas/aura/aiassistant/vouchers/pkg/voucher"
)

type globalFlags struct {
	keyPath     string
	dataDir     string
	pushGateway string
	profile     string
}

const envPassphraseVar = "AURA_VOUCHER_PASSPHRASE"

func main() {
	root := newRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	gf := &globalFlags{}
	cmd := &cobra.Command{
		Use:   "aura-voucher",
		Short: "Sponsorship voucher lifecycle tooling for Aura AI assistants",
	}
	cmd.PersistentFlags().StringVar(&gf.keyPath, "key", "", "Path to sponsor key file (default ~/.aura/assistant-vouchers/sponsor_key.json)")
	cmd.PersistentFlags().StringVar(&gf.dataDir, "data-dir", "", "Directory for issued/redeemed ledgers (default ~/.aura/assistant-vouchers)")
	cmd.PersistentFlags().StringVar(&gf.pushGateway, "pushgateway", "", "Optional Prometheus Pushgateway URL for voucher metrics")
	cmd.PersistentFlags().StringVar(&gf.profile, "profile", "default", "Named profile for multi-sponsor deployments (separate keys and ledgers)")

	cmd.AddCommand(
		newKeygenCmd(gf),
		newIssueCmd(gf),
		newRedeemCmd(gf),
		newInspectCmd(),
		newServeCmd(gf),
	)
	return cmd
}

func newKeygenCmd(gf *globalFlags) *cobra.Command {
	var passphraseFile string
	var overwrite bool
	cmd := &cobra.Command{
		Use:   "keygen",
		Short: "Generate an ed25519 sponsor key",
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := voucher.NewManager(gf.keyPath, gf.dataDir, gf.profile)
			if err != nil {
				return err
			}
			passphrase, err := readOptionalFile(passphraseFile)
			if err != nil {
				return err
			}
			pub, err := mgr.GenerateKey(strings.TrimSpace(passphrase), overwrite)
			if err != nil {
				return err
			}
			fmt.Printf("Public key: %s\n", base64.StdEncoding.EncodeToString(pub))
			fmt.Printf("Key saved to %s\n", mgr.KeyPath)
			return nil
		},
	}
	cmd.Flags().StringVar(&passphraseFile, "passphrase-file", "", "Optional file containing passphrase to encrypt the private key")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing key if present")
	return cmd
}

func newIssueCmd(gf *globalFlags) *cobra.Command {
	var amount, denom, sponsor, notes, expires, passphraseFile, nonce string
	cmd := &cobra.Command{
		Use:   "issue",
		Short: "Issue a voucher for assistant sponsorship",
		RunE: func(cmd *cobra.Command, args []string) error {
			if amount == "" {
				return fmt.Errorf("--amount is required (micro-denom, e.g. 5000000)")
			}
			if sponsor == "" {
				return fmt.Errorf("--sponsor is required (Aura address)")
			}
			expiry := time.Now().Add(30 * 24 * time.Hour)
			if expires != "" {
				parsed, err := time.Parse(time.RFC3339, expires)
				if err != nil {
					return fmt.Errorf("parse --expires: %w", err)
				}
				expiry = parsed
			}
			mgr, err := voucher.NewManager(gf.keyPath, gf.dataDir, gf.profile)
			if err != nil {
				return err
			}
			passphrase, err := resolvePassphrase(passphraseFile)
			if err != nil {
				return err
			}
			opts := voucher.IssueOptions{
				Amount:     amount,
				Denom:      denom,
				Sponsor:    sponsor,
				ExpiresAt:  expiry,
				Notes:      notes,
				Passphrase: strings.TrimSpace(passphrase),
				KeyPath:    mgr.KeyPath,
				DataDir:    mgr.DataDir,
				Nonce:      nonce,
			}
			signed, err := mgr.Issue(opts)
			if err != nil {
				return err
			}
			encoded, err := voucher.EncodeVoucher(signed)
			if err != nil {
				return err
			}
			fmt.Println(encoded)
			fmt.Fprintf(os.Stderr, "Voucher %s issued (%s %s)\n", signed.Voucher.ID, signed.Voucher.Amount, signed.Voucher.Denom)
			if gf.pushGateway != "" {
				pushMetric(gf.pushGateway, "assistant_voucher_issue_total", map[string]string{
					"sponsor": signed.Voucher.Sponsor,
					"denom":   signed.Voucher.Denom,
				})
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&amount, "amount", "", "Voucher credit amount in micro-denom (required)")
	cmd.Flags().StringVar(&denom, "denom", "uaura", "Denomination (default uaura)")
	cmd.Flags().StringVar(&sponsor, "sponsor", "", "Sponsor bech32 address (required)")
	cmd.Flags().StringVar(&notes, "notes", "", "Optional memo for auditors")
	cmd.Flags().StringVar(&expires, "expires", "", "RFC3339 expiry (default 30d from now)")
	cmd.Flags().StringVar(&passphraseFile, "passphrase-file", "", "File containing key passphrase (if encrypted)")
	cmd.Flags().StringVar(&nonce, "nonce", "", "Optional external reference/nonce")
	return cmd
}

func newRedeemCmd(gf *globalFlags) *cobra.Command {
	var voucherFile, voucherString, assistant, expectPubKey, txHash, notes string
	cmd := &cobra.Command{
		Use:   "redeem",
		Short: "Redeem a voucher locally and log the audit trail",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload := voucherString
			if payload == "" && voucherFile != "" {
				data, err := os.ReadFile(voucherFile)
				if err != nil {
					return err
				}
				payload = strings.TrimSpace(string(data))
			}
			if payload == "" {
				return fmt.Errorf("--voucher or --voucher-file is required")
			}
			mgr, err := voucher.NewManager(gf.keyPath, gf.dataDir, gf.profile)
			if err != nil {
				return err
			}
			opts := voucher.RedeemOptions{
				EncodedVoucher: payload,
				Assistant:      assistant,
				ExpectPubKey:   strings.TrimSpace(expectPubKey),
				KeyPath:        mgr.KeyPath,
				DataDir:        mgr.DataDir,
				TxHash:         txHash,
				Notes:          notes,
			}
			signed, err := mgr.Redeem(opts)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Voucher %s verified (amount %s %s)\n", signed.Voucher.ID, signed.Voucher.Amount, signed.Voucher.Denom)
			if gf.pushGateway != "" {
				pushMetric(gf.pushGateway, "assistant_voucher_redeem_total", map[string]string{
					"sponsor": signed.Voucher.Sponsor,
					"assistant": func() string {
						if assistant == "" {
							return "unknown"
						}
						return assistant
					}(),
				})
			}
			buf, _ := json.MarshalIndent(signed, "", "  ")
			fmt.Println(string(buf))
			return nil
		},
	}
	cmd.Flags().StringVar(&voucherFile, "voucher-file", "", "Path to file containing the voucher payload")
	cmd.Flags().StringVar(&voucherString, "voucher", "", "Voucher payload (base64 or JSON)")
	cmd.Flags().StringVar(&assistant, "assistant", "", "Assistant address redeeming the voucher (optional audit field)")
	cmd.Flags().StringVar(&expectPubKey, "expect-pubkey", "", "Optional base64 public key to enforce sponsor identity")
	cmd.Flags().StringVar(&txHash, "tx-hash", "", "Optional transaction hash after on-chain redemption")
	cmd.Flags().StringVar(&notes, "notes", "", "Audit notes")
	return cmd
}

func newInspectCmd() *cobra.Command {
	var voucherString string
	cmd := &cobra.Command{
		Use:   "inspect",
		Short: "Decode a voucher payload without redeeming it",
		RunE: func(cmd *cobra.Command, args []string) error {
			if voucherString == "" && len(args) > 0 {
				voucherString = args[0]
			}
			if voucherString == "" {
				return fmt.Errorf("provide a voucher payload")
			}
			signed, err := voucher.DecodeVoucher(voucherString)
			if err != nil {
				return err
			}
			buf, _ := json.MarshalIndent(signed, "", "  ")
			fmt.Println(string(buf))
			return nil
		},
	}
	cmd.Flags().StringVar(&voucherString, "voucher", "", "Voucher payload (base64 or JSON)")
	return cmd
}

func newServeCmd(gf *globalFlags) *cobra.Command {
	var listen, token, passphraseFile string
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Expose voucher issue/redeem over a REST API (for automation)",
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := voucher.NewManager(gf.keyPath, gf.dataDir, gf.profile)
			if err != nil {
				return err
			}
			passphrase, err := resolvePassphrase(passphraseFile)
			if err != nil {
				return err
			}
			mux := http.NewServeMux()
			mux.HandleFunc("/api/healthz", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, "ok")
			})
			mux.HandleFunc("/api/issue", func(w http.ResponseWriter, r *http.Request) {
				if !authorizeRequest(w, r, token) {
					return
				}
				if r.Method != http.MethodPost {
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
					return
				}
				var req issueRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					http.Error(w, "invalid payload", http.StatusBadRequest)
					return
				}
				if req.Amount == "" || req.Sponsor == "" {
					http.Error(w, "amount and sponsor are required", http.StatusBadRequest)
					return
				}
				expiry := time.Now().Add(30 * 24 * time.Hour)
				if req.Expires != "" {
					if parsed, err := time.Parse(time.RFC3339, req.Expires); err == nil {
						expiry = parsed
					}
				}
				opts := voucher.IssueOptions{
					Amount:     req.Amount,
					Denom:      pick(req.Denom, "uaura"),
					Sponsor:    req.Sponsor,
					ExpiresAt:  expiry,
					Notes:      req.Notes,
					Passphrase: passphrase,
					KeyPath:    mgr.KeyPath,
					DataDir:    mgr.DataDir,
					Nonce:      req.Nonce,
				}
				signed, err := mgr.Issue(opts)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				encoded, err := voucher.EncodeVoucher(signed)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, issueResponse{
					SignedVoucher: signed,
					Encoded:       encoded,
				})
			})

			mux.HandleFunc("/api/redeem", func(w http.ResponseWriter, r *http.Request) {
				if !authorizeRequest(w, r, token) {
					return
				}
				if r.Method != http.MethodPost {
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
					return
				}
				var req redeemRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					http.Error(w, "invalid payload", http.StatusBadRequest)
					return
				}
				if req.Voucher == "" {
					http.Error(w, "voucher required", http.StatusBadRequest)
					return
				}
				mgr, err := voucher.NewManager(gf.keyPath, gf.dataDir, gf.profile)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				opts := voucher.RedeemOptions{
					EncodedVoucher: req.Voucher,
					Assistant:      req.Assistant,
					ExpectPubKey:   req.ExpectPubKey,
					KeyPath:        mgr.KeyPath,
					DataDir:        mgr.DataDir,
					TxHash:         req.TxHash,
					Notes:          req.Notes,
				}
				signed, err := mgr.Redeem(opts)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, signed)
			})

			server := &http.Server{
				Addr:    listen,
				Handler: mux,
			}
			fmt.Fprintf(os.Stderr, "Voucher REST API listening on %s\n", listen)
			return server.ListenAndServe()
		},
	}
	cmd.Flags().StringVar(&listen, "listen", ":8787", "Address for the REST server")
	cmd.Flags().StringVar(&token, "token", "", "Optional bearer token required in Authorization header")
	cmd.Flags().StringVar(&passphraseFile, "passphrase-file", "", "File containing key passphrase (fall back to AURA_VOUCHER_PASSPHRASE env var)")
	return cmd
}

func readOptionalFile(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func resolvePassphrase(passphraseFile string) (string, error) {
	if passphraseFile != "" {
		return readOptionalFile(passphraseFile)
	}
	env := strings.TrimSpace(os.Getenv(envPassphraseVar))
	if env != "" {
		return env, nil
	}
	return "", nil
}

type issueRequest struct {
	Amount  string `json:"amount"`
	Denom   string `json:"denom"`
	Sponsor string `json:"sponsor"`
	Expires string `json:"expires"`
	Notes   string `json:"notes"`
	Nonce   string `json:"nonce"`
}

type issueResponse struct {
	SignedVoucher voucher.SignedVoucher `json:"signed_voucher"`
	Encoded       string                `json:"encoded"`
}

type redeemRequest struct {
	Voucher      string `json:"voucher"`
	Assistant    string `json:"assistant"`
	ExpectPubKey string `json:"expect_pubkey"`
	TxHash       string `json:"tx_hash"`
	Notes        string `json:"notes"`
}

func authorizeRequest(w http.ResponseWriter, r *http.Request, token string) bool {
	if token == "" {
		return true
	}
	expected := "Bearer " + token
	if strings.TrimSpace(r.Header.Get("Authorization")) != expected {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func pick(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

// pushMetric sends a single counter increment to the given Pushgateway.
func pushMetric(gateway, metricName string, labels map[string]string) {
	if gateway == "" {
		return
	}
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: metricName,
		Help: "Aura assistant voucher metric",
		ConstLabels: func() prometheus.Labels {
			out := prometheus.Labels{}
			for k, v := range labels {
				out[k] = v
			}
			return out
		}(),
	})
	counter.Add(1)
	job := "aiassistant_voucher"
	if err := push.New(gateway, job).Collector(counter).Push(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: pushgateway push failed: %v\n", err)
	}
}
