import { Outlet, NavLink } from 'react-router-dom';
import { Box, Flex, Button, HStack, Icon } from '@chakra-ui/react';
import { FaLink, FaExchangeAlt, FaChartPie, FaTint } from 'react-icons/fa';

const links = [
  { to: '/bridge', label: 'Bridge', icon: FaLink },
  { to: '/swap', label: 'Swap', icon: FaExchangeAlt },
  { to: '/portfolio', label: 'Portfolio', icon: FaChartPie },
  { to: '/pools', label: 'Pools', icon: FaTint },
];

export default function Layout() {
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
          <Button
            as="a"
            href="https://wallet.nkn.org/"
            target="_blank"
            size="sm"
            variant="outline"
            colorScheme="brand"
          >
            Connect Wallet
          </Button>
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