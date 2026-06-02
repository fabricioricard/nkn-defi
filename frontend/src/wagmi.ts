import { http, createConfig } from 'wagmi'
import { base, baseSepolia } from 'wagmi/chains'
import { metaMask } from 'wagmi/connectors'

export const config = createConfig({
  chains: [base, baseSepolia],   // Base mainnet e testnet
  connectors: [metaMask()],
  transports: {
    [base.id]: http(),
    [baseSepolia.id]: http(),
  },
})