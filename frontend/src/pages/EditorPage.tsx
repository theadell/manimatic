import React, { useState } from 'react';
import { Box } from '@mui/material';
import { Layout } from '../components/Layout';
import { createAppTheme } from '../theme/theme';
import { ThemeProvider, CssBaseline } from '@mui/material';


function EditorPage() {
  const [isDarkMode, setIsDarkMode] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const theme = React.useMemo(() => createAppTheme(isDarkMode ? 'dark' : 'light'), [isDarkMode]);


  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Layout
        isDarkMode={isDarkMode}
        onToggleTheme={() => setIsDarkMode(!isDarkMode)}
        error={error}
        onErrorClose={() => setError(null)}
      >
        <Box
          sx={{
            display: 'grid',
            gridTemplateColumns: '1fr 1fr',
            gap: 4,
            height: '100%',
          }}
        >
        </Box>
      </Layout>
    </ThemeProvider>
  );
}

export default EditorPage;
