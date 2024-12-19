import React, { useState, useCallback, useRef, useMemo } from 'react';
import { ThemeProvider, CssBaseline, Container, Grid, Snackbar, Alert, Box } from '@mui/material';
import { AnimatePresence } from 'framer-motion';
import * as monaco from 'monaco-editor';
import { createAppTheme } from './theme/theme';
import { useEventSource } from './hooks/useEventSource';
import { VideoPreview } from './components/VideoPreview';
import { ScriptEditor } from './components/ScriptEditor';
import { PromptInput } from './components/PromptInput';
import { Header } from './components/Header';

import { LoadingSkeleton } from './components/LoadingSkeleton';

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL;

function AnimationGenerator() {
  const [isDarkMode, setIsDarkMode] = useState(true);
  const [prompt, setPrompt] = useState('');
  const [script, setScript] = useState('');
  const [editedScript, setEditedScript] = useState('');
  const [videoUrl, setVideoUrl] = useState('');
  const [isGenerating, setIsGenerating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const generationTimeoutRef = useRef<number | null>(null);
  const editorRef = useRef<monaco.editor.IStandaloneCodeEditor | null>(null);

  const theme = useMemo(() => createAppTheme(isDarkMode ? 'dark' : 'light'), [isDarkMode]);

  const handleEditorDidMount = (editor: monaco.editor.IStandaloneCodeEditor) => {
    editorRef.current = editor;
  };

  const resetGeneration = useCallback(() => {
    if (generationTimeoutRef.current) {
      clearTimeout(generationTimeoutRef.current);
    }
    setIsGenerating(false);
    setError('Generation timed out. Please try again.');
  }, []);

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

  const handleCompileClick = useCallback(() => {
    // Show dialog instead of actual compilation
    setError('This feature is currently unavailable while we work on improvements.');
  }, []);

  useEventSource({
    apiBaseUrl,
    onScript: useCallback((receivedScript: string) => {
      setScript(receivedScript);
      setEditedScript(receivedScript);
      setPrompt('');
    }, []),
    onVideo: setVideoUrl,
    onError: setError,
    generationTimeoutRef,
    editorRef,
    setIsGenerating,
  });

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Box sx={{ minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
        <Header 
          isDarkMode={isDarkMode} 
          onToggleTheme={() => setIsDarkMode(!isDarkMode)} 
        />
        
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
          <Box sx={{ flex: 1 }}>
            {isGenerating ? (
              <LoadingSkeleton />
            ) : (
              <Grid container spacing={4} sx={{ height: '100%' }}>
                <Grid item xs={12} lg={6}>
                  <Box
                    sx={{
                      height: '100%',
                      borderRadius: 2,
                      overflow: 'hidden',
                      backgroundColor: 'background.paper'
                    }}
                  >
                    <AnimatePresence>
                      {videoUrl && (
                        <VideoPreview 
                          videoUrl={videoUrl} 
                          onDownload={handleDownloadVideo}
                        />
                      )}
                    </AnimatePresence>
                  </Box>
                </Grid>
                <Grid item xs={12} lg={6}>
                  <Box
                    sx={{
                      height: '100%',
                      borderRadius: 2,
                      overflow: 'hidden',
                      backgroundColor: 'background.paper'
                    }}
                  >
                    <AnimatePresence>
                      {script && (
                        <ScriptEditor
                          script={script}
                          onCopy={handleCopyScript}
                          onScriptChange={(value) => setEditedScript(value || '')}
                          onEditorMount={handleEditorDidMount}
                          onCompileClick={handleCompileClick}
                        />
                      )}
                    </AnimatePresence>
                  </Box>
                </Grid>
              </Grid>
            )}
          </Box>

          <PromptInput
            prompt={prompt}
            isGenerating={isGenerating}
            onChange={setPrompt}
            onSubmit={handleGenerate}
          />
        </Container>

        <Snackbar
          open={!!error}
          autoHideDuration={3000}
          onClose={() => setError(null)}
          anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
        >
          <Alert
            onClose={() => setError(null)}
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
    </ThemeProvider>
  );
}

export default AnimationGenerator;