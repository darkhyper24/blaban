import { apiConfig } from './config';

// Types
export interface User {
  id: string;
  email: string;
  name: string;
  provider: string;
  created_at: string;
  updated_at: string;
}

export interface AuthResponse {
  user: User;
  accessToken: string;
  refreshToken: string;
  tokenType: string;
  expiresIn: number;
}

export interface SignUpRequest {
  email: string;
  password: string;
  name: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RefreshTokenRequest {
  refresh_token: string;
}

export interface AuthApi {
  signUp: (data: SignUpRequest) => Promise<AuthResponse>;
  login: (data: LoginRequest) => Promise<AuthResponse>;
  getGoogleAuthUrl: () => Promise<void>;
  refreshToken: (token: string) => Promise<AuthResponse>;
}

// API Client
const authApi: AuthApi = {
  // Sign up with email and password
  signUp: async (data: SignUpRequest): Promise<AuthResponse> => {
    const response = await fetch(`${apiConfig.baseUrl}/users/signup`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Failed to sign up');
    }

    return response.json();
  },

  // Login with email and password
  login: async (data: LoginRequest): Promise<AuthResponse> => {
    const response = await fetch(`${apiConfig.baseUrl}/users/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Failed to login');
    }

    return response.json();
  },

  getGoogleAuthUrl: async (): Promise<void> => {  // Changed return type to void
    try {
      const response = await fetch(`${apiConfig.authUrl}/google/login`, {
        method: 'GET',
        headers: {
          'Accept': 'application/json',
          'Origin': window.location.origin
        },
        // Don't follow redirects automatically
        redirect: 'manual'
      });
      
      // Check if it's a redirect response
      if (response.status === 302 || response.type === 'opaqueredirect') {
        const redirectUrl = response.headers.get('Location') || response.url;
        window.location.href = redirectUrl;
        return;
      }

      if (!response.ok) {
        console.error('Google auth error:', response.status, response.statusText);
        throw new Error('Failed to get Google auth URL');
      }
    } catch (error) {
      console.error('Google auth error:', error);
      throw error;
    }
  },

  refreshToken: async (token: string): Promise<AuthResponse> => {
    const response = await fetch('/api/auth/refresh', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ refreshToken: token }),
    });
    
    if (!response.ok) {
      throw new Error('Failed to refresh token');
    }
    
    return response.json();
  }
};

export default authApi;