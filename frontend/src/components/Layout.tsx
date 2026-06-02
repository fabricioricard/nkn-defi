import { Outlet, NavLink } from 'react-router-dom';
import {
  Box, Flex, Button, HStack, Icon, Text,
} from '@chakra-ui/react';
import { FaLink, FaExchangeAlt, FaChartPie, FaTint, FaWallet } from 'react-icons/fa';
import { useAccount, useConnect, useDisconnect } from 'wagmi';

const links = [
  { to: '/bridge', label: 'Bridge', icon: FaLink },
  { to: '/swap', label: 'Swap', icon: FaExchangeAlt },
  { to: '/pools', label: 'Pools', icon: FaTint },
  { to: '/portfolio', label: 'Portfolio', icon: FaChartPie },
];

export default function Layout() {
  const { address, isConnected } = useAccount();
  const { connect, connectors } = useConnect();
  const { disconnect } = useDisconnect();

  return (
    <Box minH="100vh" bg="gray.900">
      <Flex
        as="nav"
        align="center"
        justify="space-between"
        p={4}
        bg="gray.800"
        borderBottom="1px solid"
        borderColor="brand.700"
      >
        <HStack spacing={4}>
          <Icon as={FaTint} boxSize={6} color="brand.400" />
          <NavLink to="/" style={{ fontWeight: 'bold', fontSize: '1.2rem', color: '#a78bfa' }}>
            NKN DeFi · InfraFi
          </NavLink>
        </HStack>
        <HStack spacing={2}>
          {links.map((link) => (
            <NavLink
              key={link.to}
              to={link.to}
              style={({ isActive }) => ({
                padding: '0.5rem 1rem',
                borderRadius: '8px',
                color: isActive ? '#7c3aed' : '#94a3b8',
                backgroundColor: isActive ? 'rgba(124, 58, 237, 0.2)' : 'transparent',
                fontWeight: 500,
              })}
            >
              <HStack spacing={1}>
                <Icon as={link.icon} />
                <span>{link.label}</span>
              </HStack>
            </NavLink>
          ))}
          {isConnected ? (
            <HStack>
              <Text fontSize="sm" color="gray.300">
                {address?.slice(0, 6)}...{address?.slice(-4)}
              </Text>
              <Button size="sm" onClick={() => disconnect()} variant="outline" colorScheme="brand">
                Disconnect
              </Button>
            </HStack>
          ) : (
            <Button
              id="connect-wallet-btn"
              size="sm"
              colorScheme="brand"
              leftIcon={<Icon as={FaWallet} />}
              onClick={() => connect({ connector: connectors[0] })}
            >
              Connect Wallet
            </Button>
          )}
        </HStack>
      </Flex>

      <Box p={8}>
        <Outlet />
      </Box>

      <Box as="footer" textAlign="center" py={4} color="gray.500" borderTop="1px solid" borderColor="brand.700">
        © 2026 NKN DeFi · Powered by InfraFi
      </Box>
    </Box>
  );
}