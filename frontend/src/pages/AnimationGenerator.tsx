import React, { useState, useCallback, useRef, useMemo } from 'react';
import { Box } from '@mui/material';
import * as monaco from 'monaco-editor';
import { createAppTheme } from '../theme/theme';
import { useEventSource } from '../hooks/useEventSource';
import { PromptInput } from '../components/PromptInput';
import { useFeatures } from '../hooks/useFeatures';
import { ContentPreview } from '../components/ContentPreview';
import { Layout } from '../components/Layout';
import { ThemeProvider, CssBaseline } from '@mui/material';

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL;
const GENERATION_TIMEOUT = 30_000;

function AnimationGenerator() {
  const [isDarkMode, setIsDarkMode] = useState(true);
  const [prompt, setPrompt] = useState('');
  const [script, setScript] = useState(``);
  const [editedScript, setEditedScript] = useState('');
  const [videoUrl, setVideoUrl] = useState('');
  const [isVideoLoading, setIsVideoLoading] = useState(false);
  const [isScriptLoading, setIsScriptLoading] = useState(false);
  const [isCompiling, setIsCompiling] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  const generationTimeoutRef = useRef<number | null>(null);
  const editorRef = useRef<monaco.editor.IStandaloneCodeEditor | null>(null);

  const theme = useMemo(() => createAppTheme(isDarkMode ? 'dark' : 'light'), [isDarkMode]);
  const { isFeatureEnabled } = useFeatures(apiBaseUrl);

  const handleEditorDidMount = useCallback((editor: monaco.editor.IStandaloneCodeEditor) => {
    editorRef.current = editor;
  }, []);

  const resetGeneration = useCallback(() => {
    if (generationTimeoutRef.current) {
      clearTimeout(generationTimeoutRef.current);
    }
    setIsScriptLoading(false);
    setIsVideoLoading(false);
    setError('Generation timed out. Please try again.');
  }, []);

  const handleGenerate = useCallback(async (e?: React.FormEvent) => {
    if (e) e.preventDefault();

    if (!prompt.trim()) {
      setError('Please enter a prompt');
      return;
    }

    setIsScriptLoading(true);
    setIsVideoLoading(true);
    setScript('');
    setEditedScript('');
    setVideoUrl('');
    setError(null);

    generationTimeoutRef.current = setTimeout(resetGeneration, GENERATION_TIMEOUT);

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
      setIsVideoLoading(false);
      setIsScriptLoading(false);
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

  const handleCompileClick = useCallback(async () => {
    if (!isFeatureEnabled('user-compile')) {
      setError('Compilation feature is currently unavailable');
      return;
    }

    setIsCompiling(true);
    setIsVideoLoading(true);

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
      setError('Failed to compile script. Please try again.');
      setIsCompiling(false);
      setIsVideoLoading(false);
    }
  }, [editedScript, isFeatureEnabled]);

  useEventSource({
    apiBaseUrl,
    onScript: useCallback((receivedScript: string) => {
      setScript(receivedScript);
      setEditedScript(receivedScript);
      setPrompt('');
    }, []),
    onVideo: useCallback((url: string) => {
      setVideoUrl(url);
      setIsVideoLoading(false);
      setIsCompiling(false);
    }, []),
    onError: setError,
    generationTimeoutRef,
    editorRef,
    setIsScriptLoading,
    setIsVideoLoading
  });

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Layout
        isDarkMode={isDarkMode}
        onToggleTheme={() => setIsDarkMode(!isDarkMode)}
        error={error}
        onErrorClose={() => setError(null)}
      >
        <Box sx={{ flex: 1 }}>
          <ContentPreview
            isVideoLoading={isVideoLoading}
            isScriptLoading={isScriptLoading}
            isCompiling={isCompiling}
            videoUrl={videoUrl}
            script={script}
            onDownload={handleDownloadVideo}
            onCopy={handleCopyScript}
            onScriptChange={(value) => setEditedScript(value || '')}
            onEditorMount={handleEditorDidMount}
            onCompileClick={handleCompileClick}
          />
        </Box>

        <PromptInput
          prompt={prompt}
          isGenerating={isVideoLoading || isScriptLoading}
          onChange={setPrompt}
          onSubmit={handleGenerate}
        />
      </Layout>
    </ThemeProvider>
  );
}

export default AnimationGenerator;