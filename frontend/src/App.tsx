import { useState, useEffect, FormEvent } from 'react';
import {
  Box,
  Container,
  Typography,
  TextField,
  Button,
  CircularProgress,
  Card,
  CardHeader,
  CardContent,
  Paper,
  createTheme,
  ThemeProvider,
  CssBaseline,
  Fade,
  Grow,
} from '@mui/material';
import Grid2 from '@mui/material/Grid2'; // New MUI Grid

interface MessageType {
  type: 'script' | 'video';
  sessionId: string;
  status: 'success' | 'error';
  content: any;
  details?: Record<string, any>;
}

const darkTheme = createTheme({
  palette: {
    mode: 'dark',
    background: {
      default: '#121212',
      paper: '#1e1e1e',
    },
    primary: {
      main: '#3fdaae',
    },
  },
  typography: {
    fontFamily: 'Inter, sans-serif',
  }
});

function App() {
  const [prompt, setPrompt] = useState('');
  const [script, setScript] = useState('');
  const [videoUrl, setVideoUrl] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!prompt.trim()) return;
    setIsLoading(true);
    setScript('');
    setVideoUrl('');

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
      setIsLoading(false);
    }
  }

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
          setIsLoading(false);
          break;
      }
    };

    eventSource.onerror = (error) => {
      console.error('EventSource failed:', error);
      eventSource.close();
      setIsLoading(false);
    };

    return () => {
      eventSource.close();
    };
  }, []);

  return (
    <ThemeProvider theme={darkTheme}>
      <CssBaseline />
      <Box 
        sx={{ 
          width: '100vw', 
          height: '100vh', 
          display: 'flex', 
          flexDirection: 'column'
        }}
      >
        {/* Header */}
        <Container sx={{ pt: 8, pb: 2, textAlign: 'center', flexShrink: 0 }} component="div">
          <Typography variant="h3" component="h1" gutterBottom>
            Manimatic
          </Typography>
          <Typography variant="subtitle1" color="text.secondary">
            Generate animations from a single prompt
          </Typography>
        </Container>

        {/* Main Content Area */}
        <Container 
          sx={{ 
            flex: 1, 
            position: 'relative', 
            pb: '100px', 
            display: 'flex', 
            flexDirection: 'column',
            overflow: 'hidden' 
          }}
          component="div"
        >
          <Paper 
            elevation={2} 
            sx={{
              p: 4, 
              borderRadius: 2, 
              bgcolor: 'background.paper', 
              flex: 1, 
              display: 'flex', 
              flexDirection: 'column',
              overflow: 'hidden'
            }}
            component="div"
          >
            {(script || videoUrl) ? (
              <Grid2 
                container 
                spacing={4} 
                sx={{ flex: 1 }}
                component="div"
              >
                {script && (
                  <Grid2 
                    component="div"
                    size={{ xs: 12, sm: 6 }}
                    sx={{ display: 'flex', flexDirection: 'column' }}
                  >
                    <Grow in={true} style={{ transformOrigin: '0 0 0' }} timeout={600}>
                      <Card variant="outlined" sx={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
                        <CardHeader title="Script" />
                        <CardContent 
                          sx={{ 
                            flex: 1, 
                            overflowY: 'auto',
                            '&::-webkit-scrollbar': { width: 6 },
                            scrollbarWidth: 'thin',
                            bgcolor: 'background.default'
                          }}
                        >
                          <Box 
                            component="pre" 
                            sx={{ 
                              whiteSpace: 'pre-wrap', 
                              fontFamily: 'monospace', 
                              fontSize: 14, 
                              color: 'text.primary',
                              m: 0
                            }}
                          >
                            {script}
                          </Box>
                        </CardContent>
                      </Card>
                    </Grow>
                  </Grid2>
                )}

                {videoUrl && (
                  <Grid2 
                    component="div"
                    size={{ xs: 12, sm: 6 }}
                    sx={{ display: 'flex', flexDirection: 'column' }}
                  >
                    <Grow in={true} style={{ transformOrigin: '0 0 0' }} timeout={800}>
                      <Card variant="outlined" sx={{ flex: 1, display: 'flex', flexDirection: 'column'}}>
                        <CardHeader title="Video" />
                        <CardContent 
                          sx={{ 
                            flex: 1, 
                            display: 'flex', 
                            alignItems: 'center', 
                            justifyContent: 'center', 
                            p: 2 
                          }}
                        >
                          <Box
                            component="video"
                            src={videoUrl}
                            controls
                            style={{
                              width: '100%', 
                              height: 'auto',
                              maxHeight: '100%',
                              borderRadius: '4px',
                              objectFit: 'contain'
                            }}
                          />
                        </CardContent>
                      </Card>
                    </Grow>
                  </Grid2>
                )}
              </Grid2>
            ) : (
              !isLoading && (
                <Fade in={!isLoading} timeout={500}>
                  <Typography variant="body1" color="text.secondary" align="center" sx={{ mt: 10 }}>
                    Enter a prompt below to generate your animation...
                  </Typography>
                </Fade>
              )
            )}
          </Paper>
        </Container>

        {/* Input Bar at bottom */}
        <Box
          component="form"
          onSubmit={handleSubmit}
          sx={{ 
            position: 'fixed', 
            bottom: 0, 
            left: 0, 
            width: '100%', 
            bgcolor: 'background.paper', 
            borderTop: '1px solid', 
            borderColor: 'grey.700', 
            py: 2 
          }}
        >
          <Container maxWidth="lg" sx={{ display: 'flex', gap: 2 }} component="div">
            <TextField
              fullWidth
              variant="outlined"
              placeholder="Enter your animation prompt..."
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
              disabled={isLoading}
              InputProps={{
                sx: { bgcolor: 'grey.900', color: 'grey.50' }
              }}
            />
            <Button 
              type="submit" 
              variant="contained" 
              color="primary" 
              disabled={isLoading}
              sx={{ minWidth: '120px' }}
            >
              {isLoading ? (
                <>
                  <CircularProgress size={20} color="inherit" sx={{ mr: 1 }}/>
                  Generating...
                </>
              ) : (
                'Generate'
              )}
            </Button>
          </Container>
        </Box>
      </Box>
    </ThemeProvider>
  );
}

export default App;
