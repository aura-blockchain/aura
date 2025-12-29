// Delegation Panel Component
// Handles delegate, undelegate, and redelegate transactions with Keplr wallet

import { formatAmount, validateAmount, showToast, showLoading, hideLoading } from '../utils/ui.js';
import { AuraWalletConnector, WalletError } from '../../wallet-connector.js';

export class DelegationPanel {
    constructor(api, wallet) {
        this.api = api;
        this.wallet = wallet;
        this.validator = null;
        this.delegatorAddress = null;
        this.actionType = 'delegate'; // delegate, undelegate, redelegate
    }

    async render(validator, delegatorAddress) {
        this.validator = validator;
        this.delegatorAddress = delegatorAddress;

        const container = document.getElementById('delegation-panel');
        if (!container) return;

        const balance = await this.api.getBalance(delegatorAddress);
        const delegations = await this.api.getDelegations(delegatorAddress);
        const currentDelegation = delegations.find(d =>
            d.validatorAddress === validator.operatorAddress
        );

        container.innerHTML = `
            <div class="delegation-info" style="margin-bottom: 1.5rem;">
                <h4>${validator.moniker}</h4>
                <p style="color: var(--text-secondary); font-size: 0.875rem;">
                    ${validator.operatorAddress}
                </p>
                <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; margin-top: 1rem;">
                    <div>
                        <div style="font-size: 0.875rem; color: var(--text-secondary);">Commission</div>
                        <div style="font-weight: 600;">${validator.commission.toFixed(2)}%</div>
                    </div>
                    <div>
                        <div style="font-size: 0.875rem; color: var(--text-secondary);">Your Delegation</div>
                        <div style="font-weight: 600;">
                            ${currentDelegation ? formatAmount(currentDelegation.balance * 1e6) : '0'} AURA
                        </div>
                    </div>
                </div>
            </div>

            <div class="action-selector" style="margin-bottom: 1.5rem;">
                <div style="display: flex; gap: 0.5rem;">
                    <button class="btn btn-primary action-btn active" data-action="delegate">
                        <i class="fas fa-plus"></i> Delegate
                    </button>
                    <button class="btn btn-secondary action-btn" data-action="undelegate" ${!currentDelegation ? 'disabled' : ''}>
                        <i class="fas fa-minus"></i> Undelegate
                    </button>
                    <button class="btn btn-secondary action-btn" data-action="redelegate" ${!currentDelegation ? 'disabled' : ''}>
                        <i class="fas fa-exchange-alt"></i> Redelegate
                    </button>
                </div>
            </div>

            <form id="delegation-form">
                <div id="delegate-form">
                    <div class="form-group">
                        <label for="delegate-amount">Amount to Delegate</label>
                        <div class="input-group">
                            <input
                                type="number"
                                id="delegate-amount"
                                placeholder="0.00"
                                min="0"
                                step="0.000001"
                                required
                            >
                            <button type="button" class="btn btn-secondary max-btn" id="delegate-max-btn">MAX</button>
                        </div>
                        <small>Available: ${formatAmount(balance * 1e6)} AURA</small>
                    </div>
                </div>

                <div id="undelegate-form" style="display: none;">
                    <div class="form-group">
                        <label for="undelegate-amount">Amount to Undelegate</label>
                        <div class="input-group">
                            <input
                                type="number"
                                id="undelegate-amount"
                                placeholder="0.00"
                                min="0"
                                step="0.000001"
                                required
                            >
                            <button type="button" class="btn btn-secondary max-btn" id="undelegate-max-btn">MAX</button>
                        </div>
                        <small>
                            Delegated: ${currentDelegation ? formatAmount(currentDelegation.balance * 1e6) : '0'} AURA
                        </small>
                    </div>
                    <div style="background: var(--bg-tertiary); padding: 1rem; border-radius: var(--radius); margin-top: 1rem;">
                        <i class="fas fa-info-circle"></i>
                        <small>Unbonding period: 21 days. Tokens will not earn rewards during this time.</small>
                    </div>
                </div>

                <div id="redelegate-form" style="display: none;">
                    <div class="form-group">
                        <label for="redelegate-validator">New Validator</label>
                        <select id="redelegate-validator" class="select-input" required>
                            <option value="">-- Select validator --</option>
                        </select>
                    </div>
                    <div class="form-group">
                        <label for="redelegate-amount">Amount to Redelegate</label>
                        <div class="input-group">
                            <input
                                type="number"
                                id="redelegate-amount"
                                placeholder="0.00"
                                min="0"
                                step="0.000001"
                                required
                            >
                            <button type="button" class="btn btn-secondary max-btn" id="redelegate-max-btn">MAX</button>
                        </div>
                        <small>
                            Delegated: ${currentDelegation ? formatAmount(currentDelegation.balance * 1e6) : '0'} AURA
                        </small>
                    </div>
                    <div style="background: var(--bg-tertiary); padding: 1rem; border-radius: var(--radius); margin-top: 1rem;">
                        <i class="fas fa-info-circle"></i>
                        <small>Redelegation is immediate but you cannot redelegate again from the destination validator for 21 days.</small>
                    </div>
                </div>

                <div id="estimation" style="background: var(--bg-tertiary); padding: 1rem; border-radius: var(--radius); margin: 1rem 0;">
                    <div style="font-weight: 600; margin-bottom: 0.5rem;">Estimated Annual Rewards</div>
                    <div id="estimated-rewards" style="font-size: 1.25rem; color: var(--success-color); font-weight: bold;">
                        0 AURA
                    </div>
                </div>

                <div id="tx-fee-estimate" style="display: flex; justify-content: space-between; padding: 0.5rem 0; font-size: 0.875rem; color: var(--text-secondary); border-top: 1px solid var(--border-color); margin-bottom: 1rem;">
                    <span>Estimated Fee:</span>
                    <span id="fee-amount">~0.006 AURA</span>
                </div>

                <button type="submit" class="btn btn-primary" style="width: 100%;" id="submit-btn">
                    <i class="fas fa-paper-plane"></i> Submit Transaction
                </button>
            </form>
        `;

        this.setupFormHandlers(balance, currentDelegation);
    }

    setupFormHandlers(balance, currentDelegation) {
        const form = document.getElementById('delegation-form');
        const actionBtns = document.querySelectorAll('.action-btn');

        // Action type switching
        actionBtns.forEach(btn => {
            btn.addEventListener('click', (e) => {
                actionBtns.forEach(b => {
                    b.classList.remove('btn-primary', 'active');
                    b.classList.add('btn-secondary');
                });
                e.target.classList.remove('btn-secondary');
                e.target.classList.add('btn-primary', 'active');

                this.actionType = e.target.dataset.action;
                this.showActionForm(this.actionType);
            });
        });

        // Max buttons
        document.getElementById('delegate-max-btn')?.addEventListener('click', () => {
            // Leave some for fees
            const maxAmount = Math.max(0, balance - 0.01);
            document.getElementById('delegate-amount').value = maxAmount.toFixed(6);
            this.updateEstimation();
        });

        document.getElementById('undelegate-max-btn')?.addEventListener('click', () => {
            if (currentDelegation) {
                document.getElementById('undelegate-amount').value = currentDelegation.balance.toFixed(6);
                this.updateEstimation();
            }
        });

        document.getElementById('redelegate-max-btn')?.addEventListener('click', () => {
            if (currentDelegation) {
                document.getElementById('redelegate-amount').value = currentDelegation.balance.toFixed(6);
                this.updateEstimation();
            }
        });

        // Amount input estimation
        const delegateAmount = document.getElementById('delegate-amount');
        const undelegateAmount = document.getElementById('undelegate-amount');
        const redelegateAmount = document.getElementById('redelegate-amount');

        [delegateAmount, undelegateAmount, redelegateAmount].forEach(input => {
            if (input) {
                input.addEventListener('input', () => this.updateEstimation());
            }
        });

        // Form submission
        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            await this.submitTransaction(balance, currentDelegation);
        });

        // Load validators for redelegation
        this.loadValidatorsForRedelegate();
    }

    async loadValidatorsForRedelegate() {
        const select = document.getElementById('redelegate-validator');
        if (!select) return;

        const validators = await this.api.getValidators();
        const otherValidators = validators.filter(v =>
            v.operatorAddress !== this.validator.operatorAddress &&
            v.status === 'BOND_STATUS_BONDED'
        );

        otherValidators.forEach(v => {
            const option = document.createElement('option');
            option.value = v.operatorAddress;
            option.textContent = `${v.moniker} - ${v.commission.toFixed(2)}% commission`;
            select.appendChild(option);
        });
    }

    showActionForm(action) {
        document.getElementById('delegate-form').style.display =
            action === 'delegate' ? 'block' : 'none';
        document.getElementById('undelegate-form').style.display =
            action === 'undelegate' ? 'block' : 'none';
        document.getElementById('redelegate-form').style.display =
            action === 'redelegate' ? 'block' : 'none';

        const submitBtn = document.getElementById('submit-btn');
        const icons = {
            delegate: 'fa-paper-plane',
            undelegate: 'fa-minus-circle',
            redelegate: 'fa-exchange-alt'
        };
        const labels = {
            delegate: 'Delegate',
            undelegate: 'Undelegate',
            redelegate: 'Redelegate'
        };
        submitBtn.innerHTML = `<i class="fas ${icons[action]}"></i> ${labels[action]}`;

        // Update fee estimate based on action
        const feeEstimates = {
            delegate: '~0.006 AURA',
            undelegate: '~0.006 AURA',
            redelegate: '~0.009 AURA'
        };
        document.getElementById('fee-amount').textContent = feeEstimates[action];

        this.updateEstimation();
    }

    updateEstimation() {
        const estimationEl = document.getElementById('estimated-rewards');
        if (!estimationEl) return;

        let amount = 0;
        if (this.actionType === 'delegate') {
            amount = parseFloat(document.getElementById('delegate-amount')?.value || 0);
        } else if (this.actionType === 'undelegate') {
            amount = -parseFloat(document.getElementById('undelegate-amount')?.value || 0);
        } else if (this.actionType === 'redelegate') {
            amount = parseFloat(document.getElementById('redelegate-amount')?.value || 0);
        }

        const apy = this.api.calculateAPY(this.validator, 7.5);
        const annualRewards = Math.abs(amount) * (apy / 100);

        if (this.actionType === 'undelegate') {
            estimationEl.textContent = `-${formatAmount(annualRewards * 1e6)} AURA/year`;
            estimationEl.style.color = 'var(--danger-color)';
        } else {
            estimationEl.textContent = `+${formatAmount(annualRewards * 1e6)} AURA/year`;
            estimationEl.style.color = 'var(--success-color)';
        }
    }

    async submitTransaction(balance, currentDelegation) {
        let amount = 0;
        let error = null;

        // Validate based on action type
        if (this.actionType === 'delegate') {
            amount = parseFloat(document.getElementById('delegate-amount').value);
            error = validateAmount(amount, balance);
        } else if (this.actionType === 'undelegate') {
            amount = parseFloat(document.getElementById('undelegate-amount').value);
            const delegated = currentDelegation ? currentDelegation.balance : 0;
            error = validateAmount(amount, delegated);
        } else if (this.actionType === 'redelegate') {
            const newValidator = document.getElementById('redelegate-validator').value;
            if (!newValidator) {
                error = 'Please select a validator';
            } else {
                amount = parseFloat(document.getElementById('redelegate-amount').value);
                const delegated = currentDelegation ? currentDelegation.balance : 0;
                error = validateAmount(amount, delegated);
            }
        }

        if (error) {
            showToast(error, 'error');
            return;
        }

        // Check wallet connection
        if (!this.wallet || !this.wallet.connected) {
            showToast('Please connect your wallet first', 'warning');
            return;
        }

        try {
            showLoading(`Processing ${this.actionType} transaction...`);

            // Convert amount to micro units (uaura)
            const microAmount = AuraWalletConnector.toMicroUnits(amount);
            let result;

            switch (this.actionType) {
                case 'delegate':
                    result = await this.wallet.delegate(
                        this.validator.operatorAddress,
                        microAmount
                    );
                    break;

                case 'undelegate':
                    result = await this.wallet.undelegate(
                        this.validator.operatorAddress,
                        microAmount
                    );
                    break;

                case 'redelegate':
                    const newValidator = document.getElementById('redelegate-validator').value;
                    result = await this.wallet.redelegate(
                        this.validator.operatorAddress,
                        newValidator,
                        microAmount
                    );
                    break;
            }

            hideLoading();

            // Show success with transaction link
            this.showTransactionSuccess(result, amount);

            // Close modal after a delay
            setTimeout(() => {
                document.getElementById('delegation-modal').classList.remove('active');
            }, 3000);

            // Refresh portfolio
            if (window.stakingDashboard?.components?.portfolio) {
                window.stakingDashboard.components.portfolio.refresh();
            }
        } catch (error) {
            hideLoading();
            console.error('Transaction failed:', error);

            if (error instanceof WalletError) {
                if (error.code === 'USER_REJECTED') {
                    showToast('Transaction was rejected', 'warning');
                } else {
                    showToast(error.message, 'error');
                }
            } else {
                showToast(error.message || 'Transaction failed', 'error');
            }
        }
    }

    showTransactionSuccess(result, amount) {
        const actionLabels = {
            delegate: 'Delegated',
            undelegate: 'Undelegated',
            redelegate: 'Redelegated'
        };

        // Create success notification with explorer link
        const container = document.getElementById('delegation-panel');
        const successHtml = `
            <div class="tx-success" style="text-align: center; padding: 2rem;">
                <i class="fas fa-check-circle" style="font-size: 3rem; color: var(--success-color); margin-bottom: 1rem;"></i>
                <h3 style="margin-bottom: 0.5rem;">Transaction Successful!</h3>
                <p style="color: var(--text-secondary); margin-bottom: 1rem;">
                    ${actionLabels[this.actionType]} ${amount.toFixed(6)} AURA
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
        container.innerHTML = successHtml;

        showToast(`Successfully ${actionLabels[this.actionType].toLowerCase()} ${amount.toFixed(6)} AURA`, 'success');
    }
}

export default DelegationPanel;
