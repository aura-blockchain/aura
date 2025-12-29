// AURA Staking Dashboard - Main Application
// Integrates with Keplr wallet for real transaction signing

import { StakingAPI } from './services/stakingAPI.js';
import { ValidatorList } from './components/ValidatorList.js';
import { ValidatorComparison } from './components/ValidatorComparison.js';
import { StakingCalculator } from './components/StakingCalculator.js';
import { DelegationPanel } from './components/DelegationPanel.js';
import { RewardsPanel } from './components/RewardsPanel.js';
import { PortfolioView } from './components/PortfolioView.js';
import { showToast, showLoading, hideLoading } from './utils/ui.js';
import { AuraWalletConnector, WalletError, WalletErrorCodes } from '../wallet-connector.js';

class StakingDashboard {
    constructor() {
        this.api = new StakingAPI();
        this.wallet = new AuraWalletConnector();
        this.walletAddress = null;
        this.components = {};
        this.init();
    }

    async init() {
        this.setupEventListeners();
        this.setupWalletEvents();
        this.initializeComponents();
        await this.loadNetworkStats();
        this.checkWalletConnection();
    }

    setupEventListeners() {
        // Navigation
        document.querySelectorAll('.nav-item').forEach(item => {
            item.addEventListener('click', (e) => this.handleNavigation(e));
        });

        // Wallet Connection
        document.getElementById('connect-wallet')?.addEventListener('click', () => this.connectWallet());
        document.getElementById('disconnect-wallet')?.addEventListener('click', () => this.disconnectWallet());

        // Modal Close Buttons
        document.querySelectorAll('.modal-close').forEach(btn => {
            btn.addEventListener('click', (e) => this.closeModal(e.target.closest('.modal')));
        });

        // Click outside modal to close
        document.querySelectorAll('.modal').forEach(modal => {
            modal.addEventListener('click', (e) => {
                if (e.target === modal) {
                    this.closeModal(modal);
                }
            });
        });

        // Refresh Portfolio
        document.getElementById('refresh-portfolio')?.addEventListener('click', () => {
            if (this.components.portfolio) {
                this.components.portfolio.refresh();
            }
        });
    }

    setupWalletEvents() {
        // Handle wallet events
        this.wallet.on('connect', ({ address }) => {
            console.log('Wallet connected:', address);
        });

        this.wallet.on('disconnect', () => {
            console.log('Wallet disconnected');
            this.handleWalletDisconnect();
        });

        this.wallet.on('accountChange', async ({ address }) => {
            console.log('Account changed:', address);
            this.walletAddress = address;
            this.updateWalletUI(address);
            await this.components.portfolio?.render(address);
            showToast('Account changed', 'info');
        });
    }

    initializeComponents() {
        // Initialize all components with wallet reference
        this.components = {
            validatorList: new ValidatorList(this.api),
            validatorComparison: new ValidatorComparison(this.api),
            stakingCalculator: new StakingCalculator(this.api),
            delegationPanel: new DelegationPanel(this.api, this.wallet),
            rewardsPanel: new RewardsPanel(this.api, this.wallet),
            portfolio: new PortfolioView(this.api, this.wallet)
        };

        // Set up component event listeners
        this.components.validatorList.on('delegate', (validator) => this.showDelegationModal(validator));
        this.components.portfolio.on('claim-rewards', () => this.showRewardsModal());
        this.components.portfolio.on('delegate', (validator) => this.showDelegationModal(validator));
    }

    handleNavigation(e) {
        const view = e.currentTarget.dataset.view;

        // Update nav active state
        document.querySelectorAll('.nav-item').forEach(item => {
            item.classList.remove('active');
        });
        e.currentTarget.classList.add('active');

        // Update view visibility
        document.querySelectorAll('.view-container').forEach(container => {
            container.classList.remove('active');
        });
        document.getElementById(`${view}-view`)?.classList.add('active');

        // Trigger component refresh if needed
        switch(view) {
            case 'validators':
                this.components.validatorList.render();
                break;
            case 'calculator':
                this.components.stakingCalculator.render();
                break;
            case 'comparison':
                this.components.validatorComparison.render();
                break;
            case 'portfolio':
                if (this.walletAddress) {
                    this.components.portfolio.render(this.walletAddress);
                }
                break;
        }
    }

    async loadNetworkStats() {
        try {
            const stats = await this.api.getNetworkStats();

            document.getElementById('total-staked').textContent =
                `${this.formatNumber(stats.totalStaked)} AURA`;
            document.getElementById('avg-apy').textContent =
                `${stats.averageAPY.toFixed(2)}%`;
            document.getElementById('active-validators').textContent =
                stats.activeValidators.toString();
            document.getElementById('inflation-rate').textContent =
                `${stats.inflationRate.toFixed(2)}%`;
        } catch (error) {
            console.error('Failed to load network stats:', error);
            showToast('Failed to load network statistics', 'error');
        }
    }

    async connectWallet() {
        try {
            showLoading('Connecting wallet...');

            // Check if Keplr is installed
            if (!AuraWalletConnector.isKeplrInstalled()) {
                hideLoading();
                this.showInstallKeplrPrompt();
                return;
            }

            // Connect using the wallet connector
            this.walletAddress = await this.wallet.connect();

            // Update UI
            this.updateWalletUI(this.walletAddress);

            // Save to localStorage for reconnection
            localStorage.setItem('aura_wallet_connected', 'true');

            // Load portfolio
            await this.components.portfolio.render(this.walletAddress);

            hideLoading();
            showToast('Wallet connected successfully', 'success');
        } catch (error) {
            hideLoading();
            console.error('Wallet connection failed:', error);

            if (error instanceof WalletError) {
                switch (error.code) {
                    case WalletErrorCodes.KEPLR_NOT_INSTALLED:
                        this.showInstallKeplrPrompt();
                        break;
                    case WalletErrorCodes.USER_REJECTED:
                        showToast('Connection request was rejected', 'warning');
                        break;
                    case WalletErrorCodes.NO_ACCOUNTS:
                        showToast('No accounts found in wallet', 'error');
                        break;
                    default:
                        showToast(error.message, 'error');
                }
            } else {
                showToast(error.message || 'Failed to connect wallet', 'error');
            }
        }
    }

    showInstallKeplrPrompt() {
        const modal = document.createElement('div');
        modal.className = 'modal active';
        modal.id = 'keplr-install-modal';
        modal.innerHTML = `
            <div class="modal-content" style="max-width: 400px;">
                <div class="modal-header">
                    <h3>Install Keplr Wallet</h3>
                    <button class="modal-close">&times;</button>
                </div>
                <div class="modal-body" style="text-align: center; padding: 2rem;">
                    <i class="fas fa-wallet" style="font-size: 3rem; color: var(--primary-color); margin-bottom: 1rem;"></i>
                    <p style="margin-bottom: 1.5rem;">
                        Keplr wallet is required to interact with the Aura blockchain.
                        Please install the Keplr browser extension to continue.
                    </p>
                    <a href="https://www.keplr.app" target="_blank" rel="noopener noreferrer"
                       class="btn btn-primary" style="width: 100%;">
                        <i class="fas fa-external-link-alt"></i> Install Keplr
                    </a>
                    <button class="btn btn-secondary" style="width: 100%; margin-top: 0.5rem;"
                            onclick="this.closest('.modal').remove();">
                        Cancel
                    </button>
                </div>
            </div>
        `;
        document.body.appendChild(modal);

        modal.querySelector('.modal-close').addEventListener('click', () => modal.remove());
        modal.addEventListener('click', (e) => {
            if (e.target === modal) modal.remove();
        });
    }

    updateWalletUI(address) {
        document.getElementById('connect-wallet').style.display = 'none';
        const walletConnected = document.getElementById('wallet-connected');
        walletConnected.style.display = 'flex';
        walletConnected.querySelector('.wallet-address').textContent =
            AuraWalletConnector.formatAddress(address);
    }

    disconnectWallet() {
        this.wallet.disconnect();
        this.walletAddress = null;
        localStorage.removeItem('aura_wallet_connected');
        this.handleWalletDisconnect();
        showToast('Wallet disconnected', 'info');
    }

    handleWalletDisconnect() {
        document.getElementById('connect-wallet').style.display = 'flex';
        document.getElementById('wallet-connected').style.display = 'none';

        // Clear portfolio view
        document.getElementById('portfolio-content').innerHTML =
            '<p class="text-center">Please connect your wallet to view your staking portfolio.</p>';
    }

    async checkWalletConnection() {
        const wasConnected = localStorage.getItem('aura_wallet_connected');
        if (wasConnected && AuraWalletConnector.isKeplrInstalled()) {
            // Auto-reconnect
            try {
                await this.connectWallet();
            } catch (error) {
                // Silent fail - user can reconnect manually
                console.log('Auto-reconnect failed:', error.message);
                localStorage.removeItem('aura_wallet_connected');
            }
        }
    }

    showDelegationModal(validator) {
        if (!this.walletAddress) {
            showToast('Please connect your wallet first', 'warning');
            return;
        }

        const modal = document.getElementById('delegation-modal');
        modal.classList.add('active');
        this.components.delegationPanel.render(validator, this.walletAddress);
    }

    showRewardsModal() {
        if (!this.walletAddress) {
            showToast('Please connect your wallet first', 'warning');
            return;
        }

        const modal = document.getElementById('rewards-modal');
        modal.classList.add('active');
        this.components.rewardsPanel.render(this.walletAddress);
    }

    closeModal(modal) {
        modal.classList.remove('active');
    }

    formatAddress(address) {
        return AuraWalletConnector.formatAddress(address);
    }

    formatNumber(num) {
        if (num >= 1e9) return `${(num / 1e9).toFixed(2)}B`;
        if (num >= 1e6) return `${(num / 1e6).toFixed(2)}M`;
        if (num >= 1e3) return `${(num / 1e3).toFixed(2)}K`;
        return num.toFixed(2);
    }

    /**
     * Get the wallet connector instance (for external access)
     */
    getWallet() {
        return this.wallet;
    }
}

// Initialize app when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    window.stakingDashboard = new StakingDashboard();
});

export default StakingDashboard;
