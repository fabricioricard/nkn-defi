import { useState, useEffect } from 'react';
import {
  VStack, Heading, Input, Button, HStack, Text,
  SimpleGrid, Card, CardHeader, CardBody, Badge,
  useToast, Select,
} from '@chakra-ui/react';
import { api } from '../api';

interface Pool {
  id: string;
  token0: string;
  token1: string;
  reserve0: string;
  reserve1: string;
  total_lp_tokens: string;
  fee_bps: number;
}

export default function Pools() {
  const [pools, setPools] = useState<Pool[]>([]);
  const [token0, setToken0] = useState('');
  const [token1, setToken1] = useState('');
  const [feeBps, setFeeBps] = useState(30);
  const toast = useToast();

  const loadPools = async () => {
    try {
      const data = await api.getPools();
      setPools(data);
    } catch (e) {
      console.error(e);
    }
  };

  useEffect(() => { loadPools(); }, []);

  const handleCreatePool = async () => {
    if (!token0 || !token1) {
      toast({ title: 'Preencha os tokens', status: 'warning' });
      return;
    }
    try {
      await api.createPool(token0, token1, feeBps);
      toast({ title: 'Pool criada!', status: 'success' });
      loadPools();
      setToken0(''); setToken1('');
    } catch (e: any) {
      toast({ title: 'Erro', description: e.message, status: 'error' });
    }
  };

  const handleAddLiquidity = async (poolId: string, amount0: string, amount1: string) => {
    if (!amount0 || !amount1) return toast({ title: 'Valores obrigatórios', status: 'warning' });
    try {
      await api.addLiquidity(poolId, amount0, amount1);
      toast({ title: 'Liquidez adicionada', status: 'success' });
      loadPools();
    } catch (e: any) {
      toast({ title: 'Erro', description: e.message, status: 'error' });
    }
  };

  const handleSwap = async (poolId: string, tokenIn: string, amountIn: string) => {
    if (!amountIn) return toast({ title: 'Quantidade obrigatória', status: 'warning' });
    try {
      const res = await api.swap(poolId, tokenIn, amountIn);
      toast({ title: 'Swap executado', description: `Recebido: ${res.amount_out} (fee: ${res.fee})`, status: 'success' });
      loadPools();
    } catch (e: any) {
      toast({ title: 'Erro', description: e.message, status: 'error' });
    }
  };

  return (
    <VStack spacing={8} align="stretch">
      <Heading size="lg" color="brand.400">Create Pool</Heading>
      <HStack bg="gray.800" p={4} borderRadius="lg" border="1px solid" borderColor="brand.700">
        <Input placeholder="Token0" value={token0} onChange={(e) => setToken0(e.target.value)} bg="gray.700" border="none" />
        <Input placeholder="Token1" value={token1} onChange={(e) => setToken1(e.target.value)} bg="gray.700" border="none" />
        <Input type="number" value={feeBps} onChange={(e) => setFeeBps(Number(e.target.value))} bg="gray.700" border="none" w="120px" />
        <Button onClick={handleCreatePool} colorScheme="brand">Create</Button>
      </HStack>

      <Heading size="lg" color="brand.400">Liquidity Pools</Heading>
      {pools.length === 0 && <Text color="gray.500">Nenhuma pool criada ainda.</Text>}
      <SimpleGrid columns={[1, 2, 3]} spacing={6}>
        {pools.map((pool) => (
          <Card key={pool.id} bg="gray.800" borderColor="brand.700" borderWidth="1px">
            <CardHeader>
              <HStack justify="space-between">
                <Heading size="sm">{pool.token0}/{pool.token1}</Heading>
                <Badge colorScheme="yellow">{pool.total_lp_tokens} LP</Badge>
              </HStack>
            </CardHeader>
            <CardBody>
              <VStack align="stretch" spacing={2}>
                <HStack justify="space-between">
                  <Text>{pool.token0} Reserve:</Text>
                  <Text color="yellow.300" fontFamily="mono">{pool.reserve0}</Text>
                </HStack>
                <HStack justify="space-between">
                  <Text>{pool.token1} Reserve:</Text>
                  <Text color="yellow.300" fontFamily="mono">{pool.reserve1}</Text>
                </HStack>
                <Text fontSize="sm">Fee: {pool.fee_bps} bps</Text>

                <Text fontWeight="semibold" mt={2}>Add Liquidity</Text>
                <HStack>
                  <Input placeholder={pool.token0} id={`amount0-${pool.id}`} size="sm" bg="gray.700" />
                  <Input placeholder={pool.token1} id={`amount1-${pool.id}`} size="sm" bg="gray.700" />
                  <Button size="sm" colorScheme="brand" onClick={() => {
                    const a0 = (document.getElementById(`amount0-${pool.id}`) as HTMLInputElement)?.value;
                    const a1 = (document.getElementById(`amount1-${pool.id}`) as HTMLInputElement)?.value;
                    handleAddLiquidity(pool.id, a0 || '', a1 || '');
                  }}>+</Button>
                </HStack>

                <Text fontWeight="semibold" mt={3}>Swap</Text>
                <HStack>
                  <Input placeholder="Amount" id={`swapIn-${pool.id}`} size="sm" bg="gray.700" />
                  <Select id={`tokenIn-${pool.id}`} size="sm" bg="gray.700" w="100px">
                    <option>{pool.token0}</option>
                    <option>{pool.token1}</option>
                  </Select>
                  <Button size="sm" colorScheme="brand" variant="outline" onClick={() => {
                    const amountIn = (document.getElementById(`swapIn-${pool.id}`) as HTMLInputElement)?.value;
                    const tokenIn = (document.getElementById(`tokenIn-${pool.id}`) as HTMLSelectElement)?.value;
                    handleSwap(pool.id, tokenIn || pool.token0, amountIn || '');
                  }}>Swap</Button>
                </HStack>
              </VStack>
            </CardBody>
          </Card>
        ))}
      </SimpleGrid>
    </VStack>
  );
}