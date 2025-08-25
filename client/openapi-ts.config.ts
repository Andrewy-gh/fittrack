import { defineConfig } from '@hey-api/openapi-ts';

export default defineConfig({
  input: '../server/docs/swagger.json',
  output: {
    format: 'prettier',
    lint: 'eslint',
    path: './src/client',
  },
  plugins: [
    '@hey-api/schemas', // Types for your Workout interface etc.
    {
      enums: 'javascript',
      name: '@hey-api/typescript',
    },
    {
      name: '@hey-api/sdk', // Replace your openapi-typescript-codegen service calls
      // transformer: false (default) since you don't need date transformation
    },
    '@tanstack/react-query', // The main benefit - auto-generated hooks!
  ],
});
 