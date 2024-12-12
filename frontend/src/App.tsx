import React, { useState, useCallback, useEffect } from 'react';
import { 
  Box, 
  Container, 
  Typography, 
  TextField, 
  IconButton, 
  Grid, 
  Snackbar, 
  Alert,
  createTheme, 
  ThemeProvider, 
  CssBaseline,
  Paper,
  CircularProgress
} from '@mui/material';
import SendIcon from '@mui/icons-material/Send';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import { motion, AnimatePresence } from 'framer-motion';

interface MessageType {
  type: 'script' | 'video';
  sessionId: string;
  status: 'success' | 'error';
  content: string;
  details?: Record<string, any>;
}

const theme = createTheme({
  palette: {
    mode: 'dark',
    background: {
      default: '#121212',
      paper: '#1e1e1e'
    },
    primary: {
      main: '#3fdaae'
    }
  },
  typography: {
    fontFamily: 'Inter, sans-serif',
    h1: {
      fontSize: '2.5rem',
      fontWeight: 700
    }
  },
  components: {
    MuiTextField: {
      styleOverrides: {
        root: {
          '& .MuiOutlinedInput-root': {
            borderRadius: 16,
            backgroundColor: 'rgba(255,255,255,0.05)',
          }
        }
      }
    }
  }
});

function AnimationGenerator() {
  const [prompt, setPrompt] = useState('');
  const [script, setScript] = useState('');
  const [videoUrl, setVideoUrl] = useState('');
  const [isGenerating, setIsGenerating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleGenerate = useCallback(async (e?: React.FormEvent) => {
    if (e) e.preventDefault();

    if (!prompt.trim()) {
      setError('Please enter a prompt');
      return;
    }

    setIsGenerating(true);
    setScript('');
    setVideoUrl('');
    setError(null);

    try {
      const response = await fetch('/api/generate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ prompt })
      });

      if (!response.ok) {
        throw new Error('Generation failed');
      }
    } catch (error) {
      console.error('Error generating content:', error);
      setError('Failed to generate animation. Please try again.');
      setIsGenerating(false);
    }
  }, [prompt]);

  const handleCopyScript = useCallback(() => {
    if (script) {
      navigator.clipboard.writeText(script)
        .then(() => setError('Script copied to clipboard'))
        .catch(() => setError('Failed to copy script'));
    }
  }, [script]);

  useEffect(() => {
    const eventSource = new EventSource('/api/events');

    eventSource.onmessage = (event) => {
      const message: MessageType = JSON.parse(event.data);
      switch (message.type) {
        case 'script':
          setScript(message.content);
          break;
        case 'video':
          setVideoUrl(message.content);
          setIsGenerating(false);
          break;
      }
    };

    eventSource.onerror = (error) => {
      console.error('EventSource failed:', error);
      eventSource.close();
      setIsGenerating(false);
      setError('Connection error. Please refresh and try again.');
    };

    return () => {
      eventSource.close();
    };
  }, []);

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Container 
        maxWidth="lg" 
        sx={{ 
          display: 'flex', 
          flexDirection: 'column', 
          height: '100vh', 
          py: 4 
        }}
      >
        <Typography 
          variant="h1" 
          align="center" 
          color="primary"
          sx={{ mb: 4 }}
        >
          Manimatic
        </Typography>

        {/* Content Area */}
        <Box 
          sx={{ 
            flexGrow: 1, 
            display: 'flex', 
            flexDirection: 'column', 
            overflow: 'hidden' 
          }}
        >
          {/* Results Grid */}
          <Grid 
            container 
            spacing={3} 
            sx={{ 
              flexGrow: 1, 
              mb: 2, 
              overflowY: 'auto' 
            }}
          >
            {/* Video Column */}
            <Grid item xs={12} md={6}>
              <AnimatePresence>
                {videoUrl && (
                  <motion.div
                    initial={{ opacity: 0, scale: 0.9 }}
                    animate={{ opacity: 1, scale: 1 }}
                    exit={{ opacity: 0, scale: 0.9 }}
                  >
                    <Paper 
                      elevation={3} 
                      sx={{ 
                        p: 2, 
                        borderRadius: 2, 
                        height: '100%',
                        display: 'flex',
                        flexDirection: 'column'
                      }}
                    >
                      <Typography 
                        variant="h6" 
                        color="primary" 
                        sx={{ mb: 2 }}
                      >
                        Generated Animation
                      </Typography>
                      <video 
                        src={videoUrl} 
                        controls 
                        style={{ 
                          width: '100%', 
                          borderRadius: 16, 
                          boxShadow: '0 10px 25px rgba(0,0,0,0.2)' 
                        }} 
                      />
                    </Paper>
                  </motion.div>
                )}
              </AnimatePresence>
            </Grid>

            {/* Script Column */}
            <Grid item xs={12} md={6}>
              <AnimatePresence>
                {script && (
                  <motion.div
                    initial={{ opacity: 0, scale: 0.9 }}
                    animate={{ opacity: 1, scale: 1 }}
                    exit={{ opacity: 0, scale: 0.9 }}
                  >
                    <Paper 
                      elevation={3} 
                      sx={{ 
                        p: 2, 
                        borderRadius: 2, 
                        height: '100%',
                        position: 'relative',
                        display: 'flex',
                        flexDirection: 'column'
                      }}
                    >
                      <Typography 
                        variant="h6" 
                        color="primary" 
                        sx={{ mb: 2 }}
                      >
                        Animation Script
                      </Typography>
                      <IconButton
                        color="primary"
                        onClick={handleCopyScript}
                        sx={{ 
                          position: 'absolute', 
                          top: 8, 
                          right: 8 
                        }}
                      >
                        <ContentCopyIcon />
                      </IconButton>
                      <Box 
                        sx={{ 
                          flexGrow: 1, 
                          overflowY: 'auto',
                          backgroundColor: 'rgba(255,255,255,0.05)',
                          borderRadius: 2,
                          p: 2
                        }}
                      >
                        <Typography 
                          component="pre" 
                          variant="body2"
                          sx={{ 
                            fontFamily: 'monospace', 
                            whiteSpace: 'pre-wrap', 
                            margin: 0
                          }}
                        >
                          {script}
                        </Typography>
                      </Box>
                    </Paper>
                  </motion.div>
                )}
              </AnimatePresence>
            </Grid>
          </Grid>

          {/* Loading Indicator */}
          {isGenerating && (
            <Box display="flex" justifyContent="center" width="100%" sx={{ mb: 2 }}>
              <CircularProgress color="primary" />
            </Box>
          )}
        </Box>

        {/* Prompt Input */}
        <Box 
          component="form" 
          onSubmit={handleGenerate}
          sx={{ 
            mt: 2, 
            width: '100%' 
          }}
        >
          <TextField
            fullWidth
            variant="outlined"
            placeholder="Describe your animation..."
            value={prompt}
            onChange={(e) => setPrompt(e.target.value)}
            disabled={isGenerating}
            InputProps={{
              endAdornment: (
                <IconButton 
                  color="primary" 
                  type="submit"
                  disabled={isGenerating}
                >
                  {isGenerating ? <CircularProgress size={24} /> : <SendIcon />}
                </IconButton>
              )
            }}
          />
        </Box>

        <Snackbar
          open={!!error}
          autoHideDuration={6000}
          onClose={() => setError(null)}
          anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
        >
          <Alert 
            onClose={() => setError(null)} 
            severity={error?.includes('copied') ? 'success' : 'error'}
            sx={{ width: '100%' }}
          >
            {error}
          </Alert>
        </Snackbar>
      </Container>
    </ThemeProvider>
  );
}

export default AnimationGenerator;