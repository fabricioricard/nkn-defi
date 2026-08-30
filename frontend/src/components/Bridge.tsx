import { useState, useEffect } from 'react';
import {
  VStack, Heading, Input, Button, Text, Code,
  Table, Thead, Tbody, Tr, Th, Td, Badge, useToast,
  HStack, Icon,
} from '@chakra-ui/react';
import { FaWallet, FaCheckCircle, FaTimes } from 'react-icons/fa';
import { useAccount } from 'wagmi';
import { api } from '../api';

interface Tx {
  id: string;
  type: string;
  amount: string;
  status: string;
  timestamp: string;
}

export default function Bridge() {
  const { address, isConnected } = useAccount();
  const [amount, setAmount] = useState('');
  const [depositAddr, setDepositAddr] = useState('');
  const [memo, setMemo] = useState('');
  const [txs, setTxs] = useState<Tx[]>([]);
  const toast = useToast();

  const loadTxs = async () => {
    if (!address) {
      setTxs([]);
      return;
    }
    try {
      const data = await api.getBridgeTransactions(address);
      setTxs(Array.isArray(data) ? data : []);
    } catch (e) {
      console.error(e);
      setTxs([]);
    }
  };

  useEffect(() => {
    loadTxs();
  }, [address]);

  const handleCancel = async (id: string) => {
    try {
      await api.cancelDeposit(id);
      toast({ title: 'Deposit cancelled', status: 'info' });
      loadTxs();
    } catch (e: any) {
      toast({ title: 'Error', description: e.message, status: 'error' });
    }
  };

  const handleBridge = async () => {
    if (!address) {
      toast({ title: 'Connect your wallet first', status: 'warning' });
      return;
    }
    if (!amount || parseFloat(amount) <= 0) {
      toast({ title: 'Enter a valid amount', status: 'warning' });
      return;
    }
    try {
      // Não passa mais tx_hash
      const data = await api.createBridgeDeposit(address, amount);
      setDepositAddr(data.deposit_address);
      setMemo(data.memo);
      toast({ title: 'Deposit address generated', status: 'success' });
      loadTxs();
    } catch (e: any) {
      toast({ title: 'Error', description: e.message, status: 'error' });
    }
  };

  return (
    <VStack spacing={6} align="stretch" maxW="600px" mx="auto">
      <Heading size="lg" color="brand.400">Bridge NKN (Mainnet → Base)</Heading>

      <VStack spacing={4} bg="gray.800" p={6} borderRadius="lg" border="1px solid" borderColor="brand.700">
        <Text fontSize="sm" color="gray.400" alignSelf="flex-start">Destination Address (Base)</Text>
        <HStack w="100%" bg="gray.700" borderRadius="md" px={3} py={2}>
          <Icon as={isConnected ? FaCheckCircle : FaWallet} color={isConnected ? 'green.400' : 'gray.500'} />
          <Text flex={1} color={isConnected ? 'white' : 'gray.500'} fontSize="sm" isTruncated>
            {isConnected ? address : 'Wallet not connected'}
          </Text>
          {!isConnected && (
            <Button size="xs" colorScheme="brand" onClick={() => document.getElementById('connect-wallet-btn')?.click()}>
              Connect
            </Button>
          )}
        </HStack>

        <Text fontSize="sm" color="gray.400" alignSelf="flex-start" mt={2}>Amount (NKN)</Text>
        <Input
          placeholder="0.0"
          type="number"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          bg="gray.700"
          border="none"
          isDisabled={!isConnected}
        />

        <Button
          onClick={handleBridge}
          colorScheme="brand"
          w="100%"
          isDisabled={!isConnected}
          leftIcon={<Icon as={FaWallet} />}
        >
          Start Bridge
        </Button>
        {depositAddr && (
          <>
            <Text color="gray.300" fontSize="sm">
              Send NKN to address: <Code colorScheme="purple">{depositAddr}</Code>
            </Text>
            <Text color="gray.300" fontSize="sm">
              Memo: <Code colorScheme="orange">{memo}</Code>
            </Text>
          </>
        )}
      </VStack>

      <Heading size="md" color="gray.400">Recent Transactions</Heading>
      <Table variant="simple" size="sm">
        <Thead>
          <Tr>
            <Th>Type</Th>
            <Th>Amount</Th>
            <Th>Status</Th>
            <Th>Time</Th>
            <Th></Th>
          </Tr>
        </Thead>
        <Tbody>
          {(Array.isArray(txs) ? txs : []).map((tx) => (
            <Tr key={tx.id}>
              <Td>{tx.type}</Td>
              <Td>{tx.amount} NKN</Td>
              <Td>
                <Badge colorScheme={tx.status === 'completed' ? 'green' : 'yellow'}>
                  {tx.status}
                </Badge>
              </Td>
              <Td>{new Date(tx.timestamp).toLocaleTimeString()}</Td>
              <Td>
                {tx.status === 'pending' && tx.type === 'deposit' && (
                  <Button
                    size="xs"
                    colorScheme="red"
                    variant="ghost"
                    leftIcon={<FaTimes />}
                    onClick={() => handleCancel(tx.id)}
                  >
                    Cancel
                  </Button>
                )}
              </Td>
            </Tr>
          ))}
        </Tbody>
      </Table>
    </VStack>
  );
}