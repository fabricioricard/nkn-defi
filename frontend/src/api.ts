const BASE = '/api/v1';

async function fetchJSON(url: string, options?: RequestInit) {
  const res = await fetch(`${BASE}${url}`, {
    headers: { 'Content-Type': 'application/json', ...options?.headers },
    ...options,
  });
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  return res.json();
}

export const api = {
  // Bridge
  createBridgeDeposit: (eth_address: string, amount: string, tx_hash?: string) =>
    fetchJSON('/bridge/deposit', {
      method: 'POST',
      body: JSON.stringify({ eth_address, amount, tx_hash }),
    }),

  getBridgeTransactions: (ethAddress: string) =>
    fetchJSON(`/bridge/transactions?eth_address=${encodeURIComponent(ethAddress)}`),

  cancelDeposit: (id: string) =>
    fetchJSON(`/bridge/deposit/${id}`, {
      method: 'DELETE',
    }),

  // Pools
  getPools: () => fetchJSON('/pools'),
  createPool: (token0: string, token1: string, fee_bps: number) =>
    fetchJSON('/pools', {
      method: 'POST',
      body: JSON.stringify({ token0, token1, fee_bps }),
    }),
  addLiquidity: (poolId: string, amount0: string, amount1: string) =>
    fetchJSON(`/pools/${poolId}/liquidity/add`, {
      method: 'POST',
      body: JSON.stringify({ amount0, amount1 }),
    }),
  swap: (poolId: string, token_in: string, amount_in: string) =>
    fetchJSON(`/pools/${poolId}/swap`, {
      method: 'POST',
      body: JSON.stringify({ token_in, amount_in }),
    }),
};