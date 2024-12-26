import React, { useState } from 'react';
import { Box, Typography, Tab, Tabs } from '@mui/material';
import ErrorOutlineIcon from '@mui/icons-material/ErrorOutline';
import TerminalIcon from '@mui/icons-material/Terminal';
import OutputIcon from '@mui/icons-material/Output';

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

function TabPanel({ children, value, index }: TabPanelProps) {
  return value === index ? (
    <Box 
      sx={{ 
        p: 2,
        maxHeight: '200px',
        minHeight: '48px',
        height: 'auto',
        overflowY: 'auto',
        backgroundColor: theme => 
          theme.palette.mode === 'dark' ? 'rgba(0,0,0,0.2)' : 'rgba(0,0,0,0.05)'
      }}
    >
      {children}
    </Box>
  ) : null;
}

interface ErrorDisplayProps {
  error: {
    message: string;
    stdout?: string;
    stderr?: string;
    line?: number;
  };
}

export const ErrorDisplay = ({ error }: ErrorDisplayProps) => {
  const [tabValue, setTabValue] = useState(0);
  const hasStdout = Boolean(error.stdout?.trim());
  const hasStderr = Boolean(error.stderr?.trim());

  const getTabs = () => {
    const tabs = [{
      icon: <ErrorOutlineIcon sx={{ fontSize: 16 }} />,
      label: "Error Message",
      content: error.message
    }];
    
    if (hasStdout) {
      tabs.push({
        icon: <OutputIcon sx={{ fontSize: 16 }} />,
        label: "Standard Output",
        content: error.stdout!
      });
    }
    
    if (hasStderr) {
      tabs.push({
        icon: <TerminalIcon sx={{ fontSize: 16 }} />,
        label: "Standard Error",
        content: error.stderr!
      });
    }
    
    return tabs;
  };

  const tabs = getTabs();

  return (
    <Box
      sx={{
        mt: 2,
        border: '1px solid',
        borderColor: 'error.light',
        borderRadius: 1,
        overflow: 'hidden',
        backgroundColor: theme => 
          theme.palette.mode === 'dark' ? 'rgba(0,0,0,0.2)' : 'rgba(255,255,255,0.9)'
      }}
    >
      <Box
        sx={{
          px: 2,
          py: 1.5,
          backgroundColor: theme => 
            theme.palette.mode === 'dark' ? 'rgba(255,0,0,0.08)' : 'rgba(255,0,0,0.03)',
          borderBottom: '1px solid',
          borderColor: 'error.light',
          display: 'flex',
          alignItems: 'center',
          gap: 1
        }}
      >
        <ErrorOutlineIcon 
          color="error" 
          sx={{ fontSize: 18 }}
        />
        <Typography 
          color="error.main" 
          sx={{ 
            fontSize: '0.9rem',
            fontWeight: 500,
            lineHeight: 1
          }}
        >
          Compilation Error{error.line ? ` at Line ${error.line}` : ''}
        </Typography>
      </Box>

      {tabs.length > 1 && (
        <Tabs 
          value={tabValue} 
          onChange={(_, newValue) => setTabValue(newValue)}
          sx={{ 
            minHeight: 42,
            borderBottom: 1, 
            borderColor: 'divider',
            '& .MuiTab-root': {
              minHeight: 42,
              py: 0.5,
              fontSize: '0.8rem'
            }
          }}
        >
          {tabs.map((tab, index) => (
            <Tab 
              key={index}
              icon={tab.icon}
              iconPosition="start"
              label={tab.label}
              sx={{ 
                textTransform: 'none',
                gap: 1
              }}
            />
          ))}
        </Tabs>
      )}

      {tabs.map((tab, index) => (
        <TabPanel key={index} value={tabValue} index={index}>
          <Typography 
            variant="body2" 
            sx={{ 
              fontSize: '0.85rem',
              fontFamily: 'Consolas, Monaco, "Courier New", monospace',
              whiteSpace: 'pre-wrap',
              wordBreak: 'break-word',
              margin: 0,
              lineHeight: 1.5
            }}
          >
            {tab.content || 'No content available'}
          </Typography>
        </TabPanel>
      ))}
    </Box>
  );
};