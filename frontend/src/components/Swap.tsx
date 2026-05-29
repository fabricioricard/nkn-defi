import { useState } from 'react';
import { VStack, Heading, Input, Button, HStack, Text, useToast, Select } from '@chakra-ui/react';

export default function Swap() {
  const [tokenIn, setTokenIn] = useState('wNKN');
  const [tokenOut, setTokenOut] = useState('USDC');
  const [amountIn, setAmountIn] = useState('');
  const [amountOut, setAmountOut] = useState('');
  const toast = useToast();

  const handleSwap = () => {
    if (!amountIn || !tokenIn || !tokenOut) {
      toast({ title: 'Preencha os campos', status: 'warning' });
      return;
    }
    // Simulação simples
    const out = (parseFloat(amountIn) * 0.997).toFixed(6);
    setAmountOut(out);
    toast({ title: `Swap estimado: ${out} ${tokenOut}`, status: 'info' });
  };

  return (
    <VStack spacing={6} maxW="450px" mx="auto">
      <Heading size="lg" color="brand.400">Swap (Base)</Heading>

      <VStack spacing={4} bg="gray.800" p={6} borderRadius="lg" border="1px solid" borderColor="brand.700" w="100%">
        <Text>From</Text>
        <HStack w="100%">
          <Select value={tokenIn} onChange={(e) => setTokenIn(e.target.value)} w="40%" bg="gray.700" border="none">
            <option value="wNKN">wNKN</option>
            <option value="USDC">USDC</option>
            <option value="USDT">USDT</option>
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
          const tmp = tokenIn;
          setTokenIn(tokenOut);
          setTokenOut(tmp);
        }}>⇅</Button>

        <Text>To (estimated)</Text>
        <HStack w="100%">
          <Select value={tokenOut} onChange={(e) => setTokenOut(e.target.value)} w="40%" bg="gray.700" border="none">
            <option value="USDC">USDC</option>
            <option value="USDT">USDT</option>
            <option value="wNKN">wNKN</option>
          </Select>
          <Input
            placeholder="0.0"
            value={amountOut}
            isReadOnly
            bg="gray.700"
            border="none"
            flex={1}
          />
        </HStack>

        <Button onClick={handleSwap} colorScheme="brand" w="100%">Swap</Button>
      </VStack>
    </VStack>
  );
}