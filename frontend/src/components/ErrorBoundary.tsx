import { Component, ReactNode } from 'react';
import { Box, Heading, Button, Text } from '@chakra-ui/react';

interface Props { children: ReactNode; }
interface State { hasError: boolean; error: Error | null; }

class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error };
  }

  render() {
    if (this.state.hasError) {
      return (
        <Box p={10} textAlign="center">
          <Heading color="red.400" mb={4}>Something went wrong</Heading>
          <Text color="gray.400" mb={4}>{this.state.error?.message || 'Unknown error'}</Text>
          <Button onClick={() => window.location.reload()} colorScheme="brand">
            Reload Page
          </Button>
        </Box>
      );
    }
    return this.props.children;
  }
}

export default ErrorBoundary;