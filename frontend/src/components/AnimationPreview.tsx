import { Box, IconButton, Paper, Typography } from '@mui/material';
import DownloadIcon from '@mui/icons-material/Download';
import { motion } from 'framer-motion';
import { VIDEO_EXTENSIONS, VIDEO_REGEX } from '../lib/consts';

type PreviewProps = {
  url: string;
  onDownload: () => void;
};

function isVideoURL(url: string): boolean {
  try {
    const urlObj = new URL(url);
    const pathname = urlObj.pathname.toLowerCase();

    return VIDEO_EXTENSIONS.some(ext => pathname.endsWith(ext));
  } catch (e) {
    console.error('Invalid URL:', e);

    // Fallback to regex check
    return VIDEO_REGEX.test(url);
  }
}




export const AnimationPreview = ({ url, onDownload }: PreviewProps) => {

  const isVideo = isVideoURL(url);

  return (
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
            {isVideo ? 'Generated Animation' : 'Generated Image'}
          </Typography>
          <IconButton
            color="primary"
            onClick={onDownload}
            title={`Download ${isVideo ? 'Video' : 'Image'}`}
            size="small"
            sx={{ 
              backgroundColor: 'action.hover',
              '&:hover': { backgroundColor: 'action.selected' }
            }}
          >
            <DownloadIcon fontSize="small" />
          </IconButton>
        </Box>
        <Box
          sx={{
            position: 'relative',
            width: '100%',
            pt: '56.25%', // 16:9 aspect ratio
            borderRadius: 2,
            overflow: 'hidden',
            backgroundColor: 'background.default'
          }}
        >
          {isVideo ? (
            <video
              src={url}
              controls
              style={{
                position: 'absolute',
                top: 0,
                left: 0,
                width: '100%',
                height: '100%',
                objectFit: 'contain'
              }}
            />
          ) : (
            <img
              src={url}
              alt="Generated visualization"
              style={{
                position: 'absolute',
                top: 0,
                left: 0,
                width: '100%',
                height: '100%',
                objectFit: 'contain'
              }}
            />
          )}
        </Box>
      </Paper>
    </motion.div>
  );
};