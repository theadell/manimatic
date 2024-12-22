import { Box, AppBar, Toolbar, Typography, IconButton, Container } from '@mui/material';
import { Moon, Sun } from 'lucide-react';
import { Link, useLocation } from 'react-router-dom';

interface HeaderProps {
  isDarkMode: boolean;
  onToggleTheme: (event: React.MouseEvent<HTMLButtonElement, MouseEvent>) => void;
}

export const Header = ({ isDarkMode, onToggleTheme }: HeaderProps) => {
  useLocation();
  
  return (
    <AppBar 
      position="sticky" 
      elevation={0}
      sx={{
        backdropFilter: 'blur(10px)',
        backgroundColor: 'transparent',
        borderBottom: '1px solid',
        borderColor: theme => theme.palette.divider
      }}
    >
      <Container maxWidth="xl">
        <Toolbar disableGutters sx={{ justifyContent: 'space-between' }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
            <Typography 
              variant="h6" 
              component={Link}
              to="/"
              sx={{ 
                fontWeight: 600,
                textDecoration: 'none',
                background: theme => 
                  theme.palette.mode === 'dark' 
                    ? 'linear-gradient(45deg, #4a90e2 30%, #6ab0ff 90%)'
                    : 'linear-gradient(45deg, #3a80d2 30%, #4a90e2 90%)',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent'
              }}
            >
              Manimatic
            </Typography>

{/* 
            <Box sx={{ ml: 4, display: 'flex', gap: 1 }}>
              <Button
                component={Link}
                to="/"
                startIcon={<Wand2 size={16} />}
                color={location.pathname === '/' ? 'primary' : 'inherit'}
                sx={{ 
                  opacity: location.pathname === '/' ? 1 : 0.7,
                  '&:hover': { opacity: 1 }
                }}
              >
                Generator
              </Button>
              <Button
                component={Link}
                to="/editor"
                startIcon={<Code size={16} />}
                color={location.pathname === '/editor' ? 'primary' : 'inherit'}
                sx={{ 
                  opacity: location.pathname === '/editor' ? 1 : 0.7,
                  '&:hover': { opacity: 1 }
                }}
              >
                Editor
              </Button>
            </Box>
                  */}

          </Box>
          <IconButton onClick={onToggleTheme} color="inherit">
            {isDarkMode ? <Sun size={20} /> : <Moon size={20} />}
          </IconButton>
        </Toolbar>
      </Container>
    </AppBar>
  );
};