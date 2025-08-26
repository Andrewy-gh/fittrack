import { client } from '@/client/client.gen';
import { stackClientApp } from '@/stack';
import { getAccessToken } from './auth';

const BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api';

client.setConfig({
  baseUrl: BASE_URL,
});

client.interceptors.request.use(async (request) => {
  const user = await stackClientApp.getUser();
  const accessToken = await getAccessToken(user);
  if (accessToken) {
    request.headers.set('x-stack-access-token', accessToken);
  }
  return request;
});