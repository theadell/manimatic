import { Box, TextField, IconButton, CircularProgress, Select, MenuItem, FormControl, InputLabel } from '@mui/material';
import SendIcon from '@mui/icons-material/Send';

interface PromptInputProps {
  prompt: string;
  isGenerating: boolean;
  onChange: (value: string) => void;
  onSubmit: (e: React.FormEvent) => void;
  models: string[];
  selectedModel: string;
  onModelChange: (model: string) => void;
  isLoadingModels: boolean;
}


export const PromptInput = ({
  prompt,
  isGenerating,
  onChange,
  onSubmit,
  models,
  selectedModel,
  onModelChange,
  isLoadingModels
}: PromptInputProps) => (
  <Box
    component="form"
    onSubmit={onSubmit}
    sx={{
      mt: 2,
      width: '100%',
      display: 'flex',
      gap: 2
    }}
  >
    <FormControl sx={{ minWidth: 120 }}>
      <InputLabel id="model-select-label">Model</InputLabel>
      <Select
        labelId="model-select-label"
        value={selectedModel}
        label="Model"
        onChange={(e) => onModelChange(e.target.value)}
        disabled={isGenerating || isLoadingModels}
        size="small"
      >
        {isLoadingModels ? (
          <MenuItem value="">
            <CircularProgress size={20} />
          </MenuItem>
        ) : (
          models.map(model => (
            <MenuItem key={model} value={model}>
              {model}
            </MenuItem>
          ))
        )}
      </Select>
    </FormControl>

    <TextField
      fullWidth
      variant="outlined"
      placeholder="Describe your animation..."
      value={prompt}
      onChange={(e) => onChange(e.target.value)}
      disabled={isGenerating || !selectedModel}
      size="small"
      InputProps={{
        endAdornment: (
          <IconButton
            color="primary"
            type="submit"
            disabled={isGenerating || !selectedModel}
            size="small"
          >
            {isGenerating ? <CircularProgress size={24} /> : <SendIcon fontSize="small" />}
          </IconButton>
        )
      }}
    />
  </Box>
);
