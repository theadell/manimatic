import { useState, useEffect } from 'react'

type FeatureKey = string

interface Feature {
  key: FeatureKey;
  description: string;
  enabled: boolean;
}

interface FeaturesResponse {
  version: string;
  features: Feature[];
}

export const useFeatures = (apiBaseUrl: string) => {
  const [features, setFeatures] = useState<Record<FeatureKey, boolean>>({});
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchFeatures = async () => {
      try {
        const response = await fetch(`${apiBaseUrl}/features`, {
          credentials: 'include'
        });
        
        if (!response.ok) {
          throw new Error('Failed to fetch features');
        }

        const data: FeaturesResponse = await response.json();
        const featureMap = data.features.reduce((acc, feature) => ({
          ...acc,
          [feature.key]: feature.enabled
        }), {});

        setFeatures(featureMap);
      } catch (err) {
        setError('Failed to load features');
        console.error('Error fetching features:', err);
      } finally {
        setIsLoading(false);
      }
    };

    fetchFeatures()
  }, [apiBaseUrl])

  const isFeatureEnabled = (key: FeatureKey) => features[key] ?? false

  return { isFeatureEnabled, isLoading, error }
}