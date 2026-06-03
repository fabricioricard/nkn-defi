import { Link as RouterLink } from 'react-router-dom';
import {
  Box, Heading, Text, Button, VStack, SimpleGrid, Icon,
} from '@chakra-ui/react';
import { FaLink, FaExchangeAlt, FaCubes, FaChartPie } from 'react-icons/fa';

const features = [
  { title: 'Bridge', desc: 'Convert NKN Mainnet to wNKN on Base', icon: FaLink, to: '/bridge' },
  { title: 'Swap', desc: 'Trade wNKN with stablecoins', icon: FaExchangeAlt, to: '/swap' },
  { title: 'Pools', desc: 'Provide liquidity and earn fees', icon: FaCubes, to: '/pools' },
  { title: 'Portfolio', desc: 'Track your assets', icon: FaChartPie, to: '/portfolio' },
];

export default function Home() {
  return (
    <VStack spacing={12} align="center" py={10}>
      <VStack spacing={4} textAlign="center" maxW="800px">
        <Heading size="2xl" color="brand.400">
          Welcome to NKN DeFi · InfraFi
        </Heading>
        <Text fontSize="lg" color="gray.400">
          The first native liquidity layer for the NKN mainnet. Bridge, swap, and provide liquidity — all powered by infrastructure finance.
        </Text>
      </VStack>

      <SimpleGrid columns={[1, 2, 4]} spacing={8} maxW="1200px">
        {features.map((f) => (
          <Box
            key={f.to}
            bg="gray.800"
            p={6}
            borderRadius="lg"
            border="1px solid"
            borderColor="brand.700"
            textAlign="center"
            _hover={{ borderColor: 'brand.400', transform: 'translateY(-2px)' }}
            transition="all 0.3s"
          >
            <Icon as={f.icon} boxSize={10} color="brand.400" mb={4} />
            <Heading size="md" color="white" mb={2}>{f.title}</Heading>
            <Text fontSize="sm" color="gray.400" mb={4}>{f.desc}</Text>
            <Button as={RouterLink} to={f.to} colorScheme="brand" size="sm">
              Launch
            </Button>
          </Box>
        ))}
      </SimpleGrid>
    </VStack>
  );
}