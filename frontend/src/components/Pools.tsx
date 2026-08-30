import { VStack, Heading, Text, Link, Button, Card, CardBody, HStack, Icon } from '@chakra-ui/react';
import { FaExternalLinkAlt } from 'react-icons/fa';

export default function Pools() {
  return (
    <VStack spacing={8} align="stretch" maxW="800px" mx="auto">
      <Heading size="lg" color="brand.400">Liquidity Pools</Heading>

      <Card bg="gray.800" borderColor="brand.700" borderWidth="1px" p={6}>
        <CardBody>
          <VStack align="stretch" spacing={4}>
            <HStack justify="space-between" flexWrap="wrap" gap={2}>
              <Heading size="md" color="white">wNKN / USDC</Heading>
              <Link
                href="https://app.uniswap.org/explore/pools/base/0xc7A1E13a0128f96994A0020Fb105f6BB070eF94a"
                isExternal
              >
                <Button size="sm" colorScheme="brand" leftIcon={<Icon as={FaExternalLinkAlt} />}>
                  Open on Uniswap
                </Button>
              </Link>
            </HStack>
            <Text fontSize="sm" color="gray.400">
              This is the official liquidity pool for the wrapped NKN token (wNKN) on Base.
              Provide liquidity to earn fees and help establish the market price of wNKN.
            </Text>
          </VStack>
        </CardBody>
      </Card>
    </VStack>
  );
}