import { createTheme, Theme } from '@mui/material';

export const createAppTheme = (mode: 'light' | 'dark'): Theme => createTheme({
  palette: {
    mode,
    ...(mode === 'dark' 
      ? {
          background: {
            default: '#121212',
            paper: '#1e1e1e'
          },
          primary: {
            main: '#4a90e2',
            light: '#6ab0ff',
            dark: '#3a80d2'
          },
          text: {
            primary: '#e0e0e0',
            secondary: '#b0b0b0'
          }
        }
      : {
          background: {
            default: '#f5f5f5',
            paper: '#ffffff'
          },
          primary: {
            main: '#3a80d2',
            light: '#4a90e2',
            dark: '#2c6cb3'
          },
          text: {
            primary: '#1a1a1a',
            secondary: '#666666'
          }
        }
    )
  },
  typography: {
    fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
    h1: {
      fontSize: '2.5rem',
      fontWeight: 600,
      letterSpacing: '-0.02em'
    }
  },
  components: {
    MuiPaper: {
      styleOverrides: {
        root: {
          borderRadius: 8,
          boxShadow: '0 4px 6px rgba(0,0,0,0.1)',
          transition: 'box-shadow 0.3s ease'
        }
      }
    },
    MuiTextField: {
      styleOverrides: {
        root: {
          '& .MuiOutlinedInput-root': {
            borderRadius: 6,
            backgroundColor: mode === 'dark' ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.02)',
            transition: 'background-color 0.2s ease'
          }
        }
      }
    }
  }
});
