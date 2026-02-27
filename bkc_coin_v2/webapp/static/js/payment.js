// Multi-chain payment system for BKC Coin
class PaymentSystem {
    constructor() {
        this.apiBase = '/api/payments';
        this.currentOrder = null;
        this.pollingInterval = null;
        this.supportedChains = [];
        this.commissionInfo = null;
        
        this.init();
    }

    async init() {
        try {
            await this.loadSupportedChains();
            await this.loadCommissionInfo();
            this.setupEventListeners();
        } catch (error) {
            console.error('Failed to initialize payment system:', error);
        }
    }

    async loadSupportedChains() {
        try {
            const response = await fetch(`${this.apiBase}/chains`);
            const data = await response.json();
            this.supportedChains = data.chains;
            this.exchangeRates = data.rates;
        } catch (error) {
            console.error('Failed to load supported chains:', error);
        }
    }

    async loadCommissionInfo() {
        try {
            const response = await fetch(`${this.apiBase}/commission`);
            this.commissionInfo = await response.json();
        } catch (error) {
            console.error('Failed to load commission info:', error);
        }
    }

    setupEventListeners() {
        // Кнопки выбора цепочки
        document.querySelectorAll('.chain-option').forEach(button => {
            button.addEventListener('click', (e) => {
                this.selectChain(e.target.dataset.chain);
            });
        });

        // Кнопка создания платежа
        const createButton = document.getElementById('create-payment');
        if (createButton) {
            createButton.addEventListener('click', () => {
                this.createPayment();
            });
        }

        // Кнопка отмены платежа
        const cancelButton = document.getElementById('cancel-payment');
        if (cancelButton) {
            cancelButton.addEventListener('click', () => {
                this.cancelPayment();
            });
        }

        // Кнопка копирования ссылки
        document.addEventListener('click', (e) => {
            if (e.target.classList.contains('copy-link')) {
                this.copyToClipboard(e.target.dataset.url);
            }
        });

        // Кнопка оплаты
        document.addEventListener('click', (e) => {
            if (e.target.classList.contains('pay-button')) {
                this.initiatePayment(e.target.dataset.chain);
            }
        });
    }

    selectChain(chain) {
        // Обновляем UI
        document.querySelectorAll('.chain-option').forEach(btn => {
            btn.classList.remove('selected');
        });
        document.querySelector(`[data-chain="${chain}"]`).classList.add('selected');

        // Обновляем информацию о цепочке
        const chainInfo = this.supportedChains.find(c => c.id === chain);
        if (chainInfo) {
            this.updateChainInfo(chainInfo);
        }

        // Валидируем платеж
        this.validatePayment();
    }

    updateChainInfo(chainInfo) {
        const infoElement = document.getElementById('chain-info');
        if (infoElement) {
            infoElement.innerHTML = `
                <div class="chain-details">
                    <img src="${chainInfo.icon}" alt="${chainInfo.name}" class="chain-icon">
                    <div class="chain-text">
                        <h3>${chainInfo.name}</h3>
                        <p>${chainInfo.description}</p>
                        <small>Exchange Rate: 1 ${chainInfo.currency} = ${this.exchangeRates[chainInfo.id]} BKC</small>
                    </div>
                </div>
            `;
        }
    }

    async validatePayment() {
        const amount = parseFloat(document.getElementById('payment-amount').value);
        const chain = document.querySelector('.chain-option.selected')?.dataset.chain;
        const type = document.getElementById('payment-type').value;

        if (!amount || !chain || !type) {
            return;
        }

        try {
            const response = await fetch(`${this.apiBase}/validate`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    amount: amount,
                    chain: chain,
                    type: type,
                    currency: chain === 'ton' ? 'TON' : 'USDT'
                })
            });

            const data = await response.json();

            if (data.valid) {
                this.updatePaymentSummary(data);
            } else {
                this.showError(data.error);
            }
        } catch (error) {
            console.error('Validation failed:', error);
            this.showError('Validation failed');
        }
    }

    updatePaymentSummary(data) {
        const summaryElement = document.getElementById('payment-summary');
        if (summaryElement) {
            summaryElement.innerHTML = `
                <div class="summary-row">
                    <span>You will receive:</span>
                    <span class="highlight">${data.net_amount.toLocaleString()} BKC</span>
                </div>
                <div class="summary-row">
                    <span>Commission (${this.commissionInfo.rates.platform}%):</span>
                    <span>${data.commission.toLocaleString()} BKC</span>
                </div>
                <div class="summary-row total">
                    <span>Total to pay:</span>
                    <span class="total-amount">${data.bkc_amount.toLocaleString()} BKC</span>
                </div>
            `;
        }
    }

    async createPayment() {
        const amount = parseFloat(document.getElementById('payment-amount').value);
        const chain = document.querySelector('.chain-option.selected')?.dataset.chain;
        const type = document.getElementById('payment-type').value;

        if (!amount || !chain || !type) {
            this.showError('Please fill all fields');
            return;
        }

        try {
            const response = await fetch(`${this.apiBase}/create`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    amount: amount,
                    chain: chain,
                    type: type,
                    currency: chain === 'ton' ? 'TON' : 'USDT'
                })
            });

            const data = await response.json();

            if (response.ok) {
                this.currentOrder = data;
                this.showPaymentInterface(data);
                this.startStatusPolling(data.order_id);
            } else {
                this.showError(data.error);
            }
        } catch (error) {
            console.error('Failed to create payment:', error);
            this.showError('Failed to create payment');
        }
    }

    showPaymentInterface(paymentData) {
        const interfaceElement = document.getElementById('payment-interface');
        if (interfaceElement) {
            interfaceElement.innerHTML = `
                <div class="payment-container">
                    <h3>Payment Created</h3>
                    <div class="payment-details">
                        <p><strong>Order ID:</strong> ${paymentData.order_id}</p>
                        <p><strong>Amount:</strong> ${paymentData.payment_url.includes('ton') ? 
                            (parseFloat(document.getElementById('payment-amount').value) + ' TON') : 
                            (parseFloat(document.getElementById('payment-amount').value) + ' USDT')}</p>
                        <p><strong>Expires:</strong> ${new Date(paymentData.expires_at).toLocaleString()}</p>
                    </div>
                    
                    <div class="payment-methods">
                        <h4>Choose Payment Method:</h4>
                        <div class="method-buttons">
                            <button class="pay-button" data-chain="ton" onclick="paymentSystem.initiatePayment('ton')">
                                <img src="/icons/ton-wallet.png" alt="TON Wallet">
                                <span>TON Wallet</span>
                            </button>
                            <button class="pay-button" data-chain="solana" onclick="paymentSystem.initiatePayment('solana')">
                                <img src="/icons/solana-wallet.png" alt="Solana Wallet">
                                <span>Solana Wallet</span>
                            </button>
                        </div>
                    </div>
                    
                    <div class="payment-alternatives">
                        <h4>Or use direct link:</h4>
                        <div class="link-container">
                            <input type="text" value="${paymentData.payment_url}" readonly class="payment-link">
                            <button class="copy-link" data-url="${paymentData.payment_url}">Copy</button>
                        </div>
                    </div>
                    
                    <div class="qr-code-container">
                        <h4>QR Code:</h4>
                        <img src="https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(paymentData.qr_code)}" 
                             alt="QR Code" class="qr-code">
                    </div>
                    
                    <div class="instructions">
                        <h4>Instructions:</h4>
                        <ol>
                            ${Object.entries(paymentData.instructions).map(([key, value]) => 
                                key.startsWith('step') ? `<li>${value}</li>` : ''
                            ).join('')}
                        </ol>
                        ${paymentData.instructions.note ? `<p class="note">${paymentData.instructions.note}</p>` : ''}
                    </div>
                    
                    <div class="payment-actions">
                        <button id="cancel-payment" class="cancel-button">Cancel Payment</button>
                        <button class="refresh-button" onclick="paymentSystem.checkStatus()">Refresh Status</button>
                    </div>
                    
                    <div id="payment-status" class="payment-status">
                        <div class="status-indicator pending"></div>
                        <span>Waiting for payment...</span>
                    </div>
                </div>
            `;
        }
    }

    initiatePayment(chain) {
        if (!this.currentOrder) return;

        const paymentUrl = this.currentOrder.payment_url;
        
        if (chain === 'ton') {
            // Для TON используем deep link
            window.location.href = paymentUrl;
        } else if (chain === 'solana') {
            // Для Solana открываем универсальную ссылку
            window.location.href = paymentUrl;
        }

        // Показываем инструкцию
        this.showPaymentInstruction(chain);
    }

    showPaymentInstruction(chain) {
        const instructions = {
            ton: `
                <div class="payment-instruction">
                    <h4>TON Payment Instructions:</h4>
                    <ol>
                        <li>Your TON wallet should open automatically</li>
                        <li>Review the transaction details</li>
                        <li>Confirm the payment</li>
                        <li>Wait for confirmation (10-30 seconds)</li>
                    </ol>
                    <p>If wallet doesn't open, <a href="${this.currentOrder.payment_url}">click here</a></p>
                </div>
            `,
            solana: `
                <div class="payment-instruction">
                    <h4>Solana Payment Instructions:</h4>
                    <ol>
                        <li>Choose your wallet from the list (Phantom, Trust, MetaMask, etc.)</li>
                        <li>Review the USDT transfer details</li>
                        <li>Confirm the transaction in your wallet</li>
                        <li>Wait for confirmation (2-5 seconds)</li>
                    </ol>
                    <p>If nothing happens, <a href="${this.currentOrder.payment_url}">click here</a></p>
                </div>
            `
        };

        const instructionElement = document.getElementById('payment-instruction');
        if (instructionElement) {
            instructionElement.innerHTML = instructions[chain] || '';
        }
    }

    startStatusPolling(orderId) {
        this.pollingInterval = setInterval(async () => {
            await this.checkStatus();
        }, 5000); // Проверяем каждые 5 секунд
    }

    async checkStatus() {
        if (!this.currentOrder) return;

        try {
            const response = await fetch(`${this.apiBase}/status/${this.currentOrder.order_id}`);
            const order = await response.json();

            this.updatePaymentStatus(order);

            if (order.status === 'confirmed') {
                this.handlePaymentSuccess(order);
            } else if (order.status === 'expired') {
                this.handlePaymentExpired(order);
            }
        } catch (error) {
            console.error('Failed to check status:', error);
        }
    }

    updatePaymentStatus(order) {
        const statusElement = document.getElementById('payment-status');
        if (statusElement) {
            const statusClasses = {
                pending: 'pending',
                confirmed: 'confirmed',
                expired: 'expired',
                cancelled: 'cancelled'
            };

            const statusTexts = {
                pending: 'Waiting for payment...',
                confirmed: 'Payment confirmed!',
                expired: 'Payment expired',
                cancelled: 'Payment cancelled'
            };

            statusElement.innerHTML = `
                <div class="status-indicator ${statusClasses[order.status]}"></div>
                <span>${statusTexts[order.status]}</span>
            `;

            if (order.transaction_hash) {
                statusElement.innerHTML += `
                    <div class="transaction-hash">
                        <small>Transaction: ${order.transaction_hash}</small>
                    </div>
                `;
            }
        }
    }

    handlePaymentSuccess(order) {
        clearInterval(this.pollingInterval);
        
        // Показываем уведомление об успехе
        this.showSuccess(`Payment confirmed! You received ${order.net_amount.toLocaleString()} BKC`);
        
        // Обновляем баланс на странице
        this.updateBalance(order.net_amount);
        
        // Перенаправляем или закрываем модальное окно
        setTimeout(() => {
            this.closePaymentInterface();
        }, 3000);
    }

    handlePaymentExpired(order) {
        clearInterval(this.pollingInterval);
        this.showError('Payment expired. Please create a new payment.');
    }

    async cancelPayment() {
        if (!this.currentOrder) return;

        try {
            const response = await fetch(`${this.apiBase}/cancel/${this.currentOrder.order_id}`, {
                method: 'POST'
            });

            if (response.ok) {
                clearInterval(this.pollingInterval);
                this.closePaymentInterface();
                this.showInfo('Payment cancelled');
            }
        } catch (error) {
            console.error('Failed to cancel payment:', error);
        }
    }

    closePaymentInterface() {
        const interfaceElement = document.getElementById('payment-interface');
        if (interfaceElement) {
            interfaceElement.innerHTML = '';
        }
        
        this.currentOrder = null;
        clearInterval(this.pollingInterval);
    }

    updateBalance(amount) {
        // Обновляем баланс на странице
        const balanceElement = document.getElementById('user-balance');
        if (balanceElement) {
            const currentBalance = parseInt(balanceElement.textContent.replace(/[^0-9]/g, ''));
            const newBalance = currentBalance + amount;
            balanceElement.textContent = newBalance.toLocaleString() + ' BKC';
        }
    }

    copyToClipboard(text) {
        navigator.clipboard.writeText(text).then(() => {
            this.showSuccess('Link copied to clipboard!');
        }).catch(err => {
            console.error('Failed to copy:', err);
        });
    }

    showError(message) {
        this.showNotification(message, 'error');
    }

    showSuccess(message) {
        this.showNotification(message, 'success');
    }

    showInfo(message) {
        this.showNotification(message, 'info');
    }

    showNotification(message, type) {
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        notification.textContent = message;
        
        document.body.appendChild(notification);
        
        setTimeout(() => {
            notification.remove();
        }, 3000);
    }

    async estimatePayment(amount, chain, type) {
        try {
            const params = new URLSearchParams({
                amount: amount.toString(),
                chain: chain,
                type: type
            });

            const response = await fetch(`${this.apiBase}/estimate?${params}`);
            return await response.json();
        } catch (error) {
            console.error('Failed to estimate payment:', error);
            return null;
        }
    }
}

// Инициализация платежной системы
let paymentSystem;
document.addEventListener('DOMContentLoaded', () => {
    paymentSystem = new PaymentSystem();
});

// Экспорт для использования в других модулях
window.paymentSystem = paymentSystem;
