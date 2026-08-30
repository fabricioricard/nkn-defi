import { useState } from 'react';
import {
  VStack, Heading, Input, Button, HStack, Text,
  Select, useToast,
} from '@chakra-ui/react';
import {
  useAccount, useWriteContract, useWaitForTransactionReceipt,
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

// ABI mínimo do SwapRouter02
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

const WNKN_ADDRESS = '0x1B24ED102b530887B4388b61FB612121f6eD635E';
const USDC_ADDRESS = '0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913';
const SWAP_ROUTER = '0x2626664c2603336E57B271c5C0b26F421741e481';
const POOL_FEE = 500; // 0,05%

export default function Swap() {
  const { address } = useAccount();
  const toast = useToast();
  const { writeContractAsync: writeApprove, isPending: isApproving } = useWriteContract();
  const { writeContractAsync: writeSwap, isPending: isSwapping } = useWriteContract();

  const [tokenIn, setTokenIn] = useState<'wNKN' | 'USDC'>('wNKN');
  const [tokenOut, setTokenOut] = useState<'wNKN' | 'USDC'>('USDC');
  const [amountIn, setAmountIn] = useState('');
  const [txHash, setTxHash] = useState<`0x${string}` | undefined>(undefined);

  const { isLoading: isWaiting, isSuccess: swapSuccess } = useWaitForTransactionReceipt({
    hash: txHash,
  });

  const handleApprove = async () => {
    if (!address || !amountIn) return;
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
    if (!address || !amountIn) return;
    const tokenInAddress = tokenIn === 'wNKN' ? WNKN_ADDRESS : USDC_ADDRESS;
    const tokenOutAddress = tokenOut === 'wNKN' ? WNKN_ADDRESS : USDC_ADDRESS;
    const amountInWei = parseUnits(amountIn, tokenIn === 'wNKN' ? 18 : 6);
    const deadline = Math.floor(Date.now() / 1000) + 60 * 20;

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
            amountOutMinimum: 0,
            sqrtPriceLimitX96: 0,
          },
        ],
      });
      setTxHash(tx);
      toast({ title: 'Swap submitted', status: 'info' });
    } catch (e: any) {
      toast({ title: 'Swap failed', description: e.message, status: 'error' });
    }
  };

  return (
    <VStack spacing={6} maxW="450px" mx="auto">
      <Heading size="lg" color="brand.400">Swap (Base)</Heading>

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
            value={swapSuccess ? 'Success' : ''}
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
          >
            Approve
          </Button>
          <Button
            onClick={handleSwap}
            colorScheme="brand"
            w="50%"
            isLoading={isSwapping || isWaiting}
            disabled={!amountIn}
          >
            {isWaiting ? 'Confirming...' : 'Swap'}
          </Button>
        </HStack>
      </VStack>
    </VStack>
  );
}