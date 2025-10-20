import { post, get } from './client';
import type { LoginRequest, LoginResponse, HealthResponse } from '@/types/api';

/**
 * Authentication API endpoints
 */

export const authApi = {
  /**
   * Login admin user
   */
  login: async (credentials: LoginRequest): Promise<LoginResponse> => {
    const response = await post<LoginResponse, LoginRequest>('/api/v1/auth/login', credentials);

    // Store token and user data in localStorage
    if (response.token) {
      localStorage.setItem('admin_token', response.token);
    }
    if (response.user) {
      localStorage.setItem('admin_user', JSON.stringify(response.user));
    }

    return response;
  },

  /**
   * Logout admin user
   */
  logout: async (): Promise<void> => {
    try {
      await post('/api/v1/auth/logout');
    } finally {
      // Always clear local storage even if API call fails
      localStorage.removeItem('admin_token');
      localStorage.removeItem('admin_user');
    }
  },

  /**
   * Check authentication status
   */
  checkAuth: async (): Promise<boolean> => {
    try {
      await get('/api/v1/auth/check');
      return true;
    } catch (error) {
      return false;
    }
  },

  /**
   * Get current user from localStorage
   */
  getCurrentUser: () => {
    const userStr = localStorage.getItem('admin_user');
    return userStr ? JSON.parse(userStr) : null;
  },

  /**
   * Get current token from localStorage
   */
  getToken: () => {
    return localStorage.getItem('admin_token');
  },

  /**
   * Check if user is authenticated (has valid token)
   */
  isAuthenticated: (): boolean => {
    return !!localStorage.getItem('admin_token');
  },
};

/**
 * Health check endpoint
 */
export const checkHealth = async (): Promise<HealthResponse> => {
  return get<HealthResponse>('/api/v1/health');
};
