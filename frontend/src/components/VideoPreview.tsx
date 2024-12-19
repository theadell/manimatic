import { Box, IconButton, Paper, Typography } from '@mui/material';
import DownloadIcon from '@mui/icons-material/Download';
import { motion } from 'framer-motion';

export const VideoPreview = ({ videoUrl, onDownload }: { videoUrl: string; onDownload: () => void }) => (
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
            Generated Animation
          </Typography>
          <IconButton
            color="primary"
            onClick={onDownload}
            title="Download Video"
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
          <video
            src={videoUrl}
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
        </Box>
      </Paper>
    </motion.div>
  );
  