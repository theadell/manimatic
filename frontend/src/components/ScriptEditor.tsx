import { Box, IconButton, Paper, Typography, Button } from '@mui/material';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import { motion } from 'framer-motion';
import Editor from '@monaco-editor/react';
import * as monaco from 'monaco-editor';

interface ScriptEditorProps {
  script: string;
  isCompiling: boolean;
  onCopy: () => void;
  onScriptChange: (value: string | undefined) => void;
  onEditorMount: (editor: monaco.editor.IStandaloneCodeEditor) => void;
  onCompileClick: () => void;
  compiledResult?: string;
}

export const ScriptEditor = ({
  script,
  isCompiling,
  onCopy,
  onScriptChange,
  onEditorMount,
  onCompileClick,
  compiledResult
}: ScriptEditorProps) => (
  <motion.div
    initial={{ opacity: 0, scale: 0.95 }}
    animate={{ opacity: 1, scale: 1 }}
    exit={{ opacity: 0, scale: 0.95 }}
  >
    <Paper
      elevation={2}
      sx={{
        p: 3,
        height: '100%',
        position: 'relative',
        display: 'flex',
        flexDirection: 'column',
        backgroundColor: theme => theme.palette.mode === 'dark' ? 'rgba(0,0,0,0.2)' : 'rgba(255,255,255,0.9)',
        backdropFilter: 'blur(10px)',
        borderRadius: 2,
        border: '1px solid',
        borderColor: 'divider'
      }}
    >
      <Box sx={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        mb: 2,
        pb: 2,
        borderBottom: '1px solid',
        borderColor: 'divider'
      }}>
        <Typography variant="h6" color="primary" sx={{ fontWeight: 600 }}>
          Animation Script
        </Typography>
        <Box sx={{ display: 'flex', gap: 1 }}>
          <IconButton
            color="primary"
            onClick={onCopy}
            title="Copy Script"
            size="small"
            sx={{ 
              backgroundColor: 'action.hover',
              '&:hover': { backgroundColor: 'action.selected' }
            }}
          ><ContentCopyIcon fontSize="small" />
          </IconButton>
          <Button
            variant="contained"
            color="primary"
            startIcon={<PlayArrowIcon />}
            size="small"
            onClick={onCompileClick}
            disabled={isCompiling}
            sx={{ 
              boxShadow: 2,
              '&:hover': { boxShadow: 4 },
              minWidth: '100px'
            }}
          >
            {isCompiling ? 'Compiling...' : 'Compile'}
          </Button>
        </Box>
      </Box>
      <Box
        sx={{
          flexGrow: 1,
          position: 'relative',
          borderRadius: 1,
          overflow: 'hidden',
          border: '1px solid',
          borderColor: 'divider',
          opacity: isCompiling ? 0.7 : 1,
          pointerEvents: isCompiling ? 'none' : 'auto',
          '& .monaco-editor': {
            paddingTop: 1,
            paddingBottom: 1
          }
        }}
      >
        <Editor
          height="600px"
          defaultLanguage="python"
          defaultValue={script}
          onMount={onEditorMount}
          onChange={onScriptChange}
          theme="vs-dark"
          options={{
            minimap: { enabled: false },
            fontSize: 14,
            lineHeight: 24,
            padding: { top: 16, bottom: 16 },
            scrollBeyondLastLine: false,
            smoothScrolling: true,
            cursorBlinking: 'smooth',
            readOnly: isCompiling
          }}
        />
      </Box>
      {compiledResult && (
        <Box
          sx={{
            mt: 2,
            p: 2,
            backgroundColor: 'background.default',
            borderRadius: 1,
            maxHeight: 150,
            overflowY: 'auto',
            border: '1px solid',
            borderColor: 'divider'
          }}
        >
          <Typography variant="body2" color="text.secondary" gutterBottom>
            Compilation Result:
          </Typography>
          <Typography
            component="pre"
            variant="body2"
            sx={{
              fontFamily: 'monospace',
              whiteSpace: 'pre-wrap',
              margin: 0,
              color: 'text.primary',
              p: 1,
              backgroundColor: 'background.paper',
              borderRadius: 1
            }}
          >
            {compiledResult}
          </Typography>
        </Box>
      )}
    </Paper>
  </motion.div>
);