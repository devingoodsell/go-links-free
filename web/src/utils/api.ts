import axios from 'axios';

export const api = axios.create({
  baseURL: 'http://localhost:8080',
});

// Add request interceptor to add token
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Add response interceptor to handle auth errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// Add a request interceptor to log requests in development
if (process.env.NODE_ENV === 'development') {
  api.interceptors.request.use(request => {
    console.log('API Request:', request.method, request.url);
    return request;
  });
} 