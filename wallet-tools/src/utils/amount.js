const DECIMAL_REGEX = /^\d+(\.\d+)?$/;

export function toMicroAmount(amount, decimals = 6) {
  if (typeof amount !== 'string') {
    throw new Error('amount must be a string');
  }
  if (!DECIMAL_REGEX.test(amount)) {
    throw new Error(`Invalid amount "${amount}". Use digits optionally with a decimal point.`);
  }

  const [whole, fraction = ''] = amount.split('.');
  if (fraction.length > decimals) {
    throw new Error(`Amount precision exceeds ${decimals} decimal places.`);
  }

  const normalizedFraction = fraction.padEnd(decimals, '0');
  const normalized = `${whole}${normalizedFraction}`.replace(/^0+/, '') || '0';
  return normalized;
}

export function formatCoins(amountMicro, denom, decimals = 6, displayDenom = 'AURA') {
  const amountStr = amountMicro.toString();
  const padded = amountStr.padStart(decimals + 1, '0');
  const whole = padded.slice(0, -decimals) || '0';
  const fraction = padded.slice(-decimals).replace(/0+$/, '');
  const value = fraction ? `${whole}.${fraction}` : whole;
  return `${value} ${displayDenom} (${amountMicro}${denom})`;
}
