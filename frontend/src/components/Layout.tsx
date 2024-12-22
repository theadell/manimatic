import React from 'react';
import { Container, Snackbar, Alert, Box } from '@mui/material';
import { Header } from './Header';

interface LayoutProps {
  children: React.ReactNode;
  isDarkMode: boolean;
  onToggleTheme: (event: React.MouseEvent<HTMLButtonElement, MouseEvent>) => void;
  error: string | null;
  onErrorClose: () => void;
}

export const Layout = ({ 
  children, 
  isDarkMode, 
  onToggleTheme, 
  error, 
  onErrorClose 
}: LayoutProps) => {
  return (
    <Box sx={{ minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
      <Header isDarkMode={isDarkMode} onToggleTheme={onToggleTheme} />
      
      <Container
        maxWidth="xl"
        sx={{
          flex: 1,
          display: 'flex',
          flexDirection: 'column',
          gap: 4,
          py: 4,
          px: { xs: 2, sm: 3, md: 4 }
        }}
      >
        {children}
      </Container>

      <Snackbar
        open={!!error}
        autoHideDuration={3000}
        onClose={onErrorClose}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert
          onClose={onErrorClose}
          severity={error?.includes('copied') ? 'success' : 'error'}
          sx={{
            backdropFilter: 'blur(10px)',
            backgroundColor: theme =>
              theme.palette.mode === 'dark' ? 'rgba(0,0,0,0.8)' : 'rgba(255,255,255,0.8)'
          }}
        >
          {error}
        </Alert>
      </Snackbar>
    </Box>
  );
};