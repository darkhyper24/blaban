const debugUrls = {
  baseUrl: import.meta.env.VITE_API_URL || 'http://localhost:8081/api',
  authUrl: import.meta.env.VITE_AUTH_API_URL || 'http://localhost:8082/api/auth',
  menuUrl: import.meta.env.VITE_MENU_API_URL || 'http://localhost:8083'
};

console.log('API Config:', debugUrls);

export const apiConfig = debugUrls;