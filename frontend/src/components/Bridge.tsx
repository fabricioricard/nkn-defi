import { useState, useEffect } from 'react';
import {
  VStack, Heading, Input, Button, Text, Code,
  Table, Thead, Tbody, Tr, Th, Td, Badge, useToast,
} from '@chakra-ui/react';
import { api } from '../api';

interface Tx {
  id: string;
  type: string;
  amount: string;
  status: string;
  timestamp: string;
}

export default function Bridge() {
  const [ethAddr, setEthAddr] = useState('');
  const [amount, setAmount] = useState('');
  const [depositAddr, setDepositAddr] = useState('');
  const [txs, setTxs] = useState<Tx[]>([]);
  const toast = useToast();

  const loadTxs = async () => {
    try {
      const data = await api.getBridgeTransactions();
      setTxs(data);
    } catch (e) {
      console.error(e);
    }
  };

  useEffect(() => { loadTxs(); }, []);

  const handleBridge = async () => {
    if (!ethAddr || !amount) {
      toast({ title: 'Preencha todos os campos', status: 'warning' });
      return;
    }
    try {
      const data = await api.createBridgeDeposit(ethAddr, amount);
      setDepositAddr(data.deposit_address);
      toast({ title: 'Endereço de depósito gerado', status: 'success' });
      loadTxs();
    } catch (e: any) {
      toast({ title: 'Erro', description: e.message, status: 'error' });
    }
  };

  return (
    <VStack spacing={6} align="stretch" maxW="600px" mx="auto">
      <Heading size="lg" color="brand.400">Bridge NKN (Mainnet → Base)</Heading>

      <VStack spacing={4} bg="gray.800" p={6} borderRadius="lg" border="1px solid" borderColor="brand.700">
        <Input
          placeholder="Ethereum address (0x...)"
          value={ethAddr}
          onChange={(e) => setEthAddr(e.target.value)}
          bg="gray.700"
          border="none"
        />
        <Input
          placeholder="Amount (NKN)"
          type="number"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          bg="gray.700"
          border="none"
        />
        <Button onClick={handleBridge} colorScheme="brand" w="100%">
          Start Bridge
        </Button>
        {depositAddr && (
          <Text color="gray.300">
            Send NKN to address: <Code colorScheme="purple">{depositAddr}</Code>
          </Text>
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
          </Tr>
        </Thead>
        <Tbody>
          {txs.map((tx) => (
            <Tr key={tx.id}>
              <Td>{tx.type}</Td>
              <Td>{tx.amount} NKN</Td>
              <Td>
                <Badge colorScheme={tx.status === 'completed' ? 'green' : 'yellow'}>
                  {tx.status}
                </Badge>
              </Td>
              <Td>{new Date(tx.timestamp).toLocaleTimeString()}</Td>
            </Tr>
          ))}
        </Tbody>
      </Table>
    </VStack>
  );
}