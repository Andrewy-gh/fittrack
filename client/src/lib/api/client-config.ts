import { client } from '@/client/client.gen';
import { stackClientApp } from '@/stack';

const BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api';

client.setConfig({
  baseUrl: BASE_URL,
});

client.interceptors.request.use(async (request) => {
  const user = await stackClientApp.getUser();
  if (user) {
    const { accessToken } = await user.getAuthJson();
    if (accessToken) {
      request.headers.set('x-stack-access-token', accessToken);
    }
  }
  return request;
});
