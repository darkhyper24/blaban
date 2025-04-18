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

// API Client
const authApi = {
  // Sign up with email and password
  signUp: async (data: SignUpRequest): Promise<AuthResponse> => {
    const response = await fetch(`${apiConfig.baseUrl}/auth/signup`, {
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
    const response = await fetch(`${apiConfig.baseUrl}/auth/login`, {
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

  // Get Google OAuth URL
  getGoogleAuthUrl: async (): Promise<string> => {
    const response = await fetch(`${apiConfig.baseUrl}/auth/google`);
    
    if (!response.ok) {
      throw new Error('Failed to get Google auth URL');
    }
    
    const data = await response.json();
    return data.url;
  },

  // Refresh tokens
  refreshToken: async (refreshToken: string): Promise<AuthResponse> => {
    const response = await fetch(`${apiConfig.baseUrl}/auth/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    if (!response.ok) {
      throw new Error('Failed to refresh token');
    }

    return response.json();
  },
};

export default authApi;