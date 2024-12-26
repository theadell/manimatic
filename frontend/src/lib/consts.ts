// List of common video file extensions
export const VIDEO_EXTENSIONS = ['.mp4', '.webm', '.mov', '.avi', '.mkv', '.flv', '.wmv', '.m4v'];

// Regex for matching video URLs (case-insensitive, accounts for query parameters)
export const VIDEO_REGEX = /\.(mp4|webm|mov|avi|mkv|flv|wmv|m4v)(\?.*)?$/i;
