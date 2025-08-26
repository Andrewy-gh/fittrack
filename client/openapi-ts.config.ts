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
    {
      name: '@tanstack/react-query',
      case: 'camelCase',
      comments: true,
      exportFromIndex: false,
      mutationOptions: {
        // For your future mutations
        case: 'camelCase',
        enabled: true,
        name: '{{name}}Mutation',
      },
      queryKeys: {
        // Generate query keys for cache invalidation
        enabled: true,
        name: '{{name}}QueryKey',
        // Include tags if you want better cache invalidation
        tags: true, // set to true if you want operation tags in keys
      },
      queryOptions: {
        // Keep default camelCase
        case: 'camelCase',
        enabled: true,
        // Enable metadata if you want operation info
        meta: false, // or customize as needed
        // Use "Options" suffix to match your current naming
        name: '{{name}}Options',
      },
    },
  ],
});
 