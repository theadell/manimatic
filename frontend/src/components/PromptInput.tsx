import { Box, TextField, IconButton, CircularProgress } from '@mui/material';
import SendIcon from '@mui/icons-material/Send';

interface PromptInputProps {
  prompt: string;
  isGenerating: boolean;
  onChange: (value: string) => void;
  onSubmit: (e: React.FormEvent) => void;
}

export const PromptInput = ({ prompt, isGenerating, onChange, onSubmit }: PromptInputProps) => (
  <Box
    component="form"
    onSubmit={onSubmit}
    sx={{ mt: 2, width: '100%' }}
  >
    <TextField
      fullWidth
      variant="outlined"
      placeholder="Describe your animation..."
      value={prompt}
      onChange={(e) => onChange(e.target.value)}
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
)