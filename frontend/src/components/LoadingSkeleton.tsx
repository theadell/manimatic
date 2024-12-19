import { Grid, Paper, Typography, Skeleton, Box } from '@mui/material';

export const LoadingSkeleton = () => (
  <Grid container spacing={3}>
    {[0, 1].map((item) => (
      <Grid item xs={12} md={6} key={item}>
        <Paper elevation={1} sx={{ p: 2, height: '100%', display: 'flex', flexDirection: 'column' }}>
          <Typography variant="h6" color="primary" sx={{ mb: 2 }}>
            {item === 0 ? 'Generated Animation' : 'Animation Script'}
          </Typography>
          {item === 0 ? (
            <Skeleton variant="rectangular" width="100%" height={300} />
          ) : (
            <Box sx={{ flexGrow: 1, display: 'flex', flexDirection: 'column', gap: 1 }}>
              {[...Array(5)].map((_, index) => (
                <Skeleton key={index} variant="text" width="100%" height={40} />
              ))}
            </Box>
          )}
        </Paper>
      </Grid>
    ))}
  </Grid>
);
