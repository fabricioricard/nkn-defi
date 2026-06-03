import {
  VStack, Heading, Text, Button, Icon,
} from '@chakra-ui/react';
import { FaWallet } from 'react-icons/fa';
import { useAccount, useConnect, useDisconnect } from 'wagmi';

export default function Portfolio() {
  const { address, isConnected } = useAccount();
  const { connect, connectors } = useConnect();
  const { disconnect } = useDisconnect();

  return (
    <VStack spacing={6} maxW="500px" mx="auto" textAlign="center" py={10}>
      <Icon as={FaWallet} boxSize={16} color="gray.500" />
      <Heading size="lg" color="gray.400">Portfolio</Heading>

      {isConnected ? (
        <>
          <Text color="gray.300">
            Connected as <strong>{address?.slice(0, 6)}...{address?.slice(-4)}</strong>
          </Text>
          <Text fontSize="sm" color="gray.500">
            Portfolio details will be available soon. Stay tuned!
          </Text>
          <Button size="sm" variant="outline" colorScheme="brand" onClick={() => disconnect()}>
            Disconnect
          </Button>
        </>
      ) : (
        <>
          <Text color="gray.500">Connect your wallet to view balances</Text>
          <Text fontSize="sm" color="gray.600">Supports wNKN (Base), USDC, USDT</Text>
          <Button
            colorScheme="brand"
            leftIcon={<Icon as={FaWallet} />}
            onClick={() => connect({ connector: connectors[0] })}
          >
            Connect Wallet
          </Button>
        </>
      )}
    </VStack>
  );
}