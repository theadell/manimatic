import { Grid, Typography, Skeleton, Box } from '@mui/material';
import { AnimatePresence } from 'framer-motion';
import { AnimationPreview } from './AnimationPreview';
import { ScriptEditor } from './ScriptEditor';
import * as monaco from 'monaco-editor';
import { CompileError } from '../types/types';



type ContentPreviewProps = {
  isVideoLoading: boolean;
  isScriptLoading: boolean;
  isCompiling: boolean;
  animationUrl: string;
  script: string;
  compilationError?: CompileError;
  onDownload: () => void;
  onCopy: () => void;
  onScriptChange: (value: string | undefined) => void;
  onEditorMount: (editor: monaco.editor.IStandaloneCodeEditor) => void;
  onCompileClick: () => void;
}

export const ContentPreview = ({
  isVideoLoading,
  isScriptLoading,
  isCompiling,
  animationUrl: videoUrl,
  script,
  compilationError,
  onDownload,
  onCopy,
  onScriptChange,
  onEditorMount,
  onCompileClick
}: ContentPreviewProps) => {

  const shouldRender = isVideoLoading || isScriptLoading || videoUrl || script;
  if (!shouldRender) {
    return null
  }

  return (
    <Grid container spacing={4} sx={{ height: '100%' }}>
      <Grid item xs={12} lg={6}>
        <Box
          sx={{
            height: '100%',
            borderRadius: 2,
            overflow: 'hidden',
            backgroundColor: 'background.paper',
            p: 3
          }}
        >
          {isVideoLoading ? (
            <>
              <Typography variant="h6" sx={{ mb: 2, display: 'flex', alignItems: 'center', gap: 1 }}>
                <Skeleton variant="circular" width={24} height={24} />
                <Skeleton width={160} />
              </Typography>
              <Skeleton
                variant="rectangular"
                width="100%"
                height={400}
                animation="wave"
                sx={{ borderRadius: 1, opacity: 0.8 }}
              />
            </>
          ) : (
            <AnimatePresence>
              {videoUrl && (
                <AnimationPreview
                  url={videoUrl}
                  onDownload={onDownload}
                />
              )}
            </AnimatePresence>
          )}
        </Box>
      </Grid>
      <Grid item xs={12} lg={6}>
        <Box
          sx={{
            height: '100%',
            borderRadius: 2,
            overflow: 'hidden',
            backgroundColor: 'background.paper',
            p: 3
          }}
        >
          {isScriptLoading ? (
            <>
              <Typography variant="h6" sx={{ mb: 2, display: 'flex', alignItems: 'center', gap: 1 }}>
                <Skeleton variant="circular" width={24} height={24} />
                <Skeleton width={140} />
              </Typography>
              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.5 }}>
                {[...Array(6)].map((_, index) => (
                  <Skeleton
                    key={index}
                    variant="text"
                    width={`${Math.random() * (100 - 70) + 70}%`}
                    height={24}
                    animation="wave"
                    sx={{ opacity: 0.8 }}
                  />
                ))}
              </Box>
            </>
          ) : (
            <AnimatePresence>
              {script && (
                <ScriptEditor
                  script={script}
                  error={compilationError}
                  onCopy={onCopy}
                  onScriptChange={onScriptChange}
                  onEditorMount={onEditorMount}
                  onCompileClick={onCompileClick}
                  isCompiling={isCompiling}
                />
              )}
            </AnimatePresence>
          )}
        </Box>
      </Grid>
    </Grid>
  );
}
