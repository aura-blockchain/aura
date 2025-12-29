// Rewards Panel Component
// Handles claiming staking rewards with Keplr wallet

import { formatAmount, showToast, showLoading, hideLoading } from '../utils/ui.js';
import { AuraWalletConnector, WalletError } from '../../wallet-connector.js';

export class RewardsPanel {
    constructor(api, wallet) {
        this.api = api;
        this.wallet = wallet;
        this.delegatorAddress = null;
        this.rewards = null;
    }

    async render(delegatorAddress) {
        this.delegatorAddress = delegatorAddress;

        const container = document.getElementById('rewards-panel');
        if (!container) return;

        container.innerHTML = '<div class="text-center">Loading rewards...</div>';

        try {
            this.rewards = await this.api.getDelegationRewards(delegatorAddress);
            const delegations = await this.api.getDelegations(delegatorAddress);

            const totalRewards = this.calculateTotalRewards(this.rewards);

            container.innerHTML = `
                <div class="rewards-summary" style="margin-bottom: 2rem;">
                    <div style="text-align: center; padding: 2rem; background: linear-gradient(135deg, var(--primary-color), var(--secondary-color)); color: white; border-radius: var(--radius);">
                        <div style="font-size: 0.875rem; opacity: 0.9; margin-bottom: 0.5rem;">Total Pending Rewards</div>
                        <div style="font-size: 2.5rem; font-weight: 700; margin-bottom: 1rem;">
                            ${formatAmount(totalRewards * 1e6)} AURA
                        </div>
                        <button id="claim-all-btn" class="btn btn-primary" style="background: white; color: var(--primary-color);" ${totalRewards === 0 ? 'disabled' : ''}>
                            <i class="fas fa-gift"></i> Claim All Rewards
                        </button>
                    </div>
                </div>

                <div id="fee-info" style="display: flex; justify-content: space-between; padding: 0.75rem; background: var(--bg-tertiary); border-radius: var(--radius); margin-bottom: 1rem; font-size: 0.875rem;">
                    <span>Estimated Fee (claim all):</span>
                    <span>~${(0.0075 * Math.max(1, this.rewards.rewards?.length || 1)).toFixed(4)} AURA</span>
                </div>

                <h4 style="margin-bottom: 1rem;">Rewards by Validator</h4>
                <div id="rewards-list">
                    ${this.renderRewardsList(delegations)}
                </div>

                ${totalRewards > 0 ? `
                    <div style="margin-top: 2rem; padding: 1rem; background: var(--bg-tertiary); border-radius: var(--radius);">
                        <label class="checkbox-label" style="display: flex; align-items: center; gap: 0.5rem;">
                            <input type="checkbox" id="auto-compound-checkbox">
                            <span>Auto-compound rewards (re-stake immediately after claiming)</span>
                        </label>
                        <small style="display: block; margin-top: 0.5rem; color: var(--text-secondary);">
                            This will send two transactions: claim rewards, then delegate to your largest validator.
                        </small>
                    </div>
                ` : ''}
            `;

            this.setupEventListeners();
        } catch (error) {
            console.error('Error loading rewards:', error);
            container.innerHTML = '<div class="text-center text-danger">Failed to load rewards</div>';
        }
    }

    calculateTotalRewards(rewards) {
        if (!rewards.total || rewards.total.length === 0) return 0;

        return rewards.total.reduce((sum, r) => {
            if (r.denom === 'uaura') {
                return sum + r.amount;
            }
            return sum;
        }, 0);
    }

    renderRewardsList(delegations) {
        if (!this.rewards.rewards || this.rewards.rewards.length === 0) {
            return '<div class="text-center" style="padding: 2rem; color: var(--text-secondary);">No pending rewards</div>';
        }

        return this.rewards.rewards.map(r => {
            const delegation = delegations.find(d => d.validatorAddress === r.validatorAddress);
            const auraReward = r.reward.find(rw => rw.denom === 'uaura');
            const rewardAmount = auraReward ? auraReward.amount : 0;

            return `
                <div class="delegation-item" style="display: flex; justify-content: space-between; align-items: center; padding: 1rem; background: var(--bg-secondary); border-radius: var(--radius); margin-bottom: 0.5rem;">
                    <div>
                        <div style="font-weight: 600; margin-bottom: 0.25rem;">
                            ${r.validatorAddress.slice(0, 20)}...
                        </div>
                        <div style="font-size: 0.875rem; color: var(--text-secondary);">
                            Delegated: ${delegation ? formatAmount(delegation.balance * 1e6) : '0'} AURA
                        </div>
                    </div>
                    <div style="text-align: right;">
                        <div style="font-size: 1.125rem; font-weight: 600; color: var(--success-color);">
                            ${formatAmount(rewardAmount * 1e6)} AURA
                        </div>
                        <button
                            class="btn btn-sm btn-primary claim-single-btn"
                            data-validator="${r.validatorAddress}"
                            ${rewardAmount === 0 ? 'disabled' : ''}
                        >
                            <i class="fas fa-gift"></i> Claim
                        </button>
                    </div>
                </div>
            `;
        }).join('');
    }

    setupEventListeners() {
        const claimAllBtn = document.getElementById('claim-all-btn');
        if (claimAllBtn) {
            claimAllBtn.addEventListener('click', () => this.claimAllRewards());
        }

        document.querySelectorAll('.claim-single-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const validatorAddress = e.target.closest('button').dataset.validator;
                this.claimRewards(validatorAddress);
            });
        });
    }

    async claimAllRewards() {
        // Check wallet connection
        if (!this.wallet || !this.wallet.connected) {
            showToast('Please connect your wallet first', 'warning');
            return;
        }

        const autoCompound = document.getElementById('auto-compound-checkbox')?.checked || false;
        const validatorAddresses = this.rewards.rewards
            .filter(r => r.reward.some(rw => rw.denom === 'uaura' && rw.amount > 0))
            .map(r => r.validatorAddress);

        if (validatorAddresses.length === 0) {
            showToast('No rewards to claim', 'info');
            return;
        }

        try {
            showLoading('Claiming all rewards...');

            // Claim all rewards in a single transaction
            const result = await this.wallet.claimAllRewards(validatorAddresses);

            if (autoCompound) {
                showLoading('Auto-compounding rewards...');
                await this.autoCompoundRewards();
            }

            hideLoading();
            this.showClaimSuccess(result, autoCompound);

            // Close modal after delay
            setTimeout(() => {
                document.getElementById('rewards-modal').classList.remove('active');
            }, 3000);

            // Refresh portfolio
            if (window.stakingDashboard?.components?.portfolio) {
                window.stakingDashboard.components.portfolio.refresh();
            }
        } catch (error) {
            hideLoading();
            console.error('Claim failed:', error);

            if (error instanceof WalletError) {
                if (error.code === 'USER_REJECTED') {
                    showToast('Transaction was rejected', 'warning');
                } else {
                    showToast(error.message, 'error');
                }
            } else {
                showToast(error.message || 'Failed to claim rewards', 'error');
            }
        }
    }

    async claimRewards(validatorAddress) {
        // Check wallet connection
        if (!this.wallet || !this.wallet.connected) {
            showToast('Please connect your wallet first', 'warning');
            return;
        }

        try {
            showLoading('Claiming rewards...');

            const result = await this.wallet.claimRewardsFromValidator(validatorAddress);

            hideLoading();
            showToast('Successfully claimed rewards!', 'success');

            // Show mini success in the button area
            const btn = document.querySelector(`[data-validator="${validatorAddress}"]`);
            if (btn) {
                btn.innerHTML = '<i class="fas fa-check"></i> Claimed';
                btn.disabled = true;
                btn.classList.remove('btn-primary');
                btn.classList.add('btn-secondary');
            }

            // Refresh rewards display after a short delay
            setTimeout(() => this.render(this.delegatorAddress), 1500);
        } catch (error) {
            hideLoading();
            console.error('Claim failed:', error);

            if (error instanceof WalletError) {
                if (error.code === 'USER_REJECTED') {
                    showToast('Transaction was rejected', 'warning');
                } else {
                    showToast(error.message, 'error');
                }
            } else {
                showToast(error.message || 'Failed to claim rewards', 'error');
            }
        }
    }

    async autoCompoundRewards() {
        // Get delegations to find the largest validator
        const delegations = await this.api.getDelegations(this.delegatorAddress);
        if (delegations.length === 0) {
            showToast('No existing delegations for auto-compound', 'warning');
            return;
        }

        // Find validator with largest delegation
        const largestDelegation = delegations.reduce((max, d) =>
            d.balance > max.balance ? d : max, delegations[0]
        );

        // Get current balance (should include newly claimed rewards)
        const balance = await this.wallet.getBalance();

        // Leave some for fees
        const amountToStake = Math.max(0, parseFloat(balance.amount) - 10000); // Keep 0.01 AURA for fees

        if (amountToStake <= 0) {
            showToast('Insufficient balance for auto-compound', 'warning');
            return;
        }

        // Delegate the claimed rewards
        await this.wallet.delegate(
            largestDelegation.validatorAddress,
            String(Math.floor(amountToStake)),
            'Auto-compounded rewards via Aura Dashboard'
        );
    }

    showClaimSuccess(result, compounded) {
        const container = document.getElementById('rewards-panel');
        container.innerHTML = `
            <div class="tx-success" style="text-align: center; padding: 2rem;">
                <i class="fas fa-check-circle" style="font-size: 3rem; color: var(--success-color); margin-bottom: 1rem;"></i>
                <h3 style="margin-bottom: 0.5rem;">Rewards Claimed!</h3>
                <p style="color: var(--text-secondary); margin-bottom: 1rem;">
                    ${compounded ? 'Rewards claimed and auto-compounded' : 'All pending rewards have been claimed'}
                </p>
                <div style="background: var(--bg-tertiary); padding: 0.75rem; border-radius: var(--radius); margin-bottom: 1rem;">
                    <div style="font-size: 0.75rem; color: var(--text-secondary); margin-bottom: 0.25rem;">Transaction Hash</div>
                    <code style="font-size: 0.8rem; word-break: break-all;">${result.transactionHash}</code>
                </div>
                <a href="${this.wallet.getExplorerUrl(result.transactionHash)}" target="_blank"
                   class="btn btn-secondary" style="width: 100%;">
                    <i class="fas fa-external-link-alt"></i> View in Explorer
                </a>
            </div>
        `;
    }
}

export default RewardsPanel;
