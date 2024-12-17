import React, { useState, useCallback, useEffect, useRef } from 'react';
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
  CircularProgress,
  Skeleton,
  Button,
  Dialog,
  DialogContent,
  DialogContentText,
  DialogTitle,
  DialogActions
} from '@mui/material';
import SendIcon from '@mui/icons-material/Send';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import DownloadIcon from '@mui/icons-material/Download';
import { motion, AnimatePresence } from 'framer-motion';
import Editor from '@monaco-editor/react';
import * as monaco from 'monaco-editor';

const theme = createTheme({
  palette: {
    mode: 'dark',
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
          borderRadius: 8, // Reduced roundness
          boxShadow: '0 4px 6px rgba(0,0,0,0.1)',
          transition: 'box-shadow 0.3s ease'
        }
      }
    },
    MuiTextField: {
      styleOverrides: {
        root: {
          '& .MuiOutlinedInput-root': {
            borderRadius: 6, // Slightly reduced roundness
            backgroundColor: 'rgba(255,255,255,0.05)',
            transition: 'background-color 0.2s ease'
          }
        }
      }
    }
  }
});


interface MessageType {
  type: 'script' | 'video' | 'compiled';
  sessionId: string;
  status: 'success' | 'error';
  content: string;
  details?: Record<string, unknown>;
}

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL;
function AnimationGenerator() {
  // State Management
  const [prompt, setPrompt] = useState('');
  const [script, setScript] = useState('');
  const [open, setOpen] = React.useState(false);
  const [editedScript, setEditedScript] = useState('');
  const [videoUrl, setVideoUrl] = useState('');
  const [compiledResult, setCompiledResult] = useState('');
  const [isGenerating, setIsGenerating] = useState(false);
  // const [isCompiling, setIsCompiling] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const generationTimeoutRef = useRef<number | null>(null);
  const editorRef = useRef<monaco.editor.IStandaloneCodeEditor | null>(null);

  const handleClickOpen = () => {
    setOpen(true);
  };

  const handleClose = () => {
    setOpen(false);
  };

  const resetGeneration = useCallback(() => {
    if (generationTimeoutRef.current) {
      clearTimeout(generationTimeoutRef.current);
    }
    setIsGenerating(false);
    setError('Generation timed out. Please try again.');
  }, []);

  const handleEditorDidMount = (editor: monaco.editor.IStandaloneCodeEditor) => {
    editorRef.current = editor;
  };

  const handleGenerate = useCallback(async (e?: React.FormEvent) => {
    if (e) e.preventDefault();

    if (!prompt.trim()) {
      setError('Please enter a prompt');
      return;
    }

    setIsGenerating(true);
    setScript('');
    setEditedScript('');
    setVideoUrl('');
    setCompiledResult('');
    setError(null);

    generationTimeoutRef.current = setTimeout(resetGeneration, 8000);

    try {
      const response = await fetch(`${apiBaseUrl}/generate`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ prompt }),
        credentials: 'include'
      });

      if (!response.ok) {
        throw new Error('Generation failed');
      }
    } catch (error) {
      console.error('Error generating content:', error);
      setError('Failed to generate animation. Please try again.');
      setIsGenerating(false);
    }
  }, [prompt, resetGeneration]);

  // Compile Script Handler
  /*
  const handleCompileScript = useCallback(async () => {
    setIsCompiling(true);
    setCompiledResult('');
    setError(null);

    try {
      const response = await fetch(`${apiBaseUrl}/compile`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ script: editedScript }),
        credentials: 'include'
      });

      if (!response.ok) {
        throw new Error('Compilation failed');
      }
    } catch (error) {
      console.error('Error compiling script:', error);
      setError('Failed to compile script. Please check your code.');
      setIsCompiling(false);
    }
  }, [editedScript]);
 */
  // Script Copy Handler
  const handleCopyScript = useCallback(() => {
    const scriptToCopy = editedScript || script;
    if (scriptToCopy) {
      navigator.clipboard.writeText(scriptToCopy)
        .then(() => {
          setError('Script copied to clipboard');
          setTimeout(() => setError(null), 2000);
        })
        .catch(() => setError('Failed to copy script'));
    }
  }, [script, editedScript]);

  // Video Download Handler
  const handleDownloadVideo = useCallback(() => {
    if (videoUrl) {
      const link = document.createElement('a');
      link.href = videoUrl;
      link.download = 'animation.mp4';
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    }
  }, [videoUrl]);

  // Server-Sent Events Effect
  useEffect(() => {
    const initializeEventSource = async () => {
      try {

        const healthzResponse = await fetch(`${apiBaseUrl}/healthz`, {
          method: 'GET',
          credentials: 'include',
        });

        if (!healthzResponse.ok) {
          throw new Error(`Health check failed with status: ${healthzResponse.status}`);
        }

        const eventSource = new EventSource(`${apiBaseUrl}/events`, { withCredentials: true });

        eventSource.onmessage = (event) => {
          const message: MessageType = JSON.parse(event.data);

          if (generationTimeoutRef.current) {
            clearTimeout(generationTimeoutRef.current);
          }

          switch (message.type) {
            case 'script':
              {
                const receivedScript = message.content;
                setScript(receivedScript);
                setEditedScript(receivedScript);
                setPrompt('');

                if (editorRef.current) {
                  editorRef.current.setValue(receivedScript);
                }

                setIsGenerating(false);
                break;
              }
            case 'video':
              setVideoUrl(message.content);
              setIsGenerating(false);
              break;
            //case 'compiled':
             // setCompiledResult(message.content);
              //setIsCompiling(false);
              //break;
          }
        };

        eventSource.onerror = (error) => {
          console.error('EventSource failed:', error);
          eventSource.close();
          setIsGenerating(false);
          //setIsCompiling(false);
          setError('Connection error. Please refresh and try again.');
        };

        return () => {
          if (generationTimeoutRef.current) {
            clearTimeout(generationTimeoutRef.current);
          }
          eventSource.close();
        };
      } catch (error) {
        console.error('Initialization failed:', error);
        setError('Health check failed. Unable to connect to the server.');
        setIsGenerating(false);
        //setIsCompiling(false);
      }
    };

    initializeEventSource()
  }, []);

  const SkeletonLoader = () => (
    <Grid container spacing={3}>
      {[0, 1].map((item) => (
        <Grid item xs={12} md={6} key={item}>
          <Paper
            elevation={1}
            sx={{
              p: 2,
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
              {item === 0 ? 'Generated Animation' : 'Animation Script'}
            </Typography>
            {item === 0 ? (
              <Skeleton
                variant="rectangular"
                width="100%"
                height={300}
              />
            ) : (
              <Box sx={{ flexGrow: 1, display: 'flex', flexDirection: 'column', gap: 1 }}>
                {[...Array(5)].map((_, index) => (
                  <Skeleton
                    key={index}
                    variant="text"
                    width="100%"
                    height={40}
                  />
                ))}
              </Box>
            )}
          </Paper>
        </Grid>
      ))}
    </Grid>
  );

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Container
        maxWidth="lg"
        sx={{
          display: 'flex',
          flexDirection: 'column',
          minHeight: '100vh',
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

        <Box
          sx={{
            flexGrow: 1,
            display: 'flex',
            flexDirection: 'column',
            overflow: 'hidden'
          }}
        >
          {isGenerating ? (
            <SkeletonLoader />
          ) : (
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
                      initial={{ opacity: 0, scale: 0.95 }}
                      animate={{ opacity: 1, scale: 1 }}
                      exit={{ opacity: 0, scale: 0.95 }}
                    >
                      <Paper
                        elevation={1}
                        sx={{
                          p: 2,
                          height: '100%',
                          display: 'flex',
                          flexDirection: 'column'
                        }}
                      >
                        <Box sx={{
                          display: 'flex',
                          justifyContent: 'space-between',
                          alignItems: 'center',
                          mb: 2
                        }}>
                          <Typography
                            variant="h6"
                            color="primary"
                          >
                            Generated Animation
                          </Typography>
                          <IconButton
                            color="primary"
                            onClick={handleDownloadVideo}
                            title="Download Video"
                            size="small"
                          >
                            <DownloadIcon fontSize="small" />
                          </IconButton>
                        </Box>
                        <video
                          src={videoUrl}
                          controls
                          style={{
                            width: '100%',
                            borderRadius: 6,
                            boxShadow: '0 2px 4px rgba(0,0,0,0.1)'
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
                      initial={{ opacity: 0, scale: 0.95 }}
                      animate={{ opacity: 1, scale: 1 }}
                      exit={{ opacity: 0, scale: 0.95 }}
                    >
                      <Paper
                        elevation={1}
                        sx={{
                          p: 2,
                          height: '100%',
                          position: 'relative',
                          display: 'flex',
                          flexDirection: 'column'
                        }}
                      >
                        <Box sx={{
                          display: 'flex',
                          justifyContent: 'space-between',
                          alignItems: 'center',
                          mb: 2
                        }}>
                          <Typography
                            variant="h6"
                            color="primary"
                          >
                            Animation Script
                          </Typography>
                          <Box>
                            <IconButton
                              color="primary"
                              onClick={handleCopyScript}
                              title="Copy Script"
                              size="small"
                              sx={{ mr: 1 }}
                            >
                              <ContentCopyIcon fontSize="small" />
                            </IconButton>
                            {/* 
                            <Button
                              variant="contained"
                              color="primary"
                              startIcon={<PlayArrowIcon />}
                              size="small"
                              disabled={isCompiling}
                              onClick={handleCompileScript}
                            >
                              {isCompiling ? 'Compiling...' : 'Compile'}
                            </Button>
                            */}
                            <Button
                              variant="contained"
                              color="primary"
                              startIcon={<PlayArrowIcon />}
                              size="small"
                              onClick={handleClickOpen}
                            >
                               {'Compile'}
                            </Button>

                            <Dialog
                              open={open}
                              onClose={handleClose}
                              aria-labelledby="alert-dialog-title"
                              aria-describedby="alert-dialog-description"
                            >
                              <DialogTitle id="alert-dialog-title">
                              {"This Feature is Taking a Break"}
                              </DialogTitle>
                              <DialogContent>
                                <DialogContentText id="alert-dialog-description">
                                This feature is taking a short break while we work on some improvements.
                                It 'll be back soon!
                                </DialogContentText>
                              </DialogContent>
                              <DialogActions>
                                <Button onClick={handleClose}>Alright</Button>
                              </DialogActions>
                            </Dialog>

                          </Box>
                        </Box>
                        <Box
                          sx={{
                            flexGrow: 1,
                            height: '100%', // Fixed height
                            position: 'relative',
                            borderRadius: 1,
                            overflow: 'hidden',
                            border: '1px solid rgba(255,255,255,0.1)'
                          }}
                        >
                          <Editor
                            height="600px"
                            defaultLanguage="python"
                            defaultValue={script}
                            onMount={handleEditorDidMount}
                            onChange={(value) => setEditedScript(value || '')}
                            theme="vs-dark"
                            options={{
                              minimap: { enabled: false },
                              fontSize: 14,
                              lineHeight: 24,
                              padding: { top: 15, bottom: 15 }
                            }}
                          />
                        </Box>
                        {compiledResult && (
                          <Box
                            sx={{
                              mt: 2,
                              p: 2,
                              backgroundColor: 'rgba(255,255,255,0.05)',
                              borderRadius: 1,
                              maxHeight: 150,
                              overflowY: 'auto'
                            }}
                          >
                            <Typography
                              variant="body2"
                              color="text.secondary"
                            >
                              Compilation Result:
                            </Typography>
                            <Typography
                              component="pre"
                              variant="body2"
                              sx={{
                                fontFamily: 'monospace',
                                whiteSpace: 'pre-wrap',
                                margin: 0,
                                color: 'text.secondary'
                              }}
                            >
                              {compiledResult}
                            </Typography>
                          </Box>
                        )}
                      </Paper>
                    </motion.div>
                  )}
                </AnimatePresence>
              </Grid>
            </Grid>
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
                  size="small"
                >
                  {isGenerating ? <CircularProgress size={24} /> : <SendIcon fontSize="small" />}
                </IconButton>
              )
            }}
          />
        </Box>

        <Snackbar
          open={!!error}
          autoHideDuration={3000}
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