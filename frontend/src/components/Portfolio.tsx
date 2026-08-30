import { VStack, Heading, Text, HStack, Icon, Badge, Divider, Spinner } from '@chakra-ui/react';
import { FaWallet } from 'react-icons/fa';
import { useAccount, useReadContracts } from 'wagmi';
import { formatUnits } from 'viem';

// ABI mínima para ERC-20
const erc20ABI = [
  {
    constant: true,
    inputs: [{ name: 'account', type: 'address' }],
    name: 'balanceOf',
    outputs: [{ name: '', type: 'uint256' }],
    type: 'function',
  },
  {
    constant: true,
    inputs: [],
    name: 'decimals',
    outputs: [{ name: '', type: 'uint8' }],
    type: 'function',
  },
  {
    constant: true,
    inputs: [],
    name: 'symbol',
    outputs: [{ name: '', type: 'string' }],
    type: 'function',
  },
];

const wNKNAddress = '0x1B24ED102b530887B4388b61FB612121f6eD635E';
const USDCAddress = '0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913';

export default function Portfolio() {
  const { address, isConnected } = useAccount();

  const { data, isLoading, isError } = useReadContracts({
    contracts: [
      { address: wNKNAddress, abi: erc20ABI, functionName: 'balanceOf', args: [address || '0x0'] },
      { address: wNKNAddress, abi: erc20ABI, functionName: 'decimals' },
      { address: wNKNAddress, abi: erc20ABI, functionName: 'symbol' },
      { address: USDCAddress, abi: erc20ABI, functionName: 'balanceOf', args: [address || '0x0'] },
      { address: USDCAddress, abi: erc20ABI, functionName: 'decimals' },
      { address: USDCAddress, abi: erc20ABI, functionName: 'symbol' },
    ],
  });

  if (!isConnected) {
    return (
      <VStack spacing={6} maxW="500px" mx="auto" textAlign="center" py={10}>
        <Icon as={FaWallet} boxSize={16} color="gray.500" />
        <Heading size="lg" color="gray.400">Connect your wallet</Heading>
        <Text color="gray.500">Connect to view your balances of wNKN and USDC on Base.</Text>
      </VStack>
    );
  }

  if (isLoading) {
    return (
      <VStack py={10}>
        <Spinner size="xl" color="brand.400" />
        <Text color="gray.400">Loading balances...</Text>
      </VStack>
    );
  }

  if (isError || !data) {
    return (
      <VStack py={10}>
        <Text color="red.400">Failed to load balances. Please try again.</Text>
      </VStack>
    );
  }

  const wNKNBalance = data[0].result as bigint;
  const wNKNDecimals = data[1].result as number;
  const wNKNSymbol = data[2].result as string;
  const USDCBalance = data[3].result as bigint;
  const USDCDecimals = data[4].result as number;
  const USDCSymbol = data[5].result as string;

  const formattedWNKN = Number(formatUnits(wNKNBalance, wNKNDecimals)).toFixed(4);
  const formattedUSDC = Number(formatUnits(USDCBalance, USDCDecimals)).toFixed(2);

  return (
    <VStack spacing={6} maxW="600px" mx="auto" py={8}>
      <Heading size="lg" color="brand.400">Portfolio</Heading>
      <Text color="gray.400">
        Connected as <strong>{address?.slice(0, 6)}...{address?.slice(-4)}</strong>
      </Text>

      <VStack spacing={4} w="100%" bg="gray.800" p={6} borderRadius="lg" border="1px solid" borderColor="brand.700">
        <HStack justify="space-between" w="100%">
          <Text color="gray.300">Token</Text>
          <Text color="gray.300">Balance</Text>
        </HStack>
        <Divider borderColor="gray.600" />
        <HStack justify="space-between" w="100%">
          <Text color="white" fontWeight="semibold">{wNKNSymbol}</Text>
          <Badge colorScheme="purple">{formattedWNKN}</Badge>
        </HStack>
        <HStack justify="space-between" w="100%">
          <Text color="white" fontWeight="semibold">{USDCSymbol}</Text>
          <Badge colorScheme="blue">{formattedUSDC}</Badge>
        </HStack>
      </VStack>
    </VStack>
  );
}