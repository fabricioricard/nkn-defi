import { useState, useEffect } from 'react';
import {
  VStack, Heading, Input, Button, HStack, Text,
  Select, useToast, Alert, AlertIcon,
} from '@chakra-ui/react';
import {
  useAccount, useWriteContract, useWaitForTransactionReceipt,
  useChainId, useReadContract,
} from 'wagmi';
import { parseUnits } from 'viem';

// ABI mínimo para ERC-20 (approve)
const erc20ABI = [
  {
    inputs: [
      { name: 'spender', type: 'address' },
      { name: 'amount', type: 'uint256' },
    ],
    name: 'approve',
    outputs: [{ name: '', type: 'bool' }],
    stateMutability: 'nonpayable',
    type: 'function',
  },
];

// ABI do SwapRouter02
const swapRouterABI = [
  {
    inputs: [
      {
        components: [
          { name: 'tokenIn', type: 'address' },
          { name: 'tokenOut', type: 'address' },
          { name: 'fee', type: 'uint24' },
          { name: 'recipient', type: 'address' },
          { name: 'deadline', type: 'uint256' },
          { name: 'amountIn', type: 'uint256' },
          { name: 'amountOutMinimum', type: 'uint256' },
          { name: 'sqrtPriceLimitX96', type: 'uint160' },
        ],
        name: 'params',
        type: 'tuple',
      },
    ],
    name: 'exactInputSingle',
    outputs: [{ name: 'amountOut', type: 'uint256' }],
    stateMutability: 'nonpayable',
    type: 'function',
  },
];

// ABI mínima da pool
const poolABI = [
  {
    inputs: [],
    name: 'token0',
    outputs: [{ name: '', type: 'address' }],
    stateMutability: 'view',
    type: 'function',
  },
  {
    inputs: [],
    name: 'token1',
    outputs: [{ name: '', type: 'address' }],
    stateMutability: 'view',
    type: 'function',
  },
  {
    inputs: [],
    name: 'slot0',
    outputs: [
      { name: 'sqrtPriceX96', type: 'uint160' },
      { name: 'tick', type: 'int24' },
      { name: 'observationIndex', type: 'uint16' },
      { name: 'observationCardinality', type: 'uint16' },
      { name: 'observationCardinalityNext', type: 'uint16' },
      { name: 'feeProtocol', type: 'uint8' },
      { name: 'unlocked', type: 'bool' },
    ],
    stateMutability: 'view',
    type: 'function',
  },
];

const WNKN_ADDRESS = '0x1B24ED102b530887B4388b61FB612121f6eD635E';
const USDC_ADDRESS = '0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913';
const SWAP_ROUTER = '0x2626664c2603336E57B271c5C0b26F421741e481';
const POOL_ADDRESS = '0xc7A1E13a0128f96994A0020Fb105f6BB070eF94a';
const POOL_FEE = 500; // 0,05%
const BASE_CHAIN_ID = 8453;
const SLIPPAGE = 0.5; // percentual fixo

export default function Swap() {
  const { address, isConnected } = useAccount();
  const chainId = useChainId();
  const toast = useToast();
  const { writeContractAsync: writeApprove, isPending: isApproving } = useWriteContract();
  const { writeContractAsync: writeSwap, isPending: isSwapping } = useWriteContract();

  const [tokenIn, setTokenIn] = useState<'wNKN' | 'USDC'>('wNKN');
  const [tokenOut, setTokenOut] = useState<'wNKN' | 'USDC'>('USDC');
  const [amountIn, setAmountIn] = useState('');
  const [estimatedOut, setEstimatedOut] = useState('');
  const [txHash, setTxHash] = useState<`0x${string}` | undefined>(undefined);

  const { isLoading: isWaiting } = useWaitForTransactionReceipt({ hash: txHash });

  const isWrongNetwork = isConnected && chainId !== BASE_CHAIN_ID;

  // Leitura do slot0 da pool para estimativa de preço
  const { data: slot0Data } = useReadContract({
    address: POOL_ADDRESS,
    abi: poolABI,
    functionName: 'slot0',
  });

  const { data: token0 } = useReadContract({
    address: POOL_ADDRESS,
    abi: poolABI,
    functionName: 'token0',
  });

  const { data: token1 } = useReadContract({
    address: POOL_ADDRESS,
    abi: poolABI,
    functionName: 'token1',
  });

  useEffect(() => {
    if (!slot0Data || !amountIn || !token0 || !token1) {
      setEstimatedOut('');
      return;
    }

    try {
      const slot0 = slot0Data as [bigint, number, number, number, number, number, boolean];
      const sqrtPriceX96 = slot0[0];
      const Q96 = 2n ** 96n;
      const priceRaw = (Number(sqrtPriceX96) / Number(Q96)) ** 2;

      const wNKNisToken0 = (token0 as string).toLowerCase() === WNKN_ADDRESS.toLowerCase();

      let price;
      if (wNKNisToken0) {
        // wNKN (token0, 18 dec) / USDC (token1, 6 dec)
        // price = sqrtPrice^2 / 10^(18-6) = priceRaw / 10^12
        price = priceRaw / 10 ** 12;
      } else {
        // USDC (token0, 6 dec) / wNKN (token1, 18 dec)
        // price of wNKN in USDC = 1 / (priceRaw / 10^12) = 10^12 / priceRaw
        price = priceRaw > 0 ? 10 ** 12 / priceRaw : 0;
      }

      const amount = parseFloat(amountIn);
      if (tokenIn === 'wNKN' && tokenOut === 'USDC') {
        setEstimatedOut((amount * price).toFixed(6));
      } else if (tokenIn === 'USDC' && tokenOut === 'wNKN') {
        setEstimatedOut(price > 0 ? (amount / price).toFixed(6) : '');
      } else {
        setEstimatedOut('');
      }
    } catch (e) {
      console.error('Price estimation error:', e);
      setEstimatedOut('');
    }
  }, [slot0Data, amountIn, tokenIn, tokenOut, token0, token1]);

  const handleApprove = async () => {
    if (!isConnected || isWrongNetwork) {
      toast({ title: 'Connect wallet and switch to Base', status: 'warning' });
      return;
    }
    if (!amountIn || parseFloat(amountIn) <= 0) {
      toast({ title: 'Enter a valid amount', status: 'warning' });
      return;
    }

    const tokenAddress = tokenIn === 'wNKN' ? WNKN_ADDRESS : USDC_ADDRESS;
    const amountWei = parseUnits(amountIn, tokenIn === 'wNKN' ? 18 : 6);

    try {
      await writeApprove({
        address: tokenAddress,
        abi: erc20ABI,
        functionName: 'approve',
        args: [SWAP_ROUTER, amountWei],
      });
      toast({ title: 'Approval submitted', status: 'info' });
    } catch (e: any) {
      toast({ title: 'Approval failed', description: e.message, status: 'error' });
    }
  };

  const handleSwap = async () => {
    if (!isConnected || isWrongNetwork) {
      toast({ title: 'Connect wallet and switch to Base', status: 'warning' });
      return;
    }
    if (!amountIn || parseFloat(amountIn) <= 0) {
      toast({ title: 'Enter a valid amount', status: 'warning' });
      return;
    }

    const tokenInAddress = tokenIn === 'wNKN' ? WNKN_ADDRESS : USDC_ADDRESS;
    const tokenOutAddress = tokenOut === 'wNKN' ? WNKN_ADDRESS : USDC_ADDRESS;
    const amountInWei = parseUnits(amountIn, tokenIn === 'wNKN' ? 18 : 6);
    const deadline = Math.floor(Date.now() / 1000) + 60 * 20;

    // Calcula amountOutMinimum baseado na estimativa e slippage
    let amountOutMinimum = 0n;
    if (estimatedOut) {
      const estBig = parseUnits(estimatedOut, tokenOut === 'wNKN' ? 18 : 6);
      const slippageFactor = BigInt(Math.floor(SLIPPAGE * 100)); // 0.5% -> 50
      amountOutMinimum = (estBig * (10000n - slippageFactor)) / 10000n;
    }

    try {
      const tx = await writeSwap({
        address: SWAP_ROUTER,
        abi: swapRouterABI,
        functionName: 'exactInputSingle',
        args: [
          {
            tokenIn: tokenInAddress,
            tokenOut: tokenOutAddress,
            fee: POOL_FEE,
            recipient: address,
            deadline: BigInt(deadline),
            amountIn: amountInWei,
            amountOutMinimum,
            sqrtPriceLimitX96: 0,
          },
        ],
      });
      setTxHash(tx);
      toast({ title: 'Swap submitted', status: 'info' });
    } catch (e: any) {
      console.error('Swap error:', e);
      toast({ title: 'Swap failed', description: e.message, status: 'error' });
    }
  };

  return (
    <VStack spacing={6} maxW="450px" mx="auto">
      <Heading size="lg" color="brand.400">Swap (Base)</Heading>

      {isWrongNetwork && (
        <Alert status="warning">
          <AlertIcon />
          Please switch your wallet to Base network.
        </Alert>
      )}

      <VStack spacing={4} bg="gray.800" p={6} borderRadius="lg" border="1px solid" borderColor="brand.700" w="100%">
        <Text alignSelf="flex-start" fontSize="sm" color="gray.400">From</Text>
        <HStack w="100%">
          <Select
            value={tokenIn}
            onChange={(e) => setTokenIn(e.target.value as 'wNKN' | 'USDC')}
            w="40%"
            bg="gray.700"
            border="none"
          >
            <option value="wNKN">wNKN</option>
            <option value="USDC">USDC</option>
          </Select>
          <Input
            placeholder="0.0"
            type="number"
            value={amountIn}
            onChange={(e) => setAmountIn(e.target.value)}
            bg="gray.700"
            border="none"
            flex={1}
          />
        </HStack>

        <Button variant="ghost" onClick={() => {
          setTokenIn(tokenOut);
          setTokenOut(tokenIn);
          setAmountIn('');
          setTxHash(undefined);
        }}>⇅</Button>

        <Text alignSelf="flex-start" fontSize="sm" color="gray.400">To (estimated)</Text>
        <HStack w="100%">
          <Select
            value={tokenOut}
            onChange={(e) => setTokenOut(e.target.value as 'wNKN' | 'USDC')}
            w="40%"
            bg="gray.700"
            border="none"
          >
            <option value="USDC">USDC</option>
            <option value="wNKN">wNKN</option>
          </Select>
          <Input
            placeholder="0.0"
            value={estimatedOut}
            isReadOnly
            bg="gray.700"
            border="none"
            flex={1}
          />
        </HStack>

        <HStack spacing={4} w="100%">
          <Button
            onClick={handleApprove}
            colorScheme="gray"
            w="50%"
            isLoading={isApproving}
            disabled={!amountIn || isWrongNetwork}
          >
            Approve
          </Button>
          <Button
            onClick={handleSwap}
            colorScheme="brand"
            w="50%"
            isLoading={isSwapping || isWaiting}
            disabled={!amountIn || isWrongNetwork}
          >
            {isWaiting ? 'Confirming...' : 'Swap'}
          </Button>
        </HStack>
      </VStack>
    </VStack>
  );
}