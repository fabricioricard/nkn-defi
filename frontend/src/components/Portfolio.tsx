import { VStack, Heading, Text, Button, Icon } from '@chakra-ui/react';
import { FaWallet } from 'react-icons/fa';

export default function Portfolio() {
  return (
    <VStack spacing={6} maxW="500px" mx="auto" textAlign="center" py={10}>
      <Icon as={FaWallet} boxSize={16} color="gray.500" />
      <Heading size="lg" color="gray.400">Portfolio</Heading>
      <Text color="gray.500">Connect your wallet to view balances</Text>
      <Text fontSize="sm" color="gray.600">Supports wNKN (Base), USDC, USDT</Text>
      <Button colorScheme="brand" leftIcon={<FaWallet />}>
        Connect Wallet
      </Button>
    </VStack>
  );
}